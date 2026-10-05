package database

import (
	"regexp"
	"testing"

	"github.com/ryanermaulid/UTS_PBLPraktikum/helper"
)

func TestAdminSeed_EmailAndPassword(t *testing.T) {
	a := AdminSeedData()
	if a.Email != AdminEmail {
		t.Fatalf("email admin = %q; want %q", a.Email, AdminEmail)
	}
	if a.Plaintext == "" {
		t.Fatalf("password admin kosong")
	}
	if len(a.Plaintext) < 8 {
		t.Fatalf("password admin %q panjangnya < 8", a.Plaintext)
	}
}

func TestStudentSeedData_CountAndUniqueness(t *testing.T) {
	s := StudentSeedData()
	if len(s) != 20 {
		t.Fatalf("jumlah mahasiswa = %d; want 20", len(s))
	}

	nimRe := regexp.MustCompile(`^[0-9]{12}$`)
	seenNIM := make(map[string]struct{}, len(s))
	seenEmail := make(map[string]struct{}, len(s))
	seenProdi := make(map[string]struct{})
	for _, p := range ProdiValues {
		seenProdi[p] = struct{}{}
	}

	for i, x := range s {
		if !nimRe.MatchString(x.NIM) {
			t.Fatalf("mahasiswa ke-%d: NIM %q tidak cocok ^[0-9]{12}$", i, x.NIM)
		}
		if _, dup := seenNIM[x.NIM]; dup {
			t.Fatalf("NIM duplikat: %s", x.NIM)
		}
		seenNIM[x.NIM] = struct{}{}

		if _, dup := seenEmail[x.Email]; dup {
			t.Fatalf("email duplikat: %s", x.Email)
		}
		seenEmail[x.Email] = struct{}{}

		if x.Email == "" {
			t.Fatalf("mahasiswa %s: email kosong", x.NIM)
		}
		if x.Nama == "" {
			t.Fatalf("mahasiswa %s: nama kosong", x.NIM)
		}
		if _, ok := seenProdi[x.Prodi]; !ok {
			t.Fatalf("mahasiswa %s: prodi %q bukan dari %v", x.NIM, x.Prodi, ProdiValues)
		}
		if x.Angkatan < 1000 || x.Angkatan > 9999 {
			t.Fatalf("mahasiswa %s: angkatan %d di luar jangkauan 1000-9999", x.NIM, x.Angkatan)
		}
		if x.Angkatan > 2026 {
			t.Fatalf("mahasiswa %s: angkatan %d lebih besar dari tahun berjalan (2026)", x.NIM, x.Angkatan)
		}
		if x.IPK != nil {
			v := *x.IPK
			if v < 0 || v > 4.0 {
				t.Fatalf("mahasiswa %s: ipk %v di luar jangkauan 0.00-4.00", x.NIM, v)
			}
		}
		if x.Plaintext == "" {
			t.Fatalf("mahasiswa %s: password kosong", x.NIM)
		}
		if x.Plaintext != x.NIM {
			t.Fatalf("mahasiswa %s: password %q tidak sama dengan NIM", x.NIM, x.Plaintext)
		}
		if _, err := helper.HashPassword(x.Plaintext); err != nil {
			t.Fatalf("mahasiswa %s: gagal hash password: %v", x.NIM, err)
		}
	}
}

func TestStudentSeedData_IPKCases(t *testing.T) {
	s := StudentSeedData()
	want := map[float64]bool{
		3.00: false, 3.45: false, 2.99: false, 2.50: false,
		2.49: false, 1.80: false, 4.00: false,
	}
	nullSeen := false
	for _, x := range s {
		if x.IPK == nil {
			nullSeen = true
			continue
		}
		v := *x.IPK
		if _, ok := want[v]; ok && !want[v] {
			want[v] = true
		}
	}
	for v, seen := range want {
		if !seen {
			t.Fatalf("kasus IPK %.2f tidak ditemukan pada data seed", v)
		}
	}
	if !nullSeen {
		t.Fatalf("tidak ada mahasiswa dengan IPK NULL pada data seed")
	}
}

func TestCourseSeedData_CountAndUniqueness(t *testing.T) {
	c := CourseSeedData()
	if len(c) != 10 {
		t.Fatalf("jumlah mata kuliah = %d; want 10", len(c))
	}

	seenKode := make(map[string]struct{}, len(c))
	quotaOneSeen := false
	quotaTwoSeen := false
	for i, x := range c {
		if x.KodeMK == "" {
			t.Fatalf("course ke-%d: kode_mk kosong", i)
		}
		if _, dup := seenKode[x.KodeMK]; dup {
			t.Fatalf("kode_mk duplikat: %s", x.KodeMK)
		}
		seenKode[x.KodeMK] = struct{}{}

		if x.NamaMK == "" {
			t.Fatalf("course %s: nama_mk kosong", x.KodeMK)
		}
		if x.SKS <= 0 {
			t.Fatalf("course %s: sks %d tidak > 0", x.KodeMK, x.SKS)
		}
		if x.Semester <= 0 {
			t.Fatalf("course %s: semester %d tidak > 0", x.KodeMK, x.Semester)
		}
		if x.Kuota < 0 {
			t.Fatalf("course %s: kuota %d tidak >= 0", x.KodeMK, x.Kuota)
		}
		if x.Kuota == 1 {
			quotaOneSeen = true
		}
		if x.Kuota == 2 {
			quotaTwoSeen = true
		}
	}
	if !quotaOneSeen {
		t.Fatalf("tidak ada mata kuliah dengan kuota=1 untuk uji kuota penuh")
	}
	if !quotaTwoSeen {
		t.Fatalf("tidak ada mata kuliah dengan kuota=2 untuk uji kuota penuh")
	}
}

func TestCourseSeedData_SKSRangePositive(t *testing.T) {
	c := CourseSeedData()
	for _, x := range c {
		if x.SKS < 2 || x.SKS > 4 {
			t.Fatalf("course %s: sks %d di luar rentang 2..4", x.KodeMK, x.SKS)
		}
	}
}
