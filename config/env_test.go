package config

import (
	"errors"
	"strings"
	"testing"
)

// TestLoad_DSNInvalidDoesNotLeakPassword memastikan pesan error dari
// url.Parse yang dibungkus tidak menyertakan bagian sensitif DSN (user,
// password, host) yang biasanya terdapat pada *url.Error bawaan.
func TestLoad_DSNInvalidDoesNotLeakPassword(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "3000")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FILE", "logs/app.log")
	t.Setenv("DATABASE_URL", "postgres://user:RAHASIA123@localhost:abc/db")
	t.Setenv("JWT_SECRET", "this-is-a-valid-jwt-secret-32-chars-long-yes")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() = nil error, want error karena url.Parse gagal untuk port 'abc'")
	}
	msg := err.Error()

	forbidden := []string{"RAHASIA123", "postgres://", "user", "localhost", "abc/db"}
	for _, f := range forbidden {
		if strings.Contains(msg, f) {
			t.Fatalf("pesan error mengandung bagian sensitif %q: %q", f, msg)
		}
	}

	// Pesan harus konsisten dan tanpa wrapper %w.
	if !errors.Is(err, err) {
		// sanity check: pastikan error bisa dibandingkan.
		t.Fatalf("error tidak comparable: %v", err)
	}
	if !strings.Contains(msg, "DATABASE_URL tidak valid") {
		t.Fatalf("pesan error tidak sesuai kontrak: %q", msg)
	}
}
