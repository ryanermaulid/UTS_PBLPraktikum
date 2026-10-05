package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ryanermaulid/UTS_PBLPraktikum/app/model"
	"github.com/ryanermaulid/UTS_PBLPraktikum/app/repository"
	"github.com/ryanermaulid/UTS_PBLPraktikum/helper"
)

// AuthServiceConfig menyimpan parameter yang dibutuhkan untuk membuat
// token JWT (secret dan TTL). Disimpan sebagai struct terpisah agar
// service tidak bergantung pada objek Config lengkap.
type AuthServiceConfig struct {
	JWTSecret         string
	JWTExpiresMinutes int
}

// UserFinder adalah dependensi minimal AuthService terhadap repository.
// Interface ini dideklarasikan di sini (bukan di package repository)
// untuk menjaga arah ketergantungan dan agar mudah di-stub pada test.
type UserFinder interface {
	GetUserByEmail(ctx context.Context, email string) (*repository.UserRecord, error)
	GetUserByIDWithStudent(ctx context.Context, id int64) (*repository.UserRecord, error)
}

// AuthService adalah service untuk endpoint auth. Mengembalikan
// *helper.AppError untuk kesalahan yang aman ditampilkan, atau error
// generic yang akan dibungkus oleh service caller.
type AuthService struct {
	users UserFinder
	cfg   AuthServiceConfig
	log   *slog.Logger
}

func NewAuthService(users UserFinder, cfg AuthServiceConfig, log *slog.Logger) *AuthService {
	if log == nil {
		log = slog.Default()
	}
	return &AuthService{users: users, cfg: cfg, log: log}
}

// Login memvalidasi kredensial dan mengembalikan token JWT. Pesan
// kesalahan untuk email tidak ada, password salah, atau akun mahasiswa
// yang di-soft delete diseragamkan menjadi Unauthorized("Kredensial
// tidak valid") agar tidak membocorkan keberadaan akun.
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	rec, err := s.users.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, helper.Unauthorized("Kredensial tidak valid")
		}
		return nil, helper.Internal(err)
	}
	if !helper.CheckPassword(rec.Password, req.Password) {
		return nil, helper.Unauthorized("Kredensial tidak valid")
	}

	token, err := helper.IssueToken(s.cfg.JWTSecret, rec.ID, rec.Role, s.cfg.JWTExpiresMinutes)
	if err != nil {
		return nil, helper.Internal(err)
	}

	return &model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.cfg.JWTExpiresMinutes * 60,
		User: model.LoginUser{
			ID:    rec.ID,
			Email: rec.Email,
			Role:  rec.Role,
		},
	}, nil
}

// GetMe merakit UserMeResponse untuk user yang sedang login. Dipanggil
// oleh handler setelah RequireAuth memastikan token valid dan user
// belum di-soft delete (cek soft-delete dilakukan oleh repository).
func (s *AuthService) GetMe(ctx context.Context, userID int64) (*model.UserMeResponse, error) {
	rec, err := s.users.GetUserByIDWithStudent(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, helper.Unauthorized("Tidak terotorisasi")
		}
		return nil, helper.Internal(err)
	}
	return &model.UserMeResponse{
		ID:      rec.ID,
		Email:   rec.Email,
		Role:    rec.Role,
		Student: rec.Student,
	}, nil
}
