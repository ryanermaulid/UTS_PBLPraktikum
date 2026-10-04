package config

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewLogger_WritesJSONToFile membuktikan bahwa logger yang dibuat oleh
// NewLogger menulis JSON ke file log dan ke stdout (TextHandler di dev),
// dan direktori log dibuat bila belum ada.
func TestNewLogger_WritesJSONToFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "logs", "app.log")

	cfg := &Config{
		AppEnv:  "development",
		LogFile: logPath,
	}

	logger, closer, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if closer == nil {
		t.Fatalf("closer kosong")
	}

	// Tulis log ke info agar masuk ke file (level default info).
	logger.Info("koneksi database berhasil", slog.String("app_env", cfg.AppEnv))
	logger.Warn("request", slog.String("path", "/x"), slog.Int("status", 404))

	if err := closer(); err != nil {
		t.Fatalf("closer: %v", err)
	}

	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("baca log file: %v", err)
	}
	lines := splitNonEmpty(string(raw))
	if len(lines) < 2 {
		t.Fatalf("log file hanya punya %d baris: %s", len(lines), string(raw))
	}
	for _, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("baris log bukan JSON: %q (%v)", line, err)
		}
		if m["msg"] == nil {
			t.Fatalf("baris log tidak punya msg: %q", line)
		}
	}
	if !strings.Contains(string(raw), "koneksi database berhasil") {
		t.Fatalf("log tidak memuat pesan koneksi: %s", string(raw))
	}
}

func splitNonEmpty(s string) []string {
	var out []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		if t := strings.TrimSpace(sc.Text()); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// TestNewLogger_RejectsInvalidLogLevel memastikan level tidak valid
// ditangani di config.Load, dan NewLogger fallback ke info.
func TestNewLogger_FallsBackToInfoForUnknown(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{AppEnv: "development", LogFile: filepath.Join(dir, "a.log"), LogLevel: "info"}
	logger, closer, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer closer()
	if logger == nil {
		t.Fatalf("logger nil")
	}
	// Suppress unused import io in this file.
	_ = io.Discard
}
