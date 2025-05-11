package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

type ProgressBar struct {
	startTime time.Time
	mu        sync.Mutex
	logger    *slog.Logger
	label     string
	total     int64
	current   int64
	width     int
	complete  bool
}

func NewProgressBar(total int64, label string, logger *slog.Logger) *ProgressBar {
	return &ProgressBar{
		total:     total,
		width:     50,
		label:     label,
		startTime: time.Now(),
		logger:    logger,
	}
}

func (p *ProgressBar) Increment(amount int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current += amount
	if p.current > p.total {
		p.current = p.total
	}

	p.render()
}

func (p *ProgressBar) Set(value int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = value
	if p.current > p.total {
		p.current = p.total
	}

	p.render()
}

func (p *ProgressBar) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.complete {
		return
	}

	p.current = p.total
	p.render()
	p.complete = true
	fmt.Fprintln(os.Stdout)
}

func (p *ProgressBar) render() {
	if p.complete {
		return
	}

	percent := float64(p.current) / float64(p.total) * 100
	filled := int(float64(p.width) * float64(p.current) / float64(p.total))

	elapsed := time.Since(p.startTime)
	var eta time.Duration
	if p.current > 0 {
		eta = time.Duration(float64(elapsed) * float64(p.total-p.current) / float64(p.current))
	}

	bar := fmt.Sprintf("\r%s [%s%s] %3.0f%% %d/%d ETA: %s ",
		p.label,
		strings.Repeat("█", filled),
		strings.Repeat("░", p.width-filled),
		percent,
		p.current,
		p.total,
		formatDuration(eta),
	)

	fmt.Fprint(os.Stdout, bar)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "0s"
	}

	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
