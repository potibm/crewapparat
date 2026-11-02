package middleware

import (
	"encoding/base64"
	"testing"
)

func TestDecodeDexSub_DexEncoded(t *testing.T) {
	const encoded = "CiRiMGMzNWMzMS1hNjgxLTQyOTgtODFlYS1jNzg3OGM1ZDBlN2USBWxvY2Fs"

	got := DecodeDexSub(encoded)

	want := "b0c35c31-a681-4298-81ea-c7878c5d0e7e"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDecodeDexSub_AdminEncoded(t *testing.T) {
	// admin userID from infra/dex-config.yaml
	const encoded = "CiQ5MmMxOGRhNi00NmU2LTQ0ZTEtYjQ2Mi0wOGJmYjhlYjJmMzgSBWxvY2Fs"
	// build expected by decoding manually
	raw, _ := base64.StdEncoding.DecodeString(encoded)
	// field 1 length-delimited
	n := int(raw[1])
	want := string(raw[2 : 2+n])

	got := DecodeDexSub(encoded)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDecodeDexSub_PlainUUID(t *testing.T) {
	const uuid = "92c18da6-46e6-44e1-b462-08bfb8eb2f38"
	if got := DecodeDexSub(uuid); got != uuid {
		t.Fatalf("got %q, want %q", got, uuid)
	}
}

func TestDecodeDexSub_RauthyOpaque(t *testing.T) {
	const id = "aMxh3stPIDrIfOjExEedYEyq"
	if got := DecodeDexSub(id); got != id {
		t.Fatalf("got %q, want %q", got, id)
	}
}

func TestDecodeDexSub_Empty(t *testing.T) {
	if got := DecodeDexSub(""); got != "" {
		t.Fatalf("got %q, want empty", got)
	}

	if got := DecodeDexSub("   "); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestDecodeDexSub_Garbage(t *testing.T) {
	cases := []string{"not-base64!!!", "aGVsbG8=", "CiQ=", "CiRub3RhdXVpZA=="}
	for _, c := range cases {
		if got := DecodeDexSub(c); got != c {
			t.Errorf("input %q: got %q, want unchanged", c, got)
		}
	}
}

func TestDecodeDexSub_URLSafe(t *testing.T) {
	// standard encoded uses +/; ensure URL-safe variant also decodes
	encoded := "CiRiMGMzNWMzMS1hNjgxLTQyOTgtODFlYS1jNzg3OGM1ZDBlN2USBWxvY2Fs"
	urlsafe := base64.URLEncoding.EncodeToString(mustDecode(t, encoded))

	got := DecodeDexSub(urlsafe)
	if got != "b0c35c31-a681-4298-81ea-c7878c5d0e7e" {
		t.Fatalf("got %q", got)
	}
}

func mustDecode(t *testing.T, s string) []byte {
	t.Helper()

	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}

	return b
}
