package helper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate dipakai oleh ValidateStruct. Instance dibuat sekali agar
// tidak membaca tag tiap kali validasi dijalankan.
var validate = validator.New(validator.WithRequiredStructEnabled())

// ValidateStruct menjalankan validasi terhadap struct menggunakan
// tag validate. Mengembalikan helper.Validation(fields) bila gagal,
// dengan nama field mengikuti tag json (atau tag struct bila json
// tidak ada). Pesan kesalahan memakai bahasa Indonesia.
//
// Untuk struct dengan nested struct bertag validate, validator akan
// menyelam ke field tersebut hanya jika tag divalidasi juga. Saat ini
// pemakaian utama adalah struct request satu tingkat.
func ValidateStruct(v any) error {
	if v == nil {
		return Validation(map[string][]string{"_": {"data kosong"}})
	}
	if reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil() {
		return Validation(map[string][]string{"_": {"data kosong"}})
	}

	if err := validate.Struct(v); err != nil {
		if ves, ok := err.(validator.ValidationErrors); ok {
			fields := map[string][]string{}
			for _, fe := range ves {
				name := jsonFieldName(v, fe)
				msg := indonesianMessage(fe)
				fields[name] = appendUnique(fields[name], msg)
			}
			return Validation(fields)
		}
		return fmt.Errorf("validasi gagal: %w", err)
	}
	return nil
}

// jsonFieldName mengembalikan nama field sesuai tag json bila ada, atau
// nama field struct (lowercase) bila tidak. Validator.Field() bisa
// berupa path bertitik untuk nested struct; kita hanya ambil segmen
// pertama karena pemakaian Tahap 3 tidak memerlukan nested validation.
func jsonFieldName(v any, fe validator.FieldError) string {
	name := fe.Field()
	if v != nil {
		if f, ok := reflect.TypeOf(v).Elem().FieldByName(fe.StructField()); ok {
			if tag := f.Tag.Get("json"); tag != "" {
				name = strings.SplitN(tag, ",", 2)[0]
			}
		}
	}
	return name
}

func indonesianMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		return fmt.Sprintf("minimal %s karakter", fe.Param())
	case "max":
		return fmt.Sprintf("maksimal %s karakter", fe.Param())
	case "oneof":
		return fmt.Sprintf("harus salah satu dari: %s", fe.Param())
	default:
		return fmt.Sprintf("tidak valid (%s)", fe.Tag())
	}
}

func appendUnique(s []string, v string) []string {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}
