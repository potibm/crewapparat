package gorm

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/domain"
)

// TestMapperRoundtrip mappt einen vollständig befüllten Domain-Satz nach db
// und zurück und vergleicht Feld für Feld.
//
// Das ist die Absicherung gegen fehlende Felder im Mapper: einst wurden OidcSub
// und Languages stillschweigend verworfen, und das fiel erst auf, wenn die
// Daten schon in der Datenbank standen.
func TestMapperRoundtrip(t *testing.T) {
	t.Run("Organizer", func(t *testing.T) {
		sub := "rauthy-sub-123"
		demozoo := uint(4711)
		in := &domain.Organizer{
			ID:              uuid.New(),
			OidcSub:         &sub,
			Handle:          "demoscene-handle",
			Realname:        "Real Name",
			Email:           "crew@example.com",
			Phone:           "+49 123",
			SlackHandle:     "slack-handle",
			SlackProfileURL: "https://slack.example/u/abc",
			AvatarURL:       "https://example.com/a.png",
			DemozooID:       &demozoo,
			Pronouns:        "they/them",
			Languages:       []string{"de", "en"},
		}

		out := toDomainOrganizer(fromDomainOrganizer(in))

		require.Equal(t, in.ID, out.ID)
		require.NotNil(t, out.OidcSub, "OidcSub wurde im Mapper verworfen")
		require.Equal(t, sub, *out.OidcSub)
		require.Equal(t, in.Handle, out.Handle)
		require.Equal(t, in.Realname, out.Realname)
		require.Equal(t, in.Email, out.Email)
		require.Equal(t, in.Phone, out.Phone)
		require.Equal(t, in.SlackHandle, out.SlackHandle)
		require.Equal(t, in.SlackProfileURL, out.SlackProfileURL)
		require.Equal(t, in.AvatarURL, out.AvatarURL)
		require.NotNil(t, out.DemozooID)
		require.Equal(t, demozoo, *out.DemozooID)
		require.Equal(t, in.Pronouns, out.Pronouns)
		require.Equal(t, in.Languages, out.Languages)
	})

	t.Run("Participation", func(t *testing.T) {
		in := &domain.Participation{
			ID:                    uuid.New(),
			OrganizerID:           uuid.New(),
			EventYearID:           "2026",
			Status:                domain.ParticipationStatusConfirmed,
			TShirtSize:            "XL",
			AdditionalPaidShirt:   true,
			PublicWebsiteConsent:  true,
			ShiftAvailableTimes:   []int{0, 1, 2, 32, 33},
			ShiftUnavailableTimes: []int{5, 6},
			ShiftPreferredPartner: []string{"uuid-a", "uuid-b"},
			Teams:                 []domain.Team{{ID: uuid.New(), Name: "Infodesk"}},
		}

		out := toDomainParticipation(fromDomainParticipation(in))

		require.Equal(t, *in, *out)
	})

	t.Run("Invite", func(t *testing.T) {
		cocAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
		reviewer := uuid.New()
		expires := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		in := &domain.Invite{
			ID:            uuid.New(),
			Email:         "new@example.com",
			Status:        domain.InviteStatusPendingReview,
			TokenHash:     "e3b0c44298fc1c149afbf4c8996fb924",
			ExpiresAt:     expires,
			ProvidedName:  "Provided Name",
			ProvidedPhone: "+49 999",
			Notes:         "extra notes",
			CoCAccepted:   true,
			CoCAcceptedAt: &cocAt,
			InvitedByID:   uuid.New(),
			ReviewedByID:  &reviewer,
			Teams:         []domain.Team{{ID: uuid.New(), Name: "Network"}},
		}

		out := toDomainInvite(fromDomainInvite(in))

		require.Equal(t, in.ID, out.ID)
		require.Equal(t, in.Email, out.Email)
		require.Equal(t, in.Status, out.Status)
		require.Equal(t, in.TokenHash, out.TokenHash)
		require.True(t, in.ExpiresAt.Equal(out.ExpiresAt))
		require.Equal(t, in.ProvidedName, out.ProvidedName)
		require.Equal(t, in.ProvidedPhone, out.ProvidedPhone)
		require.Equal(t, in.Notes, out.Notes)
		require.Equal(t, in.CoCAccepted, out.CoCAccepted)
		require.NotNil(t, out.CoCAcceptedAt)
		require.True(t, cocAt.Equal(*out.CoCAcceptedAt))
		require.Equal(t, in.InvitedByID, out.InvitedByID)
		require.NotNil(t, out.ReviewedByID)
		require.Equal(t, reviewer, *out.ReviewedByID)
		require.Equal(t, in.Teams, out.Teams)
	})

	t.Run("Team", func(t *testing.T) {
		in := &domain.Team{ID: uuid.New(), Name: "Infodesk"}
		require.Equal(t, *in, *toDomainTeam(fromDomainTeam(in)))
	})

	t.Run("EventYear", func(t *testing.T) {
		in := &domain.EventYear{Year: "2026", IsActive: true}
		require.Equal(t, *in, *toDomainEventYear(fromDomainEventYear(in)))
	})
}

// infrastructureFields kommen aus TimestampModel und AuditModel. Sie stehen
// absichtlich nur im db-Modell und nicht im Domain-Typ.
var infrastructureFields = map[string]bool{
	"CreatedAt":  true,
	"UpdatedAt":  true,
	"DeletedAt":  true,
	"CreatedBy":  true,
	"ModifiedBy": true,
	"DeletedBy":  true,
}

// TestFeldNamenParitaet vergleicht die Feldnamen von Domain- und db-Struct.
// Jedes Domain-Feld muss eine Entsprechung im db-Modell haben; umgekehrt darf
// das db-Modell nur die Infrastruktur-Felder mehr haben. So schleicht sich kein
// Tippfehler in der Spaltenzuordnung ein und kein Feld bleibt unberücksichtigt.
func TestFeldNamenParitaet(t *testing.T) {
	cases := []struct {
		name      string
		domainTyp reflect.Type
		dbTyp     reflect.Type
	}{
		{"Organizer", reflect.TypeOf(domain.Organizer{}), reflect.TypeOf(dbOrganizer{})},
		{"Participation", reflect.TypeOf(domain.Participation{}), reflect.TypeOf(dbParticipation{})},
		{"Team", reflect.TypeOf(domain.Team{}), reflect.TypeOf(dbTeam{})},
		{"EventYear", reflect.TypeOf(domain.EventYear{}), reflect.TypeOf(dbEventYear{})},
		{"Invite", reflect.TypeOf(domain.Invite{}), reflect.TypeOf(dbInvite{})},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			domainFields := fieldNames(tc.domainTyp)
			dbFields := fieldNames(tc.dbTyp)

			for _, field := range domainFields {
				require.Containsf(t, dbFields, field,
					"Domain-Feld %q fehlt im db-Modell %s", field, tc.dbTyp)
			}

			for _, field := range dbFields {
				if infrastructureFields[field] {
					continue
				}

				require.Containsf(t, domainFields, field,
					"db-Feld %q fehlt im Domain-Typ %s", field, tc.domainTyp)
			}
		})
	}
}

// fieldNames löst eingebettete Structs (TimestampModel, AuditModel) auf, damit
// nur die fachlichen Felder verglichen werden.
func fieldNames(typ reflect.Type) []string {
	names := make([]string, 0, typ.NumField())

	for i := range typ.NumField() {
		field := typ.Field(i)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			names = append(names, fieldNames(field.Type)...)

			continue
		}

		names = append(names, field.Name)
	}

	return names
}
