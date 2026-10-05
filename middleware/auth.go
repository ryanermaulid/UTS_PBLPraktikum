package middleware

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ryanermaulid/UTS_PBLPraktikum/app/repository"
	"github.com/ryanermaulid/UTS_PBLPraktikum/helper"
)

// userAuthFinder adalah dependensi minimal middleware auth. Interface
// ini didefinisikan lokal agar middleware tidak bergantung langsung
// pada package repository.
type userAuthFinder interface {
	GetUserByIDWithStudent(ctx context.Context, id int64) (*repository.UserRecord, error)
}

// RequireAuth memverifikasi header Authorization: Bearer <token>. Bila
// valid, menyimpan user lengkap di locals('auth_user') dan userID di
// locals('user_id'). Bila tidak valid, mengembalikan helper.Unauthorized
// agar ErrorHandler menjawab 401 dengan format standar.
func RequireAuth(deps AuthDeps) fiber.Handler {
	if deps.Users == nil {
		panic("RequireAuth: Users nil")
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	return func(c *fiber.Ctx) error {
		raw := bearerToken(c.Get("Authorization"))
		if raw == "" {
			return helper.Unauthorized("Token tidak ditemukan")
		}

		secret := deps.JWTSecret
		userID, _, err := helper.ParseToken(secret, raw)
		if err != nil {
			// Pesan generik; tidak membocorkan apakah format, signature,
			// atau klaim yang bermasalah.
			return helper.Unauthorized("Token tidak valid")
		}

		rec, err := deps.Users.GetUserByIDWithStudent(c.UserContext(), userID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return helper.Unauthorized("Token tidak valid")
			}
			return helper.Internal(err)
		}

		c.Locals("auth_user", rec)
		c.Locals("user_id", rec.ID)
		return c.Next()
	}
}

// RequireRole memastikan role user yang sedang login sesuai. Pasangkan
// setelah RequireAuth pada grup route.
func RequireRole(roles ...string) fiber.Handler {
	want := map[string]struct{}{}
	for _, r := range roles {
		want[r] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		rec, ok := c.Locals("auth_user").(*repository.UserRecord)
		if !ok || rec == nil {
			return helper.Unauthorized("Token tidak ditemukan")
		}
		if _, ok := want[rec.Role]; !ok {
			return helper.Forbidden("Akses ditolak untuk peran ini")
		}
		return c.Next()
	}
}

// bearerToken memotong header Authorization menjadi token. Mengembalikan
// string kosong bila header kosong atau tidak mengikuti format Bearer.
func bearerToken(h string) string {
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// AuthDeps adalah dependency yang dibutuhkan RequireAuth. Disimpan
// sebagai struct agar mudah ditambah bila ada kebutuhan baru.
type AuthDeps struct {
	Users     userAuthFinder
	JWTSecret string
	Logger    *slog.Logger
}
