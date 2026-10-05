package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/ryanermaulid/UTS_PBLPraktikum/app/model"
	"github.com/ryanermaulid/UTS_PBLPraktikum/app/repository"
	"github.com/ryanermaulid/UTS_PBLPraktikum/helper"
)

// stubUserRepo mengimplementasikan service.UserFinder untuk test.
// Tanpa koneksi database; cukup return nilai yang sudah disiapkan.
type stubUserRepo struct {
	byEmail *repository.UserRecord
	byID    *repository.UserRecord
	errByE  error
	errByID error
}

func (s *stubUserRepo) GetUserByEmail(_ context.Context, email string) (*repository.UserRecord, error) {
	if s.errByE != nil {
		return nil, s.errByE
	}
	if s.byEmail == nil {
		return nil, repository.ErrUserNotFound
	}
	if email != "" && email != s.byEmail.Email {
		return nil, repository.ErrUserNotFound
	}
	return s.byEmail, nil
}

func (s *stubUserRepo) GetUserByIDWithStudent(_ context.Context, id int64) (*repository.UserRecord, error) {
	if s.errByID != nil {
		return nil, s.errByID
	}
	if s.byID == nil {
		return nil, repository.ErrUserNotFound
	}
	if s.byID.ID != id {
		return nil, repository.ErrUserNotFound
	}
	return s.byID, nil
}

const testJWTSecret = "secret-yang-panjang-cukup-32-karakter-!"

func newTestAuthService(repo UserFinder) *AuthService {
	return NewAuthService(repo, AuthServiceConfig{
		JWTSecret:         testJWTSecret,
		JWTExpiresMinutes: 60,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func hashedPassword(t *testing.T, plain string) string {
	t.Helper()
	h, err := helper.HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword gagal: %v", err)
	}
	return h
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := &stubUserRepo{
		byEmail: &repository.UserRecord{
			ID: 7, Email: "a@b.com", Password: hashedPassword(t, "Admin12345!"), Role: "admin",
		},
	}
	svc := newTestAuthService(repo)

	resp, err := svc.Login(context.Background(), model.LoginRequest{
		Email: "a@b.com", Password: "Admin12345!",
	})
	if err != nil {
		t.Fatalf("Login gagal: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("token kosong")
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want Bearer", resp.TokenType)
	}
	if resp.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", resp.ExpiresIn)
	}
	if resp.User.ID != 7 || resp.User.Role != "admin" || resp.User.Email != "a@b.com" {
		t.Errorf("User tidak sesuai: %+v", resp.User)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := &stubUserRepo{
		byEmail: &repository.UserRecord{
			ID: 1, Email: "a@b.com", Password: hashedPassword(t, "rightpassword1"), Role: "admin",
		},
	}
	svc := newTestAuthService(repo)

	_, err := svc.Login(context.Background(), model.LoginRequest{
		Email: "a@b.com", Password: "salahpassword",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	ae, ok := helper.AsAppError(err)
	if !ok || ae.Status != 401 {
		t.Fatalf("expected 401, got %v", err)
	}
}

func TestAuthService_Login_EmailNotFound(t *testing.T) {
	repo := &stubUserRepo{}
	svc := newTestAuthService(repo)

	_, err := svc.Login(context.Background(), model.LoginRequest{
		Email: "ghost@nope.test", Password: "anything12345",
	})
	ae, ok := helper.AsAppError(err)
	if !ok || ae.Status != 401 {
		t.Fatalf("expected 401 untuk email tak ditemukan, got %v", err)
	}
}

func TestAuthService_Login_SoftDeletedStudentDenied(t *testing.T) {
	// GetUserByEmail -> repository.ErrUserNotFound karena mahasiswa
	// soft-delete. Service harus menjawab 401 tanpa membedakan.
	repo := &stubUserRepo{}
	svc := newTestAuthService(repo)

	_, err := svc.Login(context.Background(), model.LoginRequest{
		Email: "deleted@student.test", Password: "anything12345",
	})
	ae, ok := helper.AsAppError(err)
	if !ok || ae.Status != 401 {
		t.Fatalf("expected 401 untuk mahasiswa soft-delete, got %v", err)
	}
}

func TestAuthService_GetMe_Admin(t *testing.T) {
	repo := &stubUserRepo{
		byID: &repository.UserRecord{ID: 1, Email: "admin@siakad.test", Role: "admin"},
	}
	svc := newTestAuthService(repo)

	resp, err := svc.GetMe(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetMe gagal: %v", err)
	}
	if resp.Student != nil {
		t.Errorf("admin tidak boleh punya student, got %+v", resp.Student)
	}
	if resp.Role != "admin" || resp.Email != "admin@siakad.test" {
		t.Errorf("field user salah: %+v", resp)
	}
}

func TestAuthService_GetMe_Student(t *testing.T) {
	repo := &stubUserRepo{
		byID: &repository.UserRecord{
			ID: 2, Email: "187221000001@student.siakad.test", Role: "mahasiswa",
			Student: &model.StudentMeResponse{
				ID: 99, NIM: "187221000001", Nama: "Andi",
				Prodi: "Sistem Informasi", Angkatan: 2023,
			},
		},
	}
	svc := newTestAuthService(repo)

	resp, err := svc.GetMe(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetMe gagal: %v", err)
	}
	if resp.Student == nil {
		t.Fatal("Student kosong untuk role mahasiswa")
	}
	if resp.Student.ID != 99 || resp.Student.NIM != "187221000001" {
		t.Errorf("Student data salah: %+v", resp.Student)
	}
}

func TestAuthService_GetMe_InternalErrorOnRepoFailure(t *testing.T) {
	repo := &stubUserRepo{errByID: errors.New("koneksi putus")}
	svc := newTestAuthService(repo)

	_, err := svc.GetMe(context.Background(), 1)
	ae, ok := helper.AsAppError(err)
	if !ok || ae.Status != 500 {
		t.Fatalf("expected 500 untuk kesalahan internal, got %v", err)
	}
}
