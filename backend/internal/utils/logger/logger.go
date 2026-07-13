package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level     string
	Format    string
	Module    string
	AddSource bool
}

func New(cfg Config) *slog.Logger {
	return NewWithWriter(cfg, os.Stdout)
}

func NewWithWriter(cfg Config, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Level),
		AddSource: cfg.AddSource,
	}

	var base slog.Handler
	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "text":
		base = slog.NewTextHandler(w, opts)
	default:
		base = slog.NewJSONHandler(w, opts)
	}

	h := &contextHandler{base: base}
	logger := slog.New(h)
	if cfg.Module != "" {
		logger = logger.With(slog.String("module", cfg.Module))
	}
	return logger
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
