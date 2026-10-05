package model

// LoginRequest adalah body POST /auth/login. Email mengikuti format email
// standar; password minimal 8 karakter sesuai SPEC.md.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginUser adalah bagian user dari response login. Field mengikuti
// SPEC.md: id, email, role.
type LoginUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// LoginResponse adalah response POST /auth/login sukses.
type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	User        LoginUser `json:"user"`
}

// UserMeResponse adalah response GET /auth/me untuk semua role.
// Untuk mahasiswa, Student != nil dan mengikuti StudentMeResponse.
type UserMeResponse struct {
	ID      int64              `json:"id"`
	Email   string             `json:"email"`
	Role    string             `json:"role"`
	Student *StudentMeResponse `json:"student,omitempty"`
}

// StudentMeResponse adalah data mahasiswa yang disisipkan pada
// /auth/me. Field id milik tabel students (dibutuhkan antarmuka untuk
// memanggil endpoint lain seperti PUT /students/{id}).
type StudentMeResponse struct {
	ID       int64  `json:"id"`
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}
