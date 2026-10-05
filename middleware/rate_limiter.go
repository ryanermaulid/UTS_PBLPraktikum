package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/ryanermaulid/UTS_PBLPraktikum/helper"
)

// LoginRateLimiterCfg menyimpan parameter konfigurasi rate limiter.
type LoginRateLimiterCfg struct {
	Max    int           // jumlah maksimum percobaan gagal dalam jendela
	Window time.Duration // default 1 menit
}

// LoginRateLimiter mengembalikan fiber.Handler yang membatasi
// percobaan gagal per IP. Implementasi memakai limiter bawaan Fiber
// dengan SkipSuccessfulRequests=true, sehingga hanya request yang
// berakhir 4xx yang dihitung.
//
// Header Retry-After dari limiter diteruskan apa adanya; middleware
// ini hanya menerjemahkan respons 429 menjadi helper.TooManyRequests
// agar ErrorHandler dapat memformat body sesuai standar.
func LoginRateLimiter(cfg LoginRateLimiterCfg) fiber.Handler {
	max := cfg.Max
	if max <= 0 {
		max = 5
	}
	win := cfg.Window
	if win <= 0 {
		win = time.Minute
	}

	return limiter.New(limiter.Config{
		Max:                    max,
		Expiration:             win,
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: true,
		LimiterMiddleware:      limiter.SlidingWindow{},
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return helper.TooManyRequests("Terlalu banyak percobaan login. Coba lagi nanti.")
		},
	})
}
