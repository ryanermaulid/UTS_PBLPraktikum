package migrations

import "embed"

// FS mengekspos semua file *.sql di sebelah file ini. Dipakai oleh
// database.Migrate untuk menemukan skrip migration secara berurutan.
//
//go:embed *.sql
var FS embed.FS
