package config

import (
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

	logger := slog.New(slog.NewMultiHandler(consoleHandler, fileHandler))
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
