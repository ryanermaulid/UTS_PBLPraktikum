package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/ryanermaulid/UTS_PBLPraktikum/helper"
	mw "github.com/ryanermaulid/UTS_PBLPraktikum/middleware"
)

// NewApp membuat *fiber.App dengan middleware global dan ErrorHandler
// terpusat. Urutan middleware: requestid -> RequestLogger -> recover.
// ErrorHandler hanya mengembalikan status dari daftar SPEC.md (200, 201,
// 204, 401, 403, 404, 409, 422, 429, 500). Status lain dipetakan ke:
//   - 404/405 -> 404 "Rute tidak ditemukan"
//   - 4xx lain -> 422
//   - 5xx atau error tak dikenal -> 500 generik
func NewApp(cfg *Config, logger *slog.Logger) *fiber.App {
	if logger == nil {
		logger = slog.Default()
	}

	app := fiber.New(fiber.Config{
		AppName:               "SIAKAD Mini",
		DisableStartupMessage: true,
		ErrorHandler:          newErrorHandler(logger),
	})

	app.Use(requestid.New())
	app.Use(mw.RequestLogger(logger))
	app.Use(recover.New(recover.Config{
		EnableStackTrace: false,
	}))

	return app
}

// errorBody adalah envelope response error. field "errors" hanya muncul
// pada body validasi (422 dengan per-field message).
type errorBody struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	Errors    map[string][]string `json:"errors,omitempty"`
	RequestID string              `json:"request_id"`
}

// newErrorHandler mengembalikan fiber.ErrorHandler yang:
//   - memetakan error ke salah satu status yang diizinkan SPEC.md,
//   - menghasilkan body JSON seragam dengan request_id,
//   - menulis log ERROR hanya untuk kasus yang berakhir 500. *AppError
//     dengan status < 500 dan *fiber.Error 4xx tidak di-log di sini
//     karena merupakan kesalahan客户端 yang wajar.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		rid, _ := c.Locals(requestid.ConfigDefault.ContextKey).(string)
		status, message, fields := classifyError(err)

		if status >= 500 {
			logger.Error("unhandled error",
				slog.String("request_id", rid),
				slog.String("method", c.Method()),
				slog.String("path", c.Path()),
				slog.Int("status", status),
				slog.String("error", err.Error()),
			)
		}

		return c.Status(status).JSON(errorBody{
			Success:   false,
			Message:   message,
			Errors:    fields,
			RequestID: rid,
		})
	}
}

// classifyError memetakan error menjadi (status, message, fields) yang
// sesuai dengan daftar status SPEC.md. fiber.Error 404/405 dipetakan ke
// 404 "Rute tidak ditemukan". 4xx lain dipetakan ke 422. 5xx dan error
// tak dikenal ke 500. Status 200/201/204 tidak pernah dihasilkan di sini.
func classifyError(err error) (int, string, map[string][]string) {
	if err == nil {
		return fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil
	}

	if ae, ok := helper.AsAppError(err); ok {
		return ae.Status, ae.Message, ae.Fields
	}

	var fe *fiber.Error
	if errors.As(err, &fe) {
		switch fe.Code {
		case fiber.StatusNotFound, fiber.StatusMethodNotAllowed:
			return fiber.StatusNotFound, "Rute tidak ditemukan", nil
		default:
			if fe.Code >= 400 && fe.Code < 500 {
				return fiber.StatusUnprocessableEntity, "Permintaan tidak dapat diproses", nil
			}
			return fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil
		}
	}

	return fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil
}
