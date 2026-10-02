package gorm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMigrate creates every model in a fresh database. This guards against
// column definitions SQLite does not understand - a `DEFAULT
// uuid_generate_v4()` aborts the migration with a syntax error.
func TestMigrate(t *testing.T) {
	store := newTestStore(t)

	for _, table := range []string{
		"event_years",
		"teams",
		"organizers",
		"participations",
		"participation_teams",
		"invites",
	} {
		require.Truef(t, store.db.Migrator().HasTable(table), "table %s is missing", table)
	}
}

func TestMigrate_IdKollisionMitUUID(t *testing.T) {
	store := newTestStore(t)

	// Exactly one id column, and it is the primary key. An embedded int64 ID
	// next to the uuid field would be shadowed silently.
	cols, err := store.db.Migrator().ColumnTypes(&dbOrganizer{})
	require.NoError(t, err)

	var idColumns int

	for _, col := range cols {
		if col.Name() == "id" {
			idColumns++
		}
	}

	require.Equal(t, 1, idColumns, "expected exactly one id column")
}

func TestMigrate_EventYearStringPK(t *testing.T) {
	store := newTestStore(t)

	cols, err := store.db.Migrator().ColumnTypes(&dbEventYear{})
	require.NoError(t, err)

	names := make([]string, 0, len(cols))
	for _, col := range cols {
		names = append(names, col.Name())
	}

	require.Contains(t, names, "year")
	// An embedded ID field must not appear here.
	require.NotContains(t, names, "id")
}

// TestMigrate_UniqueIndizes checks the constraints the domain logic relies on.
func TestMigrate_UniqueIndizes(t *testing.T) {
	store := newTestStore(t)

	cases := []struct {
		model  interface{}
		column string
		name   string
	}{
		{&dbOrganizer{}, "email", "idx_organizers_email"},
		{&dbOrganizer{}, "oidc_sub", "idx_organizers_oidc_sub"},
		{&dbTeam{}, "name", "idx_teams_name"},
		{&dbInvite{}, "email", "idx_invites_email"},
		{&dbInvite{}, "token", "idx_invites_token"},
	}

	for _, tc := range cases {
		require.Truef(t, store.db.Migrator().HasIndex(tc.model, tc.name),
			"expected unique index %s on %s", tc.name, tc.column)
	}
}

func TestPurgeAll_VerlaesstLeereDatenbank(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.PurgeAll())

	for _, table := range []string{"organizers", "participations", "invites", "teams"} {
		require.Truef(t, store.db.Migrator().HasTable(table), "table %s is missing after PurgeAll", table)
	}
}
