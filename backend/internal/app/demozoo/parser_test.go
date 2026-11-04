package demozoo

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleDump = "COPY public.demoscene_membership (id, member_id, group_id, is_current, data_source) FROM stdin;\n" +
	"1\t99\t100\tt\t\\N\n" +
	"2\t101\t100\tf\t\\N\n" +
	"\\.\n" +
	"COPY public.demoscene_releaser (foo_id, name, is_group, country_code, id, notes) FROM stdin;\n" +
	"99\tAlpha\tf\tDE\t99\t\\N\n" +
	"100\tBeta Group\tt\t\t100\t\\N\n" +
	"101\tGamma\tf\t\t101\t\\N\n" +
	"\\.\n"

func TestParseDumpReordersTablesAndColumns(t *testing.T) {
	sceners, err := ParseDump(strings.NewReader(sampleDump))
	if err != nil {
		t.Fatalf("ParseDump failed: %v", err)
	}

	if len(sceners) != 2 {
		t.Fatalf("expected 2 sceners, got %d", len(sceners))
	}

	if sceners[0].ID != 99 {
		t.Fatalf("expected first scener id 99, got %d", sceners[0].ID)
	}

	if sceners[0].Name != "Alpha" {
		t.Fatalf("expected Alpha, got %s", sceners[0].Name)
	}

	if sceners[0].CountryCode != "DE" {
		t.Fatalf("expected DE, got %s", sceners[0].CountryCode)
	}

	if len(sceners[0].Groups) != 1 || sceners[0].Groups[0] != "Beta Group" {
		t.Fatalf("expected group Beta Group, got %v", sceners[0].Groups)
	}

	if sceners[1].ID != 101 {
		t.Fatalf("expected second scener id 101, got %d", sceners[1].ID)
	}

	if len(sceners[1].Groups) != 0 {
		t.Fatalf("expected no groups, got %v", sceners[1].Groups)
	}
}

func TestWriteCSV(t *testing.T) {
	sceners := []Scener{
		{ID: 1, Name: "A", CountryCode: "DE", Groups: []string{"G1", "G2"}},
		{ID: 2, Name: "B", CountryCode: "", Groups: nil},
	}

	var buf bytes.Buffer
	if err := WriteCSV(&buf, sceners); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	want := "id,name,country_code,groups\n1,A,DE,G1;G2\n2,B,,\n"
	if buf.String() != want {
		t.Fatalf("unexpected CSV:\n got: %q\nwant: %q", buf.String(), want)
	}
}

func TestParseDumpUnescape(t *testing.T) {
	dump := "COPY public.demoscene_releaser (id, name, is_group) FROM stdin;\n" +
		"5\tLine\\nBreak\tf\n" +
		"\\.\n"

	sceners, err := ParseDump(strings.NewReader(dump))
	if err != nil {
		t.Fatalf("ParseDump failed: %v", err)
	}

	if len(sceners) != 1 || sceners[0].Name != "Line\nBreak" {
		t.Fatalf("unescape failed: %v", sceners)
	}
}

func TestDemozooExportCommandCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()

	input := filepath.Join(tmpDir, "dump.sql")
	if err := os.WriteFile(input, []byte(sampleDump), 0o600); err != nil {
		t.Fatal(err)
	}

	output := filepath.Join(tmpDir, "out", "sceners.csv")

	// Invoke via the command factory is not wired here; instead exercise parser+writer.
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sceners, err := ParseDump(f)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		t.Fatal(err)
	}

	out, err := os.Create(output)
	if err != nil {
		t.Fatal(err)
	}

	if err := WriteCSV(out, sceners); err != nil {
		out.Close()
		t.Fatal(err)
	}

	out.Close()

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "Alpha") {
		t.Fatalf("output missing Alpha: %s", data)
	}
}
