package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// ConsoleHandler formats log output as "·   message (key=value,...)" for info/debug
// and "W: message (...)" for warn, "E: message (...)" for error.
type ConsoleHandler struct {
	out   io.Writer
	level slog.Level
}

// NewConsoleHandler creates a new console log handler.
func NewConsoleHandler(out io.Writer, level slog.Level) *ConsoleHandler {
	return &ConsoleHandler{out: out, level: level}
}

func (h *ConsoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	var prefix string
	switch {
	case r.Level >= slog.LevelError:
		prefix = "E: "
	case r.Level >= slog.LevelWarn:
		prefix = "W: "
	default:
		prefix = "·   "
	}

	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
		return true
	})

	msg := r.Message
	if len(attrs) > 0 {
		msg = fmt.Sprintf("%s (%s)", msg, strings.Join(attrs, ","))
	}

	fmt.Fprintf(h.out, "%s%s\n", prefix, msg)
	return nil
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	return h
}

// FileHandler writes log output to a file with timestamp and level.
type FileHandler struct {
	out   io.Writer
	level slog.Level
}

// NewFileHandler creates a new file log handler.
func NewFileHandler(out io.Writer, level slog.Level) *FileHandler {
	return &FileHandler{out: out, level: level}
}

func (h *FileHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *FileHandler) Handle(_ context.Context, r slog.Record) error {
	var levelStr string
	switch {
	case r.Level >= slog.LevelError:
		levelStr = "ERROR"
	case r.Level >= slog.LevelWarn:
		levelStr = "WARN"
	case r.Level >= slog.LevelInfo:
		levelStr = "INFO"
	default:
		levelStr = "DEBUG"
	}

	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
		return true
	})

	msg := r.Message
	if len(attrs) > 0 {
		msg = fmt.Sprintf("%s (%s)", msg, strings.Join(attrs, ","))
	}

	timestamp := r.Time.Format("15:04:05")
	fmt.Fprintf(h.out, "%s %s %s\n", timestamp, levelStr, msg)
	return nil
}

func (h *FileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *FileHandler) WithGroup(name string) slog.Handler {
	return h
}

// MultiHandler writes to multiple handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a handler that writes to multiple handlers.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	return h
}
