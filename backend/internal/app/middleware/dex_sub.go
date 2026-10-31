package middleware

import (
	"encoding/base64"
	"regexp"
	"strings"
)

// uuidPattern matches a canonical RFC 4122 UUID string.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// DecodeDexSub extracts the plain user ID from a Dex OIDC subject claim.
//
// Dex' local (staticPasswords) connector does not hand the configured `userID`
// out verbatim. Instead it wraps it in a protobuf message of the form
//
//	\n<userID>\x12\x05local
//
// and base64-encodes the result into the JWT `sub` claim. This helper reverses
// that wrapping so callers can use the configured user ID (which is what the
// seeder stores in `organizers.oidc_sub`).
//
// For any subject that is not in that Dex-specific format (e.g. Rauthy's opaque
// IDs or an already-plain UUID) the input is returned unchanged, so the helper
// is safe to apply unconditionally.
func DecodeDexSub(sub string) string {
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return sub
	}

	raw, err := base64.StdEncoding.DecodeString(sub)
	if err != nil {
		// Dex uses standard base64; tolerate URL-safe encoding as a fallback.
		raw, err = base64.URLEncoding.DecodeString(sub)
		if err != nil {
			return sub
		}
	}

	// Protobuf: field 1 (tag 0x0A), length-delimited.
	if len(raw) < 2 || raw[0] != 0x0A {
		return sub
	}

	n := int(raw[1])
	if n <= 0 || 2+n > len(raw) {
		return sub
	}

	userID := string(raw[2 : 2+n])
	if !uuidPattern.MatchString(userID) {
		return sub
	}

	return userID
}
