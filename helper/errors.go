package helper

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError adalah error terstruktur yang dipahami ErrorHandler. Field Err
// hanya untuk log internal; tidak pernah dikirim ke client.
type AppError struct {
	Status  int
	Message string
	Fields  map[string][]string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// Unauthorized dipakai untuk 401 (token tidak ada/salah, kredensial login
// tidak valid).
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Tidak terotorisasi"
	}
	return &AppError{Status: http.StatusUnauthorized, Message: message}
}

// Forbidden dipakai untuk 403 (role atau kepemilikan tidak sesuai).
func Forbidden(message string) *AppError {
	if message == "" {
		message = "Akses ditolak"
	}
	return &AppError{Status: http.StatusForbidden, Message: message}
}

// NotFound dipakai untuk 404.
func NotFound(message string) *AppError {
	if message == "" {
		message = "Data tidak ditemukan"
	}
	return &AppError{Status: http.StatusNotFound, Message: message}
}

// Conflict dipakai untuk 409 (mis. duplikasi resource).
func Conflict(message string) *AppError {
	if message == "" {
		message = "Konflik data"
	}
	return &AppError{Status: http.StatusConflict, Message: message}
}

// Unprocessable dipakai untuk 422 (aturan bisnis atau body rusak).
func Unprocessable(message string) *AppError {
	if message == "" {
		message = "Permintaan tidak dapat diproses"
	}
	return &AppError{Status: http.StatusUnprocessableEntity, Message: message}
}

// TooManyRequests dipakai untuk 429.
func TooManyRequests(message string) *AppError {
	if message == "" {
		message = "Terlalu banyak percobaan"
	}
	return &AppError{Status: http.StatusTooManyRequests, Message: message}
}

// Validation membungkus kesalahan per-field menjadi 422 dengan pesan
// standar "Validasi gagal".
func Validation(fields map[string][]string) *AppError {
	return &AppError{
		Status:  http.StatusUnprocessableEntity,
		Message: "Validasi gagal",
		Fields:  fields,
	}
}

// Internal membungkus error tak terduga menjadi 500 dengan pesan generik.
// Error asli hanya tersimpan di Err untuk ditulis ke log oleh ErrorHandler.
func Internal(err error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Message: "Terjadi kesalahan pada server",
		Err:     err,
	}
}

// AsAppError mengekstrak *AppError dari error biasa, atau mengembalikan
// (nil, false) bila bukan AppError.
func AsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
