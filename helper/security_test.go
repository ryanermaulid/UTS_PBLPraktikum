package helper

import (
	"strings"
	"testing"
)

func TestHashPassword_DifferentFromPlain(t *testing.T) {
	const plain = "Admin12345!"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword(%q) error: %v", plain, err)
	}
	if hash == "" {
		t.Fatalf("hash kosong")
	}
	if hash == plain {
		t.Fatalf("hash sama dengan plain; bcrypt tidak dipakai")
	}
	// bcrypt menghasilkan string yang dimulai dengan $2
	if !strings.HasPrefix(hash, "$2") {
		t.Fatalf("hash %q tidak berformat bcrypt", hash)
	}
}

func TestHashPassword_RejectsEmpty(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatalf("HashPassword(\"\") seharusnya mengembalikan error")
	}
}

func TestCheckPassword_MatchAndMismatch(t *testing.T) {
	const plain = "Admin12345!"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !CheckPassword(hash, plain) {
		t.Fatalf("CheckPassword(hash, plain) seharusnya true")
	}
	if CheckPassword(hash, "salah") {
		t.Fatalf("CheckPassword(hash, \"salah\") seharusnya false")
	}
}

func TestCheckPassword_EmptySides(t *testing.T) {
	if CheckPassword("", "x") {
		t.Fatalf("hash kosong seharusnya false")
	}
	if CheckPassword("$2a$10$abcdefghijklmnopqrstuv", "") {
		t.Fatalf("plain kosong seharusnya false")
	}
}
