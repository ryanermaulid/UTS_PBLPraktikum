package config

import (
	"strings"
	"testing"
)

// setBaseEnv mengisi variabel lingkungan minimum yang dibutuhkan agar
// Load() tidak gagal di pengecekan urutan yang lebih awal (DATABASE_URL
// valid, JWT_SECRET >= 32 karakter).
func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "3000")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FILE", "logs/app.log")
	t.Setenv("DATABASE_URL", "postgres://app:app@localhost:5432/siakad?sslmode=disable")
	t.Setenv("JWT_SECRET", "this-is-a-valid-jwt-secret-of-32-or-more-chars")
	t.Setenv("JWT_EXPIRES_MINUTES", "60")
}

// TestLoad_DSNInvalidDoesNotLeakPassword memastikan pesan error dari
// url.Parse yang dibungkus tidak menyertakan bagian sensitif DSN (user,
// password, host) yang biasanya terdapat pada *url.Error bawaan.
func TestLoad_DSNInvalidDoesNotLeakPassword(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("DATABASE_URL", "postgres://user:RAHASIA123@localhost:abc/db")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() = nil error, want error karena url.Parse gagal untuk port 'abc'")
	}
	msg := err.Error()

	forbidden := []string{"RAHASIA123", "postgres://", "localhost"}
	for _, f := range forbidden {
		if strings.Contains(msg, f) {
			t.Fatalf("pesan error mengandung bagian sensitif %q: %q", f, msg)
		}
	}

	if !strings.Contains(msg, "DATABASE_URL tidak valid") {
		t.Fatalf("pesan error tidak sesuai kontrak: %q", msg)
	}
}

func TestLoad_InvalidPortRejected(t *testing.T) {
	cases := []string{"abc", "0", "70000", "-1"}
	for _, v := range cases {
		t.Run(v, func(t *testing.T) {
			setBaseEnv(t)
			t.Setenv("PORT", v)

			_, err := Load()
			if err == nil {
				t.Fatalf("PORT=%q: want error, got nil", v)
			}
			if !strings.Contains(err.Error(), "PORT tidak valid") {
				t.Fatalf("PORT=%q: want 'PORT tidak valid' in error, got %q", v, err.Error())
			}
			if !strings.Contains(err.Error(), v) {
				t.Fatalf("PORT=%q: want value echoed in error, got %q", v, err.Error())
			}
		})
	}
}

func TestLoad_DefaultPortApplied(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("PORT kosong: want default 3000, got error %v", err)
	}
	if cfg.Port != 3000 {
		t.Fatalf("cfg.Port = %d, want 3000", cfg.Port)
	}
}

func TestLoad_InvalidJWTExpiresRejected(t *testing.T) {
	cases := []string{"abc", "-5", "0"}
	for _, v := range cases {
		t.Run(v, func(t *testing.T) {
			setBaseEnv(t)
			t.Setenv("JWT_EXPIRES_MINUTES", v)

			_, err := Load()
			if err == nil {
				t.Fatalf("JWT_EXPIRES_MINUTES=%q: want error, got nil", v)
			}
			if !strings.Contains(err.Error(), "JWT_EXPIRES_MINUTES tidak valid") {
				t.Fatalf("JWT_EXPIRES_MINUTES=%q: want error mentioning 'JWT_EXPIRES_MINUTES tidak valid', got %q", v, err.Error())
			}
		})
	}
}

func TestLoad_DefaultJWTExpiresApplied(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("JWT_EXPIRES_MINUTES", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("JWT_EXPIRES_MINUTES kosong: want default 60, got error %v", err)
	}
	if cfg.JWTExpiresMinutes != 60 {
		t.Fatalf("cfg.JWTExpiresMinutes = %d, want 60", cfg.JWTExpiresMinutes)
	}
}

func TestLoad_JWTSecretWhitespaceOnlyRejected(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("JWT_SECRET", "   ")

	_, err := Load()
	if err == nil {
		t.Fatal("JWT_SECRET='   ': want error karena whitespace-only harus ditolak")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("want error menyebut JWT_SECRET, got %q", err.Error())
	}
}

// TestLoad_JWTSecretLengthOnlyFromWhitespace memastikan bahwa panjang
// minimum 32 hanya dihitung SETELAH TrimSpace. Secret 28 karakter ditambah
// spasi leading/trailing harus tetap ditolak.
func TestLoad_JWTSecretLengthOnlyFromWhitespace(t *testing.T) {
	setBaseEnv(t)
	// 28 karakter di tengah + spasi leading/trailing. Panjang total > 32,
	// tapi setelah TrimSpace tinggal 28 (kurang dari 32).
	const inner = "abcdefghijklmnopqrstuvwxyzab" // 28 karakter (a-z + "ab")
	if len(inner) != 28 {
		t.Fatalf("koreksi pengujian: inner = %d karakter, harus 28", len(inner))
	}
	t.Setenv("JWT_SECRET", "   "+inner+"   ")

	_, err := Load()
	if err == nil {
		t.Fatal("JWT_SECRET panjang setelah trim < 32: want error, got nil")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("want error menyebut JWT_SECRET, got %q", err.Error())
	}
}

func TestLoad_JWTSecretTrimmedAndAccepted(t *testing.T) {
	setBaseEnv(t)
	// 32 karakter di tengah + spasi. Setelah TrimSpace harus tetap
	// dianggap valid.
	const inner = "abcdefghijklmnopqrstuvwxyz123456" // 26 + 6 = 32
	if len(inner) != 32 {
		t.Fatalf("koreksi pengujian: inner = %d karakter, harus 32", len(inner))
	}
	secret := "  " + inner + "  "
	t.Setenv("JWT_SECRET", secret)
	if len(secret) <= 32 {
		t.Fatalf("sebelum trim = %d, harus > 32 agar pengujian berarti", len(secret))
	}
	trimmed := strings.TrimSpace(secret)
	if trimmed != inner {
		t.Fatalf("setelah trim = %q, harus %q", trimmed, inner)
	}
	if len(trimmed) != 32 {
		t.Fatalf("setelah trim = %d, harus 32", len(trimmed))
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("JWT_SECRET setelah trim = 32 karakter: got error %v", err)
	}
	if cfg.JWTSecret != trimmed {
		t.Fatalf("cfg.JWTSecret = %q, want %q", cfg.JWTSecret, trimmed)
	}
}
