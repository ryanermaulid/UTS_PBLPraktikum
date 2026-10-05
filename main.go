package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ryanermaulid/UTS_PBLPraktikum/config"
	"github.com/ryanermaulid/UTS_PBLPraktikum/database"
	"github.com/ryanermaulid/UTS_PBLPraktikum/route"
)

const usage = `Penggunaan:
  go run .           menjalankan server
  go run . migrate   menjalankan migration
  go run . seed      menjalankan seeder`

func main() {
	code, err := dispatch(os.Args[1:])
	if err != nil {
		// Pesan error di sini hanya berisi alasan; tidak memuat DATABASE_URL
		// atau secret.
		fmt.Fprintln(os.Stderr, "galat:", err.Error())
		os.Exit(code)
	}
	os.Exit(code)
}

// dispatch mengembalikan exit code dan error (jika ada). Exit code
// mengikuti konvensi: 0 sukses, 1 kesalahan runtime, 2 argumen tak
// dikenal.
func dispatch(args []string) (int, error) {
	if len(args) == 0 {
		return runServer()
	}
	switch args[0] {
	case "migrate":
		return runMigrate()
	case "seed":
		return runSeed()
	default:
		fmt.Fprintln(os.Stderr, usage)
		return 2, fmt.Errorf("subcommand tidak dikenal: %s", args[0])
	}
}

func runServer() (int, error) {
	cfg, err := config.Load()
	if err != nil {
		return 1, err
	}

	logger, closeLog, err := config.NewLogger(cfg)
	if err != nil {
		return 1, err
	}
	slog.SetDefault(logger)
	defer func() { _ = closeLog() }()

	startCtx, cancelStart := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStart()

	pool, err := database.New(startCtx, cfg.DatabaseURL)
	if err != nil {
		return 1, err
	}
	defer pool.Close()

	logger.Info("koneksi database berhasil",
		slog.String("app_env", cfg.AppEnv),
		slog.Int("port", cfg.Port),
		slog.String("log_file", cfg.LogFile),
	)

	app := config.NewApp(cfg, logger)

	// Daftarkan seluruh endpoint aplikasi. Deps berisi pool, secret,
	// dan TTL token.
	route.Register(app, route.Deps{
		DB:         pool,
		JWTSecret:  cfg.JWTSecret,
		JWTExpires: time.Duration(cfg.JWTExpiresMinutes) * time.Minute,
	})

	serverErr := make(chan error, 1)
	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return 1, err
	case sig := <-stop:
		logger.Info("shutdown diterima", slog.String("signal", sig.String()))
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		return 1, err
	}
	return 0, nil
}

func runMigrate() (int, error) {
	pool, cleanup, err := bootstrap()
	if err != nil {
		return 1, err
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	res, err := database.Migrate(ctx, pool)
	if err != nil {
		return 1, err
	}
	if len(res.Applied) == 0 && len(res.Skipped) == 0 {
		fmt.Fprintln(os.Stderr, "Tidak ada file migration ditemukan.")
		return 0, nil
	}
	for _, n := range res.Applied {
		fmt.Fprintln(os.Stderr, "diterapkan:", n)
	}
	for _, n := range res.Skipped {
		fmt.Fprintln(os.Stderr, "dilewati:", n)
	}
	fmt.Fprintf(os.Stderr, "Selesai. Diterapkan=%d, dilewati=%d.\n", len(res.Applied), len(res.Skipped))
	return 0, nil
}

func runSeed() (int, error) {
	pool, cleanup, err := bootstrap()
	if err != nil {
		return 1, err
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	summary, err := database.Seed(ctx, pool)
	if err != nil {
		return 1, err
	}
	fmt.Fprintf(os.Stderr, "users_created=%d users_skipped=%d students_created=%d students_skipped=%d courses_created=%d courses_skipped=%d\n",
		summary.UsersCreated, summary.UsersSkipped,
		summary.StudentsCreated, summary.StudentsSkipped,
		summary.CoursesCreated, summary.CoursesSkipped,
	)
	return 0, nil
}

// bootstrap menyiapkan config, logger, dan pool untuk subcommand
// non-server. Pesan kesalahan dari config.Load dan database.New sudah
// dijaga agar tidak membocorkan nilai DATABASE_URL atau JWT_SECRET.
func bootstrap() (*pgxpool.Pool, func(), error) {
	cfg, lerr := config.Load()
	if lerr != nil {
		return nil, func() {}, lerr
	}
	lg, cl, lerr := config.NewLogger(cfg)
	if lerr != nil {
		return nil, func() {}, lerr
	}
	slog.SetDefault(lg)

	startCtx, startCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startCancel()

	p, derr := database.New(startCtx, cfg.DatabaseURL)
	if derr != nil {
		_ = cl()
		return nil, func() {}, derr
	}
	cleanup := func() {
		p.Close()
		_ = cl()
	}
	return p, cleanup, nil
}
