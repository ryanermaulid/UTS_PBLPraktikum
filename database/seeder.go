package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ryanermaulid/UTS_PBLPraktikum/helper"
)

// AdminEmail adalah email admin yang dipakai oleh seeder dan harus
// dicatat di README.
const AdminEmail = "admin@siakad.test"

// AdminPassword adalah password admin. Minimal 8 karakter.
const AdminPassword = "Admin12345!"

// StudentSeed adalah representasi murni (tanpa akses DB) dari satu
// baris mahasiswa untuk diuji.
type StudentSeed struct {
	NIM       string
	Nama      string
	Email     string
	Prodi     string
	Angkatan  int
	IPK       *float64
	Plaintext string // password plain untuk di-hash; default = NIM
}

// AdminSeed adalah representasi murni admin.
type AdminSeed struct {
	Email     string
	Plaintext string
}

// CourseSeed adalah representasi murni satu mata kuliah.
type CourseSeed struct {
	KodeMK   string
	NamaMK   string
	SKS      int
	Semester int
	Kuota    int
}

// ProdiValues adalah daftar prodi yang dipakai seeder dan dipakai pula
// untuk validasi unit test.
var ProdiValues = []string{
	"Sistem Informasi",
	"Teknik Informatika",
	"Teknik Komputer",
}

// AdminSeedData mengekspos data admin untuk unit test.
func AdminSeedData() AdminSeed { return adminSeedData() }

// StudentSeedData mengekspos data mahasiswa untuk unit test.
func StudentSeedData() []StudentSeed { return studentSeedData() }

// CourseSeedData mengekspos data mata kuliah untuk unit test.
func CourseSeedData() []CourseSeed { return courseSeedData() }

// SeedSummary merangkum jumlah baris yang dibuat dan dilewati oleh
// seeder untuk dicetak oleh subcommand.
type SeedSummary struct {
	UsersCreated    int
	UsersSkipped    int
	StudentsCreated int
	StudentsSkipped int
	CoursesCreated  int
	CoursesSkipped  int
}

// adminSeedData menghasilkan data admin deterministik.
func adminSeedData() AdminSeed {
	return AdminSeed{Email: AdminEmail, Plaintext: AdminPassword}
}

// studentSeedData menghasilkan 20 mahasiswa deterministik. NIM urut
// dari 187221000001 sampai 187221000020. IPK mencakup kasus uji
// yang dibutuhkan SPEC (3.00, 3.45, 2.99, 2.50, 2.49, 1.80, 4.00, NULL)
// dan variasi lain agar total 20.
func studentSeedData() []StudentSeed {
	nama := []string{
		"Ahmad Pratama", "Budi Santoso", "Citra Lestari", "Dewi Anggraini",
		"Eka Putri", "Fajar Nugroho", "Gita Permata", "Hendra Wijaya",
		"Indah Sari", "Joko Susilo",
		"Kartika Rahayu", "Lutfi Hakim", "Maya Safitri", "Nanda Pradana",
		"Olivia Tan", "Putu Adi", "Salwa Rahmadani", "Tio Saputra",
		"Umi Kalsum", "Vina Octavia",
	}

	// Kasus uji pada urutan NIM awal: 3.00, 3.45, 2.99, 2.50, 2.49,
	// 1.80, 4.00, NULL. Sisanya variasi acuan deterministik.
	ipkCases := []struct {
		set   bool
		value float64
	}{
		{true, 3.00}, {true, 3.45}, {true, 2.99}, {true, 2.50},
		{true, 2.49}, {true, 1.80}, {true, 4.00}, {false, 0},
	}
	variasi := []float64{
		3.10, 3.25, 2.80, 3.55, 2.65, 3.75, 3.20, 2.95, 3.40, 2.70, 3.85, 3.05,
	}

	const angkatanAwal = 2022
	seeds := make([]StudentSeed, 0, 20)
	for i := 0; i < 20; i++ {
		nim := fmt.Sprintf("187221%06d", i+1)
		prodi := ProdiValues[i%len(ProdiValues)]
		// 4 angkatan (2022-2025) dipakai bergantian: angkatan naik tiap
		// 5 mahasiswa sehingga tiap angkatan mendapat 5 mahasiswa.
		angkatan := angkatanAwal + i/5
		if angkatan > 2026 {
			angkatan = 2026
		}

		var ipk *float64
		switch {
		case i < len(ipkCases):
			if ipkCases[i].set {
				v := ipkCases[i].value
				ipk = &v
			} else {
				ipk = nil
			}
		default:
			v := variasi[i-len(ipkCases)]
			ipk = &v
		}

		seeds = append(seeds, StudentSeed{
			NIM:       nim,
			Nama:      nama[i],
			Email:     nim + "@student.siakad.test",
			Prodi:     prodi,
			Angkatan:  angkatan,
			IPK:       ipk,
			Plaintext: nim,
		})
	}
	return seeds
}

// courseSeedData menghasilkan 10 mata kuliah. Memasukkan kuota=1 dan
// kuota=2 untuk menguji jalur kuota penuh. SKS didominasi 4 agar batas
// 18 SKS mudah dicapai.
func courseSeedData() []CourseSeed {
	return []CourseSeed{
		{KodeMK: "IF101", NamaMK: "Algoritma dan Pemrograman", SKS: 3, Semester: 1, Kuota: 40},
		{KodeMK: "IF102", NamaMK: "Matematika Diskrit", SKS: 4, Semester: 1, Kuota: 1},
		{KodeMK: "IF201", NamaMK: "Struktur Data", SKS: 3, Semester: 2, Kuota: 35},
		{KodeMK: "IF202", NamaMK: "Basis Data", SKS: 4, Semester: 2, Kuota: 2},
		{KodeMK: "IF301", NamaMK: "Sistem Operasi", SKS: 3, Semester: 3, Kuota: 30},
		{KodeMK: "IF302", NamaMK: "Jaringan Komputer", SKS: 4, Semester: 3, Kuota: 25},
		{KodeMK: "IF401", NamaMK: "Rekayasa Perangkat Lunak", SKS: 2, Semester: 4, Kuota: 50},
		{KodeMK: "IF402", NamaMK: "Kecerdasan Buatan", SKS: 4, Semester: 4, Kuota: 30},
		{KodeMK: "IF501", NamaMK: "Pemrograman Web", SKS: 3, Semester: 5, Kuota: 45},
		{KodeMK: "IF502", NamaMK: "Proyek Perangkat Lunak", SKS: 4, Semester: 6, Kuota: 60},
	}
}

// Seed menulis data awal ke database dalam satu transaction. INSERT
// menggunakan ON CONFLICT DO NOTHING sehingga pemanggilan berulang tidak
// menambah baris (idempotent pada kolom unik).
//
// Fungsi ini mengembalikan SeedSummary yang merangkum jumlah baris
// yang dibuat dan dilewati. Pesan error dibungkus tanpa membocorkan
// DSN atau password.
func Seed(ctx context.Context, pool *pgxpool.Pool) (SeedSummary, error) {
	summary := SeedSummary{}
	if pool == nil {
		return summary, fmt.Errorf("seed: pool tidak tersedia")
	}

	admin := adminSeedData()
	students := studentSeedData()
	courses := courseSeedData()

	// Hash ditahan sampai dalam transaksi. Bila gagal sebelum INSERT,
	// kita tidak menulis apa-apa ke database.
	adminHash, err := helper.HashPassword(admin.Plaintext)
	if err != nil {
		return summary, fmt.Errorf("seed: gagal hash password admin: %w", err)
	}
	studentHashes := make([]string, len(students))
	for i, s := range students {
		h, err := helper.HashPassword(s.Plaintext)
		if err != nil {
			return summary, fmt.Errorf("seed: gagal hash password untuk %s: %w", s.NIM, err)
		}
		studentHashes[i] = h
	}

	err = WithTx(ctx, pool, func(dbtx DBTX) error {
		// Admin: cek dulu apakah sudah ada. ON CONFLICT (email) tetap
		// dipakai agar pemanggilan bersamaan tetap aman.
		var adminID int64
		err := dbtx.QueryRow(ctx,
			`SELECT id FROM users WHERE email = $1`, admin.Email,
		).Scan(&adminID)
		adminExists := err == nil
		if !adminExists {
			if _, err := dbtx.Exec(ctx,
				`INSERT INTO users (email, password, role) VALUES ($1, $2, 'admin')
	                 ON CONFLICT (email) DO NOTHING`,
				admin.Email, adminHash,
			); err != nil {
				return fmt.Errorf("seed: insert admin: %w", err)
			}
			summary.UsersCreated++
		} else {
			summary.UsersSkipped++
		}

		// Mahasiswa: untuk tiap baris, cek user dan student. Karena
		// (users.email) dan (students.nim) adalah UNIQUE, kita dapat
		// melakukan lookup lalu INSERT dengan ON CONFLICT DO NOTHING
		// untuk keandalan terhadap pemanggilan bersamaan.
		for i, s := range students {
			hash := studentHashes[i]

			var userID int64
			err := dbtx.QueryRow(ctx,
				`SELECT id FROM users WHERE email = $1`, s.Email,
			).Scan(&userID)
			userExists := err == nil
			if !userExists {
				if _, err := dbtx.Exec(ctx,
					`INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa')
	                     ON CONFLICT (email) DO NOTHING`,
					s.Email, hash,
				); err != nil {
					return fmt.Errorf("seed: insert user %s: %w", s.Email, err)
				}
				if err := dbtx.QueryRow(ctx,
					`SELECT id FROM users WHERE email = $1`, s.Email,
				).Scan(&userID); err != nil {
					return fmt.Errorf("seed: ambil user id untuk %s: %w", s.Email, err)
				}
				summary.UsersCreated++
			}

			var nimExists bool
			var existingStudentID int64
			err = dbtx.QueryRow(ctx,
				`SELECT id FROM students WHERE nim = $1`, s.NIM,
			).Scan(&existingStudentID)
			nimExists = err == nil
			if !nimExists {
				// Kirim NULL sungguhan untuk IPK; cast eksplisit
				// ($6::numeric) agar tidak bergantung pada default
				// encoding pgx untuk string kosong.
				var ipkArg any
				if s.IPK == nil {
					ipkArg = nil
				} else {
					ipkArg = *s.IPK
				}
				if _, err := dbtx.Exec(ctx,
					`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
	                     VALUES ($1, $2, $3, $4, $5, $6::numeric)
	                     ON CONFLICT (nim) DO NOTHING`,
					userID, s.NIM, s.Nama, s.Prodi, s.Angkatan, ipkArg,
				); err != nil {
					return fmt.Errorf("seed: insert student %s: %w", s.NIM, err)
				}
				summary.StudentsCreated++
			} else {
				summary.StudentsSkipped++
			}
		}

		for _, c := range courses {
			var existingCourseID int64
			err := dbtx.QueryRow(ctx,
				`SELECT id FROM courses WHERE kode_mk = $1`, c.KodeMK,
			).Scan(&existingCourseID)
			exists := err == nil
			if !exists {
				if _, err := dbtx.Exec(ctx,
					`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
	                     VALUES ($1, $2, $3, $4, $5)
	                     ON CONFLICT (kode_mk) DO NOTHING`,
					c.KodeMK, c.NamaMK, c.SKS, c.Semester, c.Kuota,
				); err != nil {
					return fmt.Errorf("seed: insert course %s: %w", c.KodeMK, err)
				}
				summary.CoursesCreated++
			} else {
				summary.CoursesSkipped++
			}
		}

		return nil
	})
	if err != nil {
		return summary, err
	}
	return summary, nil
}
