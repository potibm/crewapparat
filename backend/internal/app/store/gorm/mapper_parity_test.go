package gorm

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/domain"
)

// TestMapperRoundtrip maps a fully populated domain record to db and back,
// comparing field by field.
//
// This guards against fields missing from a mapper: OidcSub and Languages were
// silently dropped at one point, and that only surfaced once the data was
// already in the database.
func TestMapperRoundtrip(t *testing.T) {
	t.Run("Organizer", func(t *testing.T) {
		sub := "rauthy-sub-123"
		demozoo := uint(4711)
		in := &domain.Organizer{
			ID:                  uuid.New(),
			OidcSub:             &sub,
			Handle:              "demoscene-handle",
			Realname:            "Real Name",
			Email:               "crew@example.com",
			Phone:               "+49 123",
			SlackHandle:         "slack-handle",
			SlackProfileURL:     "https://slack.example/u/abc",
			MediaWikiProfileURL: "https://wiki.example.org/index.php/User:Test",
			AvatarURL:           "https://example.com/a.png",
			DemozooID:           &demozoo,
			FirstYearAsOrganizer: func() *int {
				v := 2019

				return &v
			}(),
			Pronouns:  "they/them",
			Languages: []string{"de", "en"},
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
		require.Equal(t, in.MediaWikiProfileURL, out.MediaWikiProfileURL)
		require.NotNil(t, out.FirstYearAsOrganizer)
		require.Equal(t, *in.FirstYearAsOrganizer, *out.FirstYearAsOrganizer)
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
			ShiftPreferredTimes:   []int{0, 1, 2, 32, 33},
			ShiftUnavailableTimes: []int{5, 6},
			ShiftPreferredPartner: []string{"uuid-a", "uuid-b"},
			Teams:                 []domain.Team{{ID: uuid.New(), Name: "Infodesk", HasShifts: true}},
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
			Token:         "8f14e45fceea167a5a36dedd4bea2543",
			ExpiresAt:     expires,
			Notes:         "extra notes",
			CoCAccepted:   true,
			CoCAcceptedAt: &cocAt,
			InvitedByID:   uuid.New(),
			ReviewedByID:  &reviewer,
		}

		out := toDomainInvite(fromDomainInvite(in))

		require.Equal(t, in.ID, out.ID)
		require.Equal(t, in.Email, out.Email)
		require.Equal(t, in.Status, out.Status)
		require.Equal(t, in.Token, out.Token)
		require.True(t, in.ExpiresAt.Equal(out.ExpiresAt))
		require.Equal(t, in.Notes, out.Notes)
		require.Equal(t, in.CoCAccepted, out.CoCAccepted)
		require.NotNil(t, out.CoCAcceptedAt)
		require.True(t, cocAt.Equal(*out.CoCAcceptedAt))
		require.Equal(t, in.InvitedByID, out.InvitedByID)
		require.NotNil(t, out.ReviewedByID)
		require.Equal(t, reviewer, *out.ReviewedByID)
	})

	t.Run("Team", func(t *testing.T) {
		in := &domain.Team{ID: uuid.New(), Name: "Infodesk", Color: "#FFD6A5", HasShifts: true, Active: true}
		require.Equal(t, *in, *toDomainTeam(fromDomainTeam(in)))
	})

	t.Run("EventYear", func(t *testing.T) {
		in := &domain.EventYear{Year: "2026", IsActive: true}
		require.Equal(t, *in, *toDomainEventYear(fromDomainEventYear(in)))
	})

	t.Run("CodeOfConduct", func(t *testing.T) {
		now := time.Now()
		in := &domain.CodeOfConduct{
			ID:          uuid.New(),
			Version:     "1.2.0",
			Status:      domain.CodeOfConductStatusPublished,
			ChangeNote:  "Clarified the reporting process",
			Text:        "# Hello\n\nWorld",
			PublishedAt: &now,
		}
		out := toDomainCodeOfConduct(fromDomainCodeOfConduct(in))
		require.Equal(t, in.ID, out.ID)
		require.Equal(t, in.Version, out.Version)
		require.Equal(t, in.Status, out.Status)
		require.Equal(t, in.ChangeNote, out.ChangeNote)
		require.Equal(t, in.Text, out.Text)
		require.NotNil(t, out.PublishedAt)
		require.True(t, now.Equal(*out.PublishedAt))
	})

	t.Run("CodeOfConductAcceptance", func(t *testing.T) {
		acceptedAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
		in := &domain.CodeOfConductAcceptance{
			ID:              uuid.New(),
			OrganizerID:     uuid.New(),
			CodeOfConductID: uuid.New(),
			AcceptedAt:      acceptedAt,
		}

		out := toDomainCodeOfConductAcceptance(fromDomainCodeOfConductAcceptance(in))

		require.Equal(t, in.ID, out.ID)
		require.Equal(t, in.OrganizerID, out.OrganizerID)
		require.Equal(t, in.CodeOfConductID, out.CodeOfConductID)
		require.True(t, acceptedAt.Equal(out.AcceptedAt))
	})
}

// apiOnlyFields exist only in the domain type, not in the db model. They are
// populated by the repository from joined data and never persisted.
var apiOnlyFields = map[string]bool{
	"CurrentParticipationStatus": true,
	"CurrentParticipationTeams":  true,
	"AcceptedCount":              true,
}

// infrastructureFields come from TimestampModel and AuditModel. They exist
// only in the db model, deliberately not in the domain type.
var infrastructureFields = map[string]bool{
	"CreatedAt":  true,
	"UpdatedAt":  true,
	"DeletedAt":  true,
	"CreatedBy":  true,
	"ModifiedBy": true,
	"DeletedBy":  true,
}

// TestFeldNamenParitaet compares the field names of domain and db struct.
// Every domain field needs a counterpart in the db model; conversely the db
// model may only add the infrastructure fields. That way no typo slips into the
// column mapping and no field stays unaccounted for.
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
		{"CodeOfConduct", reflect.TypeOf(domain.CodeOfConduct{}), reflect.TypeOf(dbCodeOfConduct{})},
		{
			"CodeOfConductAcceptance",
			reflect.TypeOf(domain.CodeOfConductAcceptance{}),
			reflect.TypeOf(dbCodeOfConductAcceptance{}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			domainFields := fieldNames(tc.domainTyp)
			dbFields := fieldNames(tc.dbTyp)

			for _, field := range domainFields {
				if apiOnlyFields[field] {
					continue
				}

				require.Containsf(t, dbFields, field,
					"domain field %q is missing in db model %s", field, tc.dbTyp)
			}

			for _, field := range dbFields {
				if infrastructureFields[field] {
					continue
				}

				require.Containsf(t, domainFields, field,
					"db field %q is missing in domain type %s", field, tc.domainTyp)
			}
		})
	}
}

// fieldNames resolves embedded structs (TimestampModel, AuditModel) so that
// only domain fields are compared.
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
