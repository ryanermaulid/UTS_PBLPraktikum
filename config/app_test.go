package config

import (
	"bufio"
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

	"github.com/ryanermaulid/UTS_PBLPraktikum/helper"
)

// newTestApp membuat fiber.App khusus pengujian dengan middleware yang
// sama dengan NewApp, lalu menambahkan route sementara untuk menguji
// ErrorHandler. Tidak menambah route ke aplikasi sebenarnya.
func newTestApp(logger *slog.Logger) *fiber.App {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          newErrorHandler(logger),
	})
	app.Use(requestid.New())
	app.Use(recover.New(recover.Config{EnableStackTrace: false}))

	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("boom")
	})
	app.Get("/apperror", func(c *fiber.Ctx) error {
		return helper.NotFound("Data tidak ditemukan")
	})
	app.Get("/generic", func(c *fiber.Ctx) error {
		// Pesan ini seharusnya tidak pernah muncul di body response.
		return fiber.NewError(fiber.StatusInternalServerError, "rahasia internal: db down")
	})
	app.Get("/fiber400", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "bad")
	})
	app.Get("/fiber404", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "missing")
	})
	app.Get("/fiber500", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "bad")
	})

	return app
}

type testResp struct {
	Status int
	Body   []byte
	JSON   map[string]any
}

func doRequest(t *testing.T, app *fiber.App, method, path string) testResp {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test(%s %s): %v", method, path, err)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("baca body: %v", err)
	}
	out := testResp{Status: resp.StatusCode, Body: raw}
	if len(raw) > 0 {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err == nil {
			out.JSON = m
		}
	}
	return out
}

func TestErrorHandler_UnknownRoute_ReturnsNotFoundJSON(t *testing.T) {
	app := newTestApp(nil)
	r := doRequest(t, app, "GET", "/api/v1/tidak-ada")

	if r.Status != fiber.StatusNotFound {
		t.Fatalf("status = %d, want 404", r.Status)
	}
	if r.JSON["success"] != false {
		t.Fatalf("success = %v, want false", r.JSON["success"])
	}
	if r.JSON["message"] != "Rute tidak ditemukan" {
		t.Fatalf("message = %v, want 'Rute tidak ditemukan'", r.JSON["message"])
	}
	rid, _ := r.JSON["request_id"].(string)
	if rid == "" {
		t.Fatalf("request_id kosong: %v", r.JSON)
	}
}

func TestErrorHandler_PanicReturns500GenericNoStackTrace(t *testing.T) {
	app := newTestApp(nil)
	r := doRequest(t, app, "GET", "/panic")

	if r.Status != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", r.Status)
	}
	if r.JSON["success"] != false {
		t.Fatalf("success = %v, want false", r.JSON["success"])
	}
	if msg, _ := r.JSON["message"].(string); msg != "Terjadi kesalahan pada server" {
		t.Fatalf("message = %v, want pesan generik 500", r.JSON["message"])
	}
	combined := strings.ToLower(string(r.Body))
	for _, leak := range []string{"goroutine", "boom", "stack trace", "recovered"} {
		if strings.Contains(combined, leak) {
			t.Fatalf("response bocor detail internal (%s): %s", leak, string(r.Body))
		}
	}
}

func TestErrorHandler_AppErrorPassesThrough(t *testing.T) {
	app := newTestApp(nil)
	r := doRequest(t, app, "GET", "/apperror")

	if r.Status != fiber.StatusNotFound {
		t.Fatalf("status = %d, want 404", r.Status)
	}
	if msg, _ := r.JSON["message"].(string); msg != "Data tidak ditemukan" {
		t.Fatalf("message = %v, want 'Data tidak ditemukan'", r.JSON["message"])
	}
	if rid, _ := r.JSON["request_id"].(string); rid == "" {
		t.Fatalf("request_id kosong: %v", r.JSON)
	}
}

func TestErrorHandler_GenericErrorReturns500NoLeak(t *testing.T) {
	app := newTestApp(nil)
	r := doRequest(t, app, "GET", "/generic")

	if r.Status != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", r.Status)
	}
	if msg, _ := r.JSON["message"].(string); msg != "Terjadi kesalahan pada server" {
		t.Fatalf("message = %v, want pesan generik 500", r.JSON["message"])
	}
	if strings.Contains(string(r.Body), "rahasia internal") {
		t.Fatalf("response bocor pesan error internal: %s", string(r.Body))
	}
}

func TestErrorHandler_FiberNotFoundMappedToRouteFormat(t *testing.T) {
	app := newTestApp(nil)
	r := doRequest(t, app, "GET", "/fiber404")

	if r.Status != fiber.StatusNotFound {
		t.Fatalf("status = %d, want 404", r.Status)
	}
	if msg, _ := r.JSON["message"].(string); msg != "Rute tidak ditemukan" {
		t.Fatalf("message = %v, want 'Rute tidak ditemukan'", r.JSON["message"])
	}
}

func TestErrorHandler_Fiber400MappedTo422(t *testing.T) {
	app := newTestApp(nil)
	r := doRequest(t, app, "GET", "/fiber400")

	if r.Status != fiber.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", r.Status)
	}
}

func TestErrorHandler_Fiber500MappedTo500(t *testing.T) {
	app := newTestApp(nil)
	r := doRequest(t, app, "GET", "/fiber500")

	if r.Status != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", r.Status)
	}
	if msg, _ := r.JSON["message"].(string); msg != "Terjadi kesalahan pada server" {
		t.Fatalf("message = %v, want pesan generik 500", r.JSON["message"])
	}
}

// TestErrorHandler_UnknownRouteDoesNotLogError memastikan 404 untuk rute
// tak dikenal tidak memicu log ERROR. Logger diarahkan ke buffer JSON
// sehingga setiap baris log dapat diinspeksi.
func TestErrorHandler_UnknownRouteDoesNotLogError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	app := newTestApp(logger)
	r := doRequest(t, app, "GET", "/api/v1/tidak-ada")
	if r.Status != fiber.StatusNotFound {
		t.Fatalf("status = %d, want 404", r.Status)
	}
	if r.JSON["success"] != false {
		t.Fatalf("success = %v, want false", r.JSON["success"])
	}

	sc := bufio.NewScanner(strings.NewReader(buf.String()))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("baris log bukan JSON: %q (%v)", line, err)
		}
		if lvl, _ := m["level"].(string); lvl == "ERROR" {
			t.Fatalf("ErrorHandler menulis log ERROR untuk status 4xx: %s", line)
		}
	}
}

// TestErrorHandler_PanicDoesLogError memastikan 500 dari panic memang
// memicu log ERROR; pasangan dari test di atas agar cakupan tetap jelas.
func TestErrorHandler_PanicDoesLogError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	app := newTestApp(logger)
	r := doRequest(t, app, "GET", "/panic")
	if r.Status != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", r.Status)
	}

	sc := bufio.NewScanner(strings.NewReader(buf.String()))
	foundErrLevel := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		if lvl, _ := m["level"].(string); lvl == "ERROR" {
			foundErrLevel = true
		}
	}
	if !foundErrLevel {
		t.Fatalf("ErrorHandler tidak menulis log ERROR untuk 500; log:\n%s", buf.String())
	}
}
