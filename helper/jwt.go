package helper

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken dikembalikan ketika token tidak valid karena bentuk,
// signature, atau klaim yang tidak sesuai. Pemakai dapat membandingkan
// dengan errors.Is untuk membedakan dengan error lain tanpa membuka
// pesan internal ke client.
var ErrInvalidToken = errors.New("token tidak valid")

// jwtClaims adalah klaim internal yang dipakai aplikasi. Field sub berisi
// id user (string) dan role menyimpan peran (admin/mahasiswa).
type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// IssueToken membuat token JWT HS256 dengan klaim sub=userID, role,
// iat, dan exp=now+ttl. ttlMinutes <= 0 dianggap tidak valid karena
// akan menghasilkan token yang langsung kedaluwarsa.
func IssueToken(secret string, userID int64, role string, ttlMinutes int) (string, error) {
	if secret == "" {
		return "", errors.New("secret JWT kosong")
	}
	if ttlMinutes <= 0 {
		return "", errors.New("durasi token tidak valid")
	}

	now := time.Now()
	claims := jwtClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttlMinutes) * time.Minute)),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// ParseToken memverifikasi signature HS256 dan mengembalikan userID serta
// role yang tersimpan dalam klaim. Mengembalikan ErrInvalidToken (atau
// error jwt bawaan) bila token rusak, kedaluwarsa, atau menggunakan
// algoritma lain.
func ParseToken(secret, raw string) (int64, string, error) {
	if secret == "" {
		return 0, "", errors.New("secret JWT kosong")
	}
	if raw == "" {
		return 0, "", ErrInvalidToken
	}

	claims := &jwtClaims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: algoritma tidak didukung", ErrInvalidToken)
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !tok.Valid {
		return 0, "", ErrInvalidToken
	}
	if claims.Subject == "" {
		return 0, "", fmt.Errorf("%w: klaim subject kosong", ErrInvalidToken)
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("%w: sub bukan angka", ErrInvalidToken)
	}
	return id, claims.Role, nil
}
