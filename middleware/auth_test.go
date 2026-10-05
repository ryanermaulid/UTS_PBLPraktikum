package middleware

import "testing"

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"only scheme", "Bearer", ""},
		{"valid", "Bearer abc.def.ghi", "abc.def.ghi"},
		{"valid with extra space", "Bearer   xyz", "xyz"},
		{"lowercase scheme", "bearer abc", "abc"},
		{"wrong scheme", "Basic abc", ""},
		{"two spaces one value", "Bearer abc def", "abc def"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bearerToken(tt.in)
			if got != tt.want {
				t.Errorf("bearerToken(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
