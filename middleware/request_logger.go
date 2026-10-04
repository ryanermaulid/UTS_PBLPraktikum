package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// RequestLogger mencatat method, path, status final, dan latency tiap
// request. request_id dibaca setelah middleware requestid dipasang lebih
// dulu. Token, password, dan body tidak pernah dicatat.
//
// Jika handler mengembalikan error, ErrorHandler aplikasi dipanggil
// secara eksplisit di dalam middleware ini agar status akhir tercermin
// sebelum dicatat. Tujuannya: baris log mencerminkan status HTTP yang
// benar-benar diterima client, bukan 200 default dari Fiber.
//
// Level log mengikuti status final: 5xx = ERROR, 4xx = WARN, selain itu
// INFO. Field "error" tidak dicantumkan agar pesan internal tidak ikut
// tersimpan di log request.
func RequestLogger(logger *slog.Logger) fiber.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()
		if err != nil {
			if handlerErr := c.App().ErrorHandler(c, err); handlerErr != nil {
				_ = handlerErr
			}
		}

		status := c.Response().StatusCode()
		rid, _ := c.Locals(requestid.ConfigDefault.ContextKey).(string)
		dur := time.Since(start)

		attrs := []slog.Attr{
			slog.String("request_id", rid),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Int64("duration_ms", dur.Milliseconds()),
		}

		ctx := context.Background()
		switch {
		case status >= 500:
			logger.LogAttrs(ctx, slog.LevelError, "request", attrs...)
		case status >= 400:
			logger.LogAttrs(ctx, slog.LevelWarn, "request", attrs...)
		default:
			logger.LogAttrs(ctx, slog.LevelInfo, "request", attrs...)
		}

		return nil
	}
}
