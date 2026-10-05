package route

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ryanermaulid/UTS_PBLPraktikum/app/repository"
	"github.com/ryanermaulid/UTS_PBLPraktikum/app/service"
	mw "github.com/ryanermaulid/UTS_PBLPraktikum/middleware"
)

// Deps menyimpan semua dependensi yang dibutuhkan untuk mendaftarkan
// rute. Dipakai agar signature Register tidak panjang.
type Deps struct {
	DB         *pgxpool.Pool
	JWTSecret  string
	JWTExpires time.Duration
}

// time.Duration dipakai untuk konsistensi dengan limiter.Config.

// Register mendaftarkan seluruh endpoint SIAKAD Mini pada aplikasi
// Fiber. Grup root adalah /api/v1 sesuai SPEC.md. Endpoint auth
// dipasang lebih dulu karena dipakai oleh endpoint lain.
func Register(app *fiber.App, deps Deps) {
	if deps.DB == nil {
		panic("route.Register: DB nil")
	}

	users := repository.NewUserRepository(deps.DB)
	authSvc := service.NewAuthService(users, service.AuthServiceConfig{
		JWTSecret:         deps.JWTSecret,
		JWTExpiresMinutes: jwtExpiresMinutes(deps.JWTExpires),
	}, slog.Default())

	api := app.Group("/api/v1")

	authGroup := api.Group("/auth")
	// Rate limiter dipasang sebelum handler login. Handler auth.Login
	// tetap menggunakan helper.AppError agar ErrorHandler menjawab 422
	// untuk validasi dan 401 untuk kredensial salah.
	authGroup.Post("/login",
		mw.LoginRateLimiter(mw.LoginRateLimiterCfg{}),
		loginHandler(authSvc),
	)

	// /auth/me wajib token.
	authGroup.Get("/me",
		mw.RequireAuth(mw.AuthDeps{
			Users:     users,
			JWTSecret: deps.JWTSecret,
			Logger:    slog.Default(),
		}),
		meHandler(authSvc),
	)
}

// jwtExpiresMinutes mengubah time.Duration menjadi menit (int) sesuai
// yang diharapkan service.
func jwtExpiresMinutes(d time.Duration) int {
	if d <= 0 {
		return 60
	}
	return int(d.Minutes())
}
