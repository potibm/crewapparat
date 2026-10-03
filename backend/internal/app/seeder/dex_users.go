package seeder

import (
	"github.com/potibm/crewapparat/internal/app/domain"
)

// The two Dex users below are the organizers a developer logs in as during
// local work. Their oidc_sub values are the `userID` entries of the matching
// accounts in infra/dex-config.yaml: Dex' local connector hands that field on
// as the OIDC `sub` claim.
//
// Seeding the sub explicitly means the organizer records are already linked on
// first login. TestDexUsersMatchDexConfig fails as soon as one of the userIDs in
// the Dex config is rotated, so the values cannot silently drift apart.
//
// Note that Dex runs with `storage: type: memory`. That regenerates signing
// keys on every restart, but not user IDs - those come from the config file.

const (
	AdminEmail = "admin@example.com"
	UserEmail  = "user@example.com"

	AdminOidcSub = "92c18da6-46e6-44e1-b462-08bfb8eb2f38"
	UserOidcSub  = "b0c35c31-a681-4298-81ea-c7878c5d0e7e"
)

// DexUser describes one of the pre-seeded organizers that pairs with a local
// Dex account.
type DexUser struct {
	Email     string
	OidcSub   string
	Handle    string
	Realname  string
	Pronouns  string
	Languages []string
	Teams     []string
}

// DexUsers returns the two organizers every local environment gets.
func DexUsers() []DexUser {
	return []DexUser{
		{
			Email:     AdminEmail,
			OidcSub:   AdminOidcSub,
			Handle:    "admin",
			Realname:  "Alex Admin",
			Pronouns:  "they/them",
			Languages: []string{"de", "en"},
			Teams:     []string{"Infodesk", "Kasse"},
		},
		{
			Email:     UserEmail,
			OidcSub:   UserOidcSub,
			Handle:    "crew",
			Realname:  "Sam Crew",
			Pronouns:  "she/her",
			Languages: []string{"de"},
			Teams:     []string{"Netzwerk"},
		},
	}
}

// ToOrganizer turns the definition into a domain record.
func (u DexUser) ToOrganizer() *domain.Organizer {
	sub := u.OidcSub

	return &domain.Organizer{
		OidcSub:     &sub,
		Email:       u.Email,
		Handle:      u.Handle,
		Realname:    u.Realname,
		Pronouns:    u.Pronouns,
		Languages:   u.Languages,
		Phone:       "+49 30 000000",
		SlackHandle: u.Handle,
	}
}

// DefaultActiveYear is the event year the seed marks as active.
const DefaultActiveYear = "2026"

// previousYears are seeded inactive so the annual check-in has a prior year to
// prefill from.
var previousYears = []string{"2025"}
