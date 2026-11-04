package demozoo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCSV_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sceners.csv")
	sceners := []Scener{
		{ID: 1, Name: "Alpha", CountryCode: "DE", Groups: []string{"G1", "G2"}},
		{ID: 2, Name: "Beta", CountryCode: "", Groups: nil},
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := WriteCSV(f, sceners); err != nil {
		f.Close()
		t.Fatal(err)
	}

	f.Close()

	got, err := ReadCSV(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}

	if got[0].ID != 1 || got[0].Name != "Alpha" || got[0].CountryCode != "DE" {
		t.Fatalf("bad first: %+v", got[0])
	}

	if len(got[0].Groups) != 2 || got[0].Groups[0] != "G1" || got[0].Groups[1] != "G2" {
		t.Fatalf("bad groups: %+v", got[0].Groups)
	}

	if got[1].ID != 2 || len(got[1].Groups) != 0 {
		t.Fatalf("bad second: %+v", got[1])
	}
}

func TestReadCSV_MissingFile(t *testing.T) {
	_, err := ReadCSV(filepath.Join(t.TempDir(), "missing.csv"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadCSV_EmptyGroups(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "empty.csv")
	if err := os.WriteFile(path, []byte("id,name,country_code,groups\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := ReadCSV(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 0 {
		t.Fatalf("expected 0, got %d", len(got))
	}
}
