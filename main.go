package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryanermaulid/UTS_PBLPraktikum/config"
	"github.com/ryanermaulid/UTS_PBLPraktikum/database"
)

func main() {
	if err := run(); err != nil {
		// Pesan error di sini hanya berisi alasan; tidak memuat DATABASE_URL
		// atau secret.
		slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("startup gagal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger, closeLog, err := config.NewLogger(cfg)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)
	defer func() { _ = closeLog() }()

	startCtx, cancelStart := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStart()

	pool, err := database.New(startCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	logger.Info("koneksi database berhasil",
		slog.String("app_env", cfg.AppEnv),
		slog.Int("port", cfg.Port),
		slog.String("log_file", cfg.LogFile),
	)

	app := config.NewApp(cfg, logger)

	// Daftarkan route di Tahap 2/3. Untuk Tahap 1 belum ada endpoint
	// publik; server hanya menjalankan middleware global dan ErrorHandler.

	serverErr := make(chan error, 1)
	go func() {
		if err := app.Listen(":" + itoa(cfg.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case sig := <-stop:
		logger.Info("shutdown diterima", slog.String("signal", sig.String()))
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		return err
	}
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
