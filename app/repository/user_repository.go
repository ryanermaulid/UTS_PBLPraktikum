package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/ryanermaulid/UTS_PBLPraktikum/app/model"
	"github.com/ryanermaulid/UTS_PBLPraktikum/database"
)

// ErrUserNotFound dikembalikan ketika user tidak ada atau, untuk role
// mahasiswa, ketika baris students sudah di-soft delete. Service
// memanfaatkan sentinel ini untuk menjawab 401 tanpa membocorkan apakah
// email memang tidak ada atau akun dinonaktifkan.
var ErrUserNotFound = errors.New("user tidak ditemukan")

// UserRecord adalah representasi satu user yang dipakai service. Untuk
// mahasiswa, Student akan terisi; untuk admin, Student selalu nil.
type UserRecord struct {
	ID       int64
	Email    string
	Password string
	Role     string
	Student  *model.StudentMeResponse
}

// UserRepository adalah akses data untuk entitas user. Konstruktor
// menerima database.DBTX (pool atau transaction) sehingga service dapat
// menjalankan query dalam transaction bila diperlukan.
type UserRepository struct {
	db database.DBTX
}

func NewUserRepository(db database.DBTX) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByEmail mencari user berdasarkan email. Mengembalikan
// ErrUserNotFound bila:
//   - email tidak ditemukan, atau
//   - user adalah mahasiswa dan baris students terkait sudah di-soft
//     delete (deleted_at IS NOT NULL).
//
// Klausa tambahan `(u.role <> 'mahasiswa' OR s.user_id IS NOT NULL)`
// memastikan mahasiswa tanpa students aktif tidak lolos WHERE,
// sehingga Scan menerima ErrNoRows -> ErrUserNotFound.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*UserRecord, error) {
	const q = `
        SELECT u.id, u.email, u.password, u.role,
               s.id, s.nim, s.nama, s.prodi, s.angkatan
        FROM users u
        LEFT JOIN students s
               ON s.user_id = u.id
              AND s.deleted_at IS NULL
        WHERE u.email = $1
          AND (u.role <> 'mahasiswa' OR s.user_id IS NOT NULL)
        LIMIT 1`
	row := r.db.QueryRow(ctx, q, email)

	var (
		rec      UserRecord
		stuID    *int64
		stuNIM   *string
		stuNama  *string
		stuProdi *string
		stuAng   *int
	)
	if err := row.Scan(
		&rec.ID, &rec.Email, &rec.Password, &rec.Role,
		&stuID, &stuNIM, &stuNama, &stuProdi, &stuAng,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if stuID != nil {
		rec.Student = &model.StudentMeResponse{
			ID:       *stuID,
			NIM:      *stuNIM,
			Nama:     *stuNama,
			Prodi:    *stuProdi,
			Angkatan: *stuAng,
		}
	}
	return &rec, nil
}

// GetUserByIDWithStudent mengambil user berdasarkan id. Sama seperti
// GetUserByEmail, mahasiswa yang sudah di-soft delete diperlakukan
// sebagai user yang tidak ditemukan (ErrUserNotFound) karena klausa
// WHERE menyaring baris students yang sudah dihapus.
func (r *UserRepository) GetUserByIDWithStudent(ctx context.Context, id int64) (*UserRecord, error) {
	const q = `
        SELECT u.id, u.email, u.password, u.role,
               s.id, s.nim, s.nama, s.prodi, s.angkatan
        FROM users u
        LEFT JOIN students s
               ON s.user_id = u.id
              AND s.deleted_at IS NULL
        WHERE u.id = $1
          AND (u.role <> 'mahasiswa' OR s.user_id IS NOT NULL)
        LIMIT 1`
	row := r.db.QueryRow(ctx, q, id)

	var (
		rec      UserRecord
		stuID    *int64
		stuNIM   *string
		stuNama  *string
		stuProdi *string
		stuAng   *int
	)
	if err := row.Scan(
		&rec.ID, &rec.Email, &rec.Password, &rec.Role,
		&stuID, &stuNIM, &stuNama, &stuProdi, &stuAng,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if stuID != nil {
		rec.Student = &model.StudentMeResponse{
			ID:       *stuID,
			NIM:      *stuNIM,
			Nama:     *stuNama,
			Prodi:    *stuProdi,
			Angkatan: *stuAng,
		}
	}
	return &rec, nil
}
