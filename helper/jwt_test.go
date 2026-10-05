package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testJWTSecret = "rahasia-super-panjang-32-karakter-atau-lebih"

func TestIssueAndParseToken_RoundTrip(t *testing.T) {
	tok, err := IssueToken(testJWTSecret, 42, "admin", 60)
	if err != nil {
		t.Fatalf("IssueToken gagal: %v", err)
	}
	if tok == "" {
		t.Fatal("token kosong")
	}

	gotID, gotRole, err := ParseToken(testJWTSecret, tok)
	if err != nil {
		t.Fatalf("ParseToken gagal: %v", err)
	}
	if gotID != 42 {
		t.Errorf("userID = %d, want 42", gotID)
	}
	if gotRole != "admin" {
		t.Errorf("role = %q, want %q", gotRole, "admin")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	tok, err := IssueToken(testJWTSecret, 1, "mahasiswa", 60)
	if err != nil {
		t.Fatalf("IssueToken gagal: %v", err)
	}
	other := "secret-lain-yang-panjang-cukup-32+"
	if _, _, err := ParseToken(other, tok); err == nil {
		t.Fatal("expected error untuk secret berbeda")
	} else if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestIssueToken_InvalidTTL(t *testing.T) {
	if _, err := IssueToken(testJWTSecret, 1, "admin", 0); err == nil {
		t.Fatal("expected error untuk TTL nol")
	}
	if _, err := IssueToken(testJWTSecret, 1, "admin", -5); err == nil {
		t.Fatal("expected error untuk TTL negatif")
	}
}

func TestIssueToken_EmptySecret(t *testing.T) {
	if _, err := IssueToken("", 1, "admin", 60); err == nil {
		t.Fatal("expected error untuk secret kosong")
	}
}

func TestParseToken_EmptyInput(t *testing.T) {
	if _, _, err := ParseToken(testJWTSecret, ""); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestParseToken_Expired(t *testing.T) {
	tok := jwtExpiredToken(t, testJWTSecret)
	if _, _, err := ParseToken(testJWTSecret, tok); err == nil {
		t.Fatal("expected error untuk token kedaluwarsa")
	} else if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestParseToken_RejectsNoneAlg(t *testing.T) {
	malicious := jwtUnsignedToken(t)
	if _, _, err := ParseToken(testJWTSecret, malicious); err == nil {
		t.Fatal("expected error untuk token unsigned")
	} else if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestParseToken_TamperedSignature(t *testing.T) {
	tok, err := IssueToken(testJWTSecret, 1, "admin", 60)
	if err != nil {
		t.Fatalf("IssueToken gagal: %v", err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("bentuk token tidak valid: %q", tok)
	}
	// Decode signature, ganti byte di tengah, encode ulang agar pasti
	// berbeda dari nilai HMAC asli.
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decode signature gagal: %v", err)
	}
	mid := len(sig) / 2
	sig[mid] ^= 0xFF
	parts[2] = base64.RawURLEncoding.EncodeToString(sig)
	bad := strings.Join(parts, ".")
	if _, _, err := ParseToken(testJWTSecret, bad); err == nil {
		t.Fatal("expected error untuk signature yang diubah")
	} else if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

// jwtExpiredToken membangun token HS256 yang klaim exp-nya sudah lewat.
// Dipakai untuk menguji ParseToken tanpa bergantung pada waktu internal
// yang sulit diatur dari test.
func jwtExpiredToken(t *testing.T, secret string) string {
	t.Helper()
	header := `{"alg":"HS256","typ":"JWT"}`
	now := time.Now().Unix()
	payload := `{"sub":"1","role":"admin","iat":` + strconv.FormatInt(now-120, 10) + `,"exp":` + strconv.FormatInt(now-60, 10) + `}`
	signingInput := base64URLEncode([]byte(header)) + "." + base64URLEncode([]byte(payload))
	return signingInput + "." + hs256Sign(secret, signingInput)
}

func jwtUnsignedToken(t *testing.T) string {
	t.Helper()
	header := `{"alg":"none","typ":"JWT"}`
	payload := `{"sub":"1","role":"admin"}`
	return base64URLEncode([]byte(header)) + "." + base64URLEncode([]byte(payload)) + "."
}

func base64URLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func hs256Sign(secret, signingInput string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
