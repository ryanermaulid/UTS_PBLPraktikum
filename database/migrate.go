package database

import (
	"context"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ryanermaulid/UTS_PBLPraktikum/migrations"
)

// MigrationResult merangkum hasil satu panggilan Migrate.
type MigrationResult struct {
	Applied []string
	Skipped []string
}

// Migrate menjalankan seluruh migration di package migrations secara
// berurutan, dalam satu transaction per file. Migration yang sudah
// tercatat di schema_migrations akan dilewati.
//
// Aman dipanggil berulang kali: tabel schema_migrations menjadi catatan
// versi dan setiap file hanya dieksekusi satu kali.
//
// Pesan error dijaga agar tidak membocorkan nilai rahasia (DSN atau
// password); jika pgx mengembalikan error yang mengandung informasi
// tersebut, sanitizeDSN akan menyamarkannya sebelum dibungkus.
func Migrate(ctx context.Context, pool *pgxpool.Pool) (MigrationResult, error) {
	if pool == nil {
		return MigrationResult{}, fmt.Errorf("migrate: pool tidak tersedia")
	}

	res := MigrationResult{}

	if err := WithTx(ctx, pool, func(dbtx DBTX) error {
		if _, err := dbtx.Exec(ctx, `
            CREATE TABLE IF NOT EXISTS schema_migrations (
                version    TEXT        PRIMARY KEY,
                applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
            )`); err != nil {
			return fmt.Errorf("migrate: gagal membuat tabel schema_migrations: %w", err)
		}

		entries, err := fs.ReadDir(migrations.FS, ".")
		if err != nil {
			return fmt.Errorf("migrate: gagal membaca direktori migration: %w", err)
		}

		var names []string
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			n := e.Name()
			if len(n) >= 4 && n[len(n)-4:] == ".sql" {
				names = append(names, n)
			}
		}
		sort.Strings(names)

		for _, name := range names {
			var existing string
			err := dbtx.QueryRow(ctx,
				`SELECT version FROM schema_migrations WHERE version = $1`, name,
			).Scan(&existing)
			if err == nil {
				res.Skipped = append(res.Skipped, name)
				continue
			}
			if err != pgx.ErrNoRows {
				return fmt.Errorf("migrate: gagal membaca schema_migrations: %w", err)
			}

			script, err := migrations.FS.ReadFile(name)
			if err != nil {
				return fmt.Errorf("migrate: gagal membaca %s: %w", name, err)
			}

			if _, err := dbtx.Exec(ctx, string(script)); err != nil {
				return fmt.Errorf("migrate: gagal menjalankan %s: %w", name, err)
			}

			if _, err := dbtx.Exec(ctx,
				`INSERT INTO schema_migrations (version) VALUES ($1)`, name,
			); err != nil {
				return fmt.Errorf("migrate: gagal mencatat %s: %w", name, err)
			}

			res.Applied = append(res.Applied, name)
		}
		return nil
	}); err != nil {
		return res, err
	}

	return res, nil
}
