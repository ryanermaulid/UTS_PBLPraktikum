package config

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger membuat slog.Logger yang menulis ke konsol dan ke file dengan
// rotasi lumberjack. Pada APP_ENV=development, output konsol berupa teks
// untuk keterbacaan. Pada production, output konsol berupa JSON.
// File log selalu JSON agar mudah di-parse.
func NewLogger(cfg *Config) (*slog.Logger, func() error, error) {
	level := parseLevel(cfg.LogLevel)

	if dir := filepath.Dir(cfg.LogFile); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, err
		}
	}

	fileWriter := &lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // hari
		Compress:   true,
	}

	consoleOpts := &slog.HandlerOptions{Level: level}
	var consoleHandler slog.Handler
	if strings.EqualFold(cfg.AppEnv, "production") {
		consoleHandler = slog.NewJSONHandler(os.Stdout, consoleOpts)
	} else {
		consoleHandler = slog.NewTextHandler(os.Stdout, consoleOpts)
	}
	fileHandler := slog.NewJSONHandler(io.Writer(fileWriter), &slog.HandlerOptions{Level: level})

	logger := slog.New(&multiHandler{handlers: []slog.Handler{consoleHandler, fileHandler}})
	return logger, fileWriter.Close, nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
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

// multiHandler menggabungkan beberapa slog.Handler; setiap log diteruskan
// ke semuanya. Implementasi minimal untuk Tahap 1.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(_ context.Context, l slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(context.Background(), l) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range m.handlers {
		if err := h.Handle(ctx, r); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: next}
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: next}
}
