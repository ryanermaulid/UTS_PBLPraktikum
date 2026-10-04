package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// newRLApp membuat fiber.App dengan urutan middleware sesuai NewApp:
// requestid -> RequestLogger -> recover. Route yang ditambah hanya
// untuk pengujian. ErrorHandler memakai default Fiber agar status
// fiber.ErrNotFound (404) diteruskan apa adanya.
func newRLApp(logger *slog.Logger, withPanicRoute bool) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(requestid.New())
	app.Use(RequestLogger(logger))
	app.Use(recover.New(recover.Config{EnableStackTrace: false}))

	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	if withPanicRoute {
		app.Get("/panic", func(c *fiber.Ctx) error {
			panic("boom")
		})
	}
	return app
}

// findRequestLine mencari baris log JSON pertama yang memiliki field
// "msg" sama dengan "request" dan path tertentu. Mengembalikan level dan
// status, atau nilai default bila tidak ditemukan.
func findRequestLine(t *testing.T, raw string, path string) (string, int, bool) {
	t.Helper()
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		if m["msg"] != "request" {
			continue
		}
		if m["path"] != path {
			continue
		}
		lvl, _ := m["level"].(string)
		st := 0
		switch v := m["status"].(type) {
		case float64:
			st = int(v)
		}
		return lvl, st, true
	}
	return "", 0, false
}

func TestRequestLogger_UnknownRouteLogs404AsWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	app := newRLApp(logger, false)
	req := httptest.NewRequest("GET", "/api/v1/tidak-ada", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, resp.Body)

	lvl, status, ok := findRequestLine(t, buf.String(), "/api/v1/tidak-ada")
	if !ok {
		t.Fatalf("tidak ada baris log untuk unknown route; log:\n%s", buf.String())
	}
	if status != 404 {
		t.Fatalf("status log = %d, want 404", status)
	}
	if lvl != "WARN" {
		t.Fatalf("level log = %s, want WARN", lvl)
	}
}

func TestRequestLogger_PanicLogs500AsError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	app := newRLApp(logger, true)
	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, resp.Body)

	lvl, status, ok := findRequestLine(t, buf.String(), "/panic")
	if !ok {
		t.Fatalf("tidak ada baris log untuk /panic; log:\n%s", buf.String())
	}
	if status != 500 {
		t.Fatalf("status log = %d, want 500", status)
	}
	if lvl != "ERROR" {
		t.Fatalf("level log = %s, want ERROR", lvl)
	}
}

func TestRequestLogger_NormalRouteLogs200AsInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	app := newRLApp(logger, false)
	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, resp.Body)

	lvl, status, ok := findRequestLine(t, buf.String(), "/ok")
	if !ok {
		t.Fatalf("tidak ada baris log untuk /ok; log:\n%s", buf.String())
	}
	if status != 200 {
		t.Fatalf("status log = %d, want 200", status)
	}
	if lvl != "INFO" {
		t.Fatalf("level log = %s, want INFO", lvl)
	}
}
