package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX adalah subset pgx yang dipakai oleh repository sehingga dapat
// dijalankan baik di luar maupun di dalam transaction.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// New membuka pgxpool dan memverifikasi koneksi dengan Ping. Mengembalikan
// error (bukan log.Fatal) agar pemanggil (main) dapat memutuskan alur.
// Pesan error dijaga agar tidak membocorkan kredensial DSN.
func New(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("konfigurasi pool database tidak valid: %w", sanitizeDSN(dsn, err))
	}
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool database: %w", sanitizeDSN(dsn, err))
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal menghubungi database: %w", sanitizeDSN(dsn, err))
	}
	return pool, nil
}

// sanitizeDSN mengganti setiap kemunculan DSN atau password yang terdapat
// dalam pesan error dengan placeholder, sehingga pesan akhir aman ditulis
// ke log atau stderr tanpa membocorkan kredensial.
func sanitizeDSN(dsn string, err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if dsn != "" {
		msg = strings.ReplaceAll(msg, dsn, "[DSN_REDACTED]")
	}
	if u, perr := url.Parse(dsn); perr == nil {
		if u.User != nil {
			if pwd, ok := u.User.Password(); ok && pwd != "" {
				msg = strings.ReplaceAll(msg, pwd, "[PASSWORD_REDACTED]")
			}
		}
		if u.RawQuery != "" {
			for k, vs := range u.Query() {
				lower := strings.ToLower(k)
				if strings.Contains(lower, "password") || strings.Contains(lower, "secret") {
					for _, v := range vs {
						msg = strings.ReplaceAll(msg, v, "[REDACTED]")
					}
				}
			}
		}
	}
	return fmt.Errorf("%s", msg)
}

// WithTx menjalankan fn dalam satu transaction. fn menerima DBTX yang
// terikat ke transaction tersebut. Rollback otomatis bila fn mengembalikan
// error atau terjadi panic; commit bila fn sukses.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(DBTX) error) (err error) {
	tx, beginErr := pool.Begin(ctx)
	if beginErr != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", beginErr)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				err = fmt.Errorf("%w (rollback: %v)", err, rbErr)
			}
			return
		}
		if cmErr := tx.Commit(ctx); cmErr != nil {
			err = fmt.Errorf("gagal commit transaksi: %w", cmErr)
		}
	}()

	return fn(tx)
}
