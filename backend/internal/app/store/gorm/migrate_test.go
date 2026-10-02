package gorm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMigrate legt alle Modelle in einer frischen Datenbank an. Der Test ist
// die Absicherung gegen Spaltendefinitionen, die SQLite nicht versteht - ein
// `DEFAULT uuid_generate_v4()` bricht die Migration mit einem Syntaxfehler ab.
func TestMigrate(t *testing.T) {
	store := newTestStore(t)

	for _, table := range []string{
		"event_years",
		"teams",
		"organizers",
		"participations",
		"participation_teams",
		"invites",
		"invite_teams",
	} {
		require.Truef(t, store.db.Migrator().HasTable(table), "Tabelle %s fehlt", table)
	}
}

func TestMigrate_IdKollisionMitUUID(t *testing.T) {
	store := newTestStore(t)

	// Genau eine id-Spalte, und sie ist der Primärschlüssel. Ein eingebettetes
	// int64-ID neben dem uuid-Feld würde still überschrieben.
	cols, err := store.db.Migrator().ColumnTypes(&dbOrganizer{})
	require.NoError(t, err)

	var idColumns int

	for _, col := range cols {
		if col.Name() == "id" {
			idColumns++
		}
	}

	require.Equal(t, 1, idColumns, "erwartet genau eine id-Spalte")
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
	// Ein eingebettetes ID-Feld darf hier nicht auftauchen.
	require.NotContains(t, names, "id")
}

// TestMigrate_UniqueIndizes prüft die Constraints, auf die sich die
// Geschäftslogik verlässt.
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
		{&dbInvite{}, "token_hash", "idx_invites_token_hash"},
	}

	for _, tc := range cases {
		require.Truef(t, store.db.Migrator().HasIndex(tc.model, tc.name),
			"erwartet Unique-Index %s auf %s", tc.name, tc.column)
	}
}

func TestPurgeAll_VerlaesstLeereDatenbank(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.PurgeAll())

	for _, table := range []string{"organizers", "participations", "invites", "teams"} {
		require.Truef(t, store.db.Migrator().HasTable(table), "Tabelle %s fehlt nach PurgeAll", table)
	}
}
