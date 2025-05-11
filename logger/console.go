package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Console struct {
	Logger    *slog.Logger
	ShowTime  bool
	Colorized bool
}

func NewConsole(opts *RichLoggerOptions) *Console {
	if opts == nil {
		opts = DefaultOptions()
	}

	return &Console{
		Logger:    NewRichLogger(opts),
		ShowTime:  true,
		Colorized: opts.EnableColors,
	}
}

func (c *Console) StartTimer(name string) *Timer {
	return &Timer{
		Name:      name,
		StartTime: time.Now(),
		Console:   c,
	}
}

func (c *Console) Success(format string, args ...interface{}) {
	msg := "✓ " + fmt.Sprintf(format, args...)
	if c.Colorized {
		msg = Green + Bold + msg + Reset
	}
	c.Logger.Info(msg)
}

func (c *Console) Info(format string, args ...interface{}) {
	msg := "ℹ " + fmt.Sprintf(format, args...)
	if c.Colorized {
		msg = Blue + Bold + msg + Reset
	}
	c.Logger.Info(msg)
}

func (c *Console) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if c.Colorized {
		msg = White + msg + Reset
	}
	c.Logger.Info(msg)
}

func (c *Console) Warn(format string, args ...interface{}) {
	msg := "⚠ " + fmt.Sprintf(format, args...)
	if c.Colorized {
		msg = Yellow + Bold + msg + Reset
	}
	c.Logger.Warn(msg)
}

func (c *Console) Error(format string, args ...interface{}) {
	msg := "✖ " + fmt.Sprintf(format, args...)
	if c.Colorized {
		msg = Red + Bold + msg + Reset
	}
	c.Logger.Error(msg)
}

func (c *Console) Fatal(format string, args ...interface{}) {
	msg := "💀 " + fmt.Sprintf(format, args...)
	if c.Colorized {
		msg = BgRed + White + Bold + msg + Reset
	}
	c.Logger.Error(msg)
	os.Exit(1)
}

func (c *Console) StartSpinner(message string) *Spinner {
	s := &Spinner{
		Message: message,
		Frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		Console: c,
		Done:    make(chan bool),
	}

	s.Start()
	return s
}

func (c *Console) NewProgressBar(total int64, label string) *ProgressBar {
	return NewProgressBar(total, label, c.Logger)
}

func (c *Console) NewTable(headers []string) *Table {
	return NewTable(headers, c.Logger)
}

func (c *Console) Box(title string, content string) {
	lines := splitLines(content)
	maxWidth := len(title)

	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	maxWidth += 4

	fmt.Println("┌" + "─" + title + "─" + strings.Repeat("─", maxWidth-len(title)-2) + "┐")

	for _, line := range lines {
		fmt.Println("│ " + line + strings.Repeat(" ", maxWidth-len(line)) + " │")
	}

	fmt.Println("└" + strings.Repeat("─", maxWidth+2) + "┘")
}

func splitLines(text string) []string {
	return strings.Split(text, "\n")
}
