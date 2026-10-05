package migrations

import (
	"io"
	"sort"
	"strings"
	"testing"
)

func TestFS_Contains001Init(t *testing.T) {
	script, err := FS.ReadFile("001_init.sql")
	if err != nil {
		t.Fatalf("ReadFile 001_init.sql: %v", err)
	}
	body := string(script)
	for _, want := range []string{
		"CREATE TABLE users",
		"CREATE TABLE students",
		"CREATE TABLE courses",
		"CREATE TABLE enrollments",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("001_init.sql tidak mengandung %q", want)
		}
	}
}

func TestFS_OrderedSQLFiles(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		t.Fatalf("tidak ada file *.sql pada embed FS")
	}
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	for i := range names {
		if names[i] != sorted[i] {
			t.Fatalf("urutan embed berbeda dari sort asc: %v != %v", names, sorted)
		}
	}
	if names[0] != "001_init.sql" {
		t.Fatalf("file SQL pertama = %q; want 001_init.sql", names[0])
	}
}

func TestFS_NotEmptyAfterRead(t *testing.T) {
	f, err := FS.Open("001_init.sql")
	if err != nil {
		t.Fatalf("Open 001_init.sql: %v", err)
	}
	defer f.Close()
	buf := make([]byte, 64)
	n, err := f.Read(buf)
	if n == 0 || (err != nil && err != io.EOF) {
		t.Fatalf("baca 001_init.sql gagal: n=%d err=%v", n, err)
	}
}
