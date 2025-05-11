package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
)

type RichLoggerOptions struct {
	Output           io.Writer
	TimeFormat       string
	Level            slog.Level
	AddSource        bool
	EnableJSON       bool
	EnableColors     bool
	ShowLoggerName   bool
	TimestampInJSON  bool
	CompactJSON      bool
	EnableSeparators bool
}

func DefaultOptions() *RichLoggerOptions {
	return &RichLoggerOptions{
		Level:            slog.LevelInfo,
		AddSource:        false,
		EnableColors:     true,
		TimeFormat:       "2006-01-02 15:04:05.000",
		Output:           os.Stdout,
		ShowLoggerName:   true,
		TimestampInJSON:  true,
		CompactJSON:      false,
		EnableSeparators: true,
	}
}

type RichHandler struct {
	opts    *RichLoggerOptions
	mu      sync.Mutex
	attrs   []slog.Attr
	groups  []string
	loggers map[string]bool
}

func NewRichHandler(opts *RichLoggerOptions) *RichHandler {
	if opts == nil {
		opts = DefaultOptions()
	}
	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	return &RichHandler{
		opts:    opts,
		loggers: make(map[string]bool),
	}
}

func (h *RichHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level
}

func (h *RichHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := h.clone()
	h2.attrs = append(h2.attrs, attrs...)
	return h2
}

func (h *RichHandler) WithGroup(name string) slog.Handler {
	h2 := h.clone()
	h2.groups = append(h2.groups, name)
	return h2
}

func (h *RichHandler) clone() *RichHandler {
	h2 := &RichHandler{
		opts:    h.opts,
		attrs:   make([]slog.Attr, len(h.attrs)),
		groups:  make([]string, len(h.groups)),
		loggers: make(map[string]bool),
	}
	copy(h2.attrs, h.attrs)
	copy(h2.groups, h.groups)
	for k, v := range h.loggers {
		h2.loggers[k] = v
	}
	return h2
}

func (h *RichHandler) Handle(ctx context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.opts.EnableJSON {
		return h.handleJSON(ctx, record)
	}

	return h.handleText(ctx, record)
}

func (h *RichHandler) handleJSON(ctx context.Context, record slog.Record) error {
	jsonMap := make(map[string]interface{})

	// Add timestamp
	if h.opts.TimestampInJSON {
		jsonMap["time"] = record.Time.Format(h.opts.TimeFormat)
	}

	// Add level
	jsonMap["level"] = record.Level.String()

	// Add source if enabled
	if h.opts.AddSource && record.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{record.PC})
		f, _ := fs.Next()
		jsonMap["source"] = fmt.Sprintf("%s:%d", f.File, f.Line)
	}

	// Add message
	jsonMap["msg"] = record.Message

	// Add attributes
	addAttrs := func(a slog.Attr) bool {
		jsonMap[a.Key] = a.Value.Any()
		return true
	}

	for _, a := range h.attrs {
		addAttrs(a)
	}

	record.Attrs(func(a slog.Attr) bool {
		addAttrs(a)
		return true
	})

	var jsonData []byte
	var err error
	if h.opts.CompactJSON {
		jsonData, err = json.Marshal(jsonMap)
	} else {
		jsonData, err = json.MarshalIndent(jsonMap, "", "  ")
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(h.opts.Output, string(jsonData))
	return err
}

func (h *RichHandler) handleText(ctx context.Context, record slog.Record) error {
	var builder strings.Builder

	levelColors := map[slog.Level]string{
		slog.LevelDebug: Cyan,
		slog.LevelInfo:  Green,
		slog.LevelWarn:  Yellow,
		slog.LevelError: Red,
	}

	levelColor := levelColors[record.Level]
	if !h.opts.EnableColors {
		levelColor = ""
	}

	timeStr := record.Time.Format(h.opts.TimeFormat)
	if h.opts.EnableColors {
		builder.WriteString(Blue)
	}
	builder.WriteString(timeStr)
	builder.WriteString(" ")
	if h.opts.EnableColors {
		builder.WriteString(Reset)
	}

	levelStr := fmt.Sprintf("%-5s", strings.ToUpper(record.Level.String()))
	if h.opts.EnableColors {
		builder.WriteString(levelColor)
		builder.WriteString(Bold)
	}
	builder.WriteString(levelStr)
	if h.opts.EnableColors {
		builder.WriteString(Reset)
	}
	builder.WriteString(" ")

	if h.opts.AddSource && record.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{record.PC})
		f, _ := fs.Next()
		sourceFile := f.File
		if lastSlash := strings.LastIndex(sourceFile, "/"); lastSlash >= 0 {
			sourceFile = sourceFile[lastSlash+1:]
		}
		if h.opts.EnableColors {
			builder.WriteString(Magenta)
		}
		builder.WriteString(fmt.Sprintf("%s:%d", sourceFile, f.Line))
		if h.opts.EnableColors {
			builder.WriteString(Reset)
		}
		builder.WriteString(" ")
	}

	if h.opts.EnableColors {
		builder.WriteString(White)
		builder.WriteString(Bold)
	}
	builder.WriteString(record.Message)
	if h.opts.EnableColors {
		builder.WriteString(Reset)
	}

	if h.opts.EnableSeparators {
		builder.WriteString("\n")
		if h.opts.EnableColors {
			builder.WriteString(Blue)
		}
		builder.WriteString(strings.Repeat("─", 80))
		if h.opts.EnableColors {
			builder.WriteString(Reset)
		}
	}

	_, err := fmt.Fprintln(h.opts.Output, builder.String())
	return err
}

func NewRichLogger(opts *RichLoggerOptions) *slog.Logger {
	if opts == nil {
		opts = DefaultOptions()
	}
	handler := NewRichHandler(opts)
	return slog.New(handler)
}
