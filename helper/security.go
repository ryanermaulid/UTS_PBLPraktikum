package helper

import "golang.org/x/crypto/bcrypt"

// HashPassword mengembalikan hash bcrypt dari password plain. Cost
// default dipakai; nilainya cukup kuat untuk produksi dan tetap cepat
// untuk dipakai di seeder.
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", ErrEmptyPassword
	}
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// CheckPassword membandingkan hash bcrypt dengan password plain.
// Mengembalikan false jika hash atau plain kosong, atau jika hash tidak
// cocok. Pesan error internal tidak dibocorkan.
func CheckPassword(hash, plain string) bool {
	if hash == "" || plain == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
