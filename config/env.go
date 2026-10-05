package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	Port              int
	DatabaseURL       string
	JWTSecret         string
	JWTExpiresMinutes int
	LogLevel          string
	LogFile           string
}

// Load membaca konfigurasi dari environment dan file .env (jika ada).
// Mengembalikan error bila DATABASE_URL kosong/tidak valid atau JWT_SECRET
// tidak memenuhi syarat. Pesan error tidak memuat nilai DATABASE_URL atau
// bagian sensitif lain.
func Load() (*Config, error) {
	// .env bersifat opsional. Tidak apa-apa bila tidak ada di production.
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		Port:              3000,
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:         strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTExpiresMinutes: 60,
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFile:           getEnv("LOG_FILE", "logs/app.log"),
	}

	if raw, ok := os.LookupEnv("PORT"); ok && raw != "" {
		p, err := strconv.Atoi(raw)
		if err != nil || p < 1 || p > 65535 {
			return nil, fmt.Errorf("PORT tidak valid: %s", raw)
		}
		cfg.Port = p
	}

	if raw, ok := os.LookupEnv("JWT_EXPIRES_MINUTES"); ok && raw != "" {
		m, err := strconv.Atoi(raw)
		if err != nil || m <= 0 {
			return nil, fmt.Errorf("JWT_EXPIRES_MINUTES tidak valid: %s", raw)
		}
		cfg.JWTExpiresMinutes = m
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL wajib diisi di environment atau .env")
	}
	if _, err := url.Parse(cfg.DatabaseURL); err != nil {
		// Pesan sengaja generik agar tidak membocorkan DSN yang biasanya
		// memuat kredensial. *url.Error bawaan menyertakan URL lengkap.
		return nil, errors.New("DATABASE_URL tidak valid")
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET wajib diisi")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET wajib diisi dengan panjang minimal 32 karakter")
	}

	if !isValidLogLevel(cfg.LogLevel) {
		return nil, fmt.Errorf("LOG_LEVEL tidak valid: %s (pilih debug/info/warn/error)", cfg.LogLevel)
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func isValidLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	}
	return false
}
