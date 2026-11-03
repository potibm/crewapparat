package seeder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
	"github.com/potibm/crewapparat/internal/app/store/gorm"
)

// dexConfigPath points at the Dex config, relative to this package.
const dexConfigPath = "../../../../infra/dex-config.yaml"

// TestDexUsersMatchDexConfig is the reason DexUser can carry an oidc_sub at all.
//
// Dex' local connector hands the `userID` of a staticPasswords entry out as the
// OIDC `sub` claim, and both accounts use fixed IDs. If somebody rotates one of
// them in the Dex config, the seeded organizer keeps the old sub and the login
// no longer finds it - silently, because nothing else compares the two values.
// This test turns that into a failure.
func TestDexUsersMatchDexConfig(t *testing.T) {
	type dexConfig struct {
		StaticPasswords []struct {
			Email  string `yaml:"email"`
			UserID string `yaml:"userID"`
		} `yaml:"staticPasswords"`
	}

	raw, err := os.ReadFile(filepath.Clean(dexConfigPath))
	require.NoError(t, err, "the Dex config must stay readable at %s", dexConfigPath)

	var cfg dexConfig
	require.NoError(t, yaml.Unmarshal(raw, &cfg), "the Dex config must stay parsable")

	require.NotEmpty(t, cfg.StaticPasswords, "no staticPasswords found in %s", dexConfigPath)

	fromConfig := make(map[string]string, len(cfg.StaticPasswords))
	for _, entry := range cfg.StaticPasswords {
		fromConfig[entry.Email] = entry.UserID
	}

	for _, def := range DexUsers() {
		t.Run(def.Email, func(t *testing.T) {
			expected, ok := fromConfig[def.Email]
			require.Truef(t, ok,
				"%s is seeded as a Dex user but has no staticPasswords entry in %s",
				def.Email, dexConfigPath)
			require.Equalf(t, expected, def.OidcSub,
				"oidc_sub in the seeder differs from the Dex userID in %s. Dex hands "+
					"userID out as the sub claim, so a login would stop matching this "+
					"organizer.", dexConfigPath)
		})
	}
}

func TestDexUsers_HaveUniqueEmailsAndSubs(t *testing.T) {
	emails := map[string]bool{}
	subs := map[string]bool{}

	for _, def := range DexUsers() {
		require.False(t, emails[def.Email], "duplicate email %s", def.Email)
		require.False(t, subs[def.OidcSub], "duplicate oidc_sub %s", def.OidcSub)

		emails[def.Email] = true
		subs[def.OidcSub] = true
	}
}

func TestDexUser_ToOrganizerSetsSubAndEmail(t *testing.T) {
	for _, def := range DexUsers() {
		organizer := def.ToOrganizer()

		require.NotNil(t, organizer.OidcSub)
		assert.Equal(t, def.OidcSub, *organizer.OidcSub)
		assert.Equal(t, def.Email, organizer.Email)
		assert.Equal(t, def.Handle, organizer.Handle)
		assert.Equal(t, def.Languages, organizer.Languages)
	}
}

func TestSeeder_RunCreatesEverything(t *testing.T) {
	repos := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, New(repos, WithOrganizerCount(10), WithSeed(7)).Run(ctx))

	teams, teamTotal, err := repos.Teams.List(ctx, listAll(), repository.TeamListFilters{})
	require.NoError(t, err)
	require.Len(t, teams, len(teamNames))
	require.Equal(t, int64(len(teamNames)), teamTotal)

	organizers, organizerTotal, err := repos.Organizers.List(ctx, listAll(), repository.OrganizerListFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(10+len(DexUsers())), organizerTotal,
		"expected the two Dex users plus the generated organizers")
	require.Len(t, organizers, 10+len(DexUsers()))

	active, err := repos.EventYears.GetActive(ctx)
	require.NoError(t, err)
	assert.Equal(t, DefaultActiveYear, active.Year)

	previous, err := repos.EventYears.GetByYear(ctx, previousYears[len(previousYears)-1])
	require.NoError(t, err)
	assert.False(t, previous.IsActive, "there must be exactly one active year")

	coc, err := repos.CodeOfConducts.GetLatestPublished(ctx)
	require.NoError(t, err)
	assert.Equal(t, domain.CodeOfConductStatusPublished, coc.Status)
	assert.Equal(t, "1.0.0", coc.Version)
	assert.Equal(t, "Initial version", coc.ChangeNote)
	assert.NotEmpty(t, coc.Text)

	for _, def := range DexUsers() {
		t.Run("dex/"+def.Email, func(t *testing.T) {
			found, err := repos.Organizers.GetByEmail(ctx, def.Email)
			require.NoError(t, err)
			assert.Equal(t, def.Handle, found.Handle)
			require.NotNil(t, found.OidcSub)
			assert.Equal(t, def.OidcSub, *found.OidcSub)

			bySub, err := repos.Organizers.GetByOidcSub(ctx, def.OidcSub)
			require.NoError(t, err, "the login must resolve the organizer by sub")
			assert.Equal(t, found.ID, bySub.ID)
		})
	}
}

// TestSeeder_RunIsIdempotent is why `database seed` is a command of its own: a
// second run must not fail on a unique constraint.
func TestSeeder_RunIsIdempotent(t *testing.T) {
	repos := newTestStore(t)
	ctx := context.Background()

	seed := New(repos, WithOrganizerCount(8), WithSeed(3))

	require.NoError(t, seed.Run(ctx))

	_, organizersAfterFirst, err := repos.Organizers.List(ctx, listAll(), repository.OrganizerListFilters{})
	require.NoError(t, err)

	_, participationsAfterFirst, err := repos.Participations.List(
		ctx, listAll(), repository.ParticipationListFilters{},
	)
	require.NoError(t, err)

	require.NoError(t, seed.Run(ctx), "the second run must pass")

	_, organizersAfterSecond, err := repos.Organizers.List(ctx, listAll(), repository.OrganizerListFilters{})
	require.NoError(t, err)
	assert.Equal(t, organizersAfterFirst, organizersAfterSecond,
		"the second run must not duplicate organizers")

	_, participationsAfterSecond, err := repos.Participations.List(
		ctx, listAll(), repository.ParticipationListFilters{},
	)
	require.NoError(t, err)
	assert.Equal(t, participationsAfterFirst, participationsAfterSecond,
		"the second run must not duplicate participations")
}

// TestSeeder_PreservesLinkedOidcSub protects the linking path: a real login
// already produced a sub, and the seeder must not overwrite it.
func TestSeeder_PreservesLinkedOidcSub(t *testing.T) {
	repos := newTestStore(t)
	ctx := context.Background()

	linked := "9a2f5c10-0000-4000-8000-000000000001"
	organizer := &domain.Organizer{
		Email:    AdminEmail,
		Handle:   "admin",
		Realname: "Original Name",
		OidcSub:  &linked,
	}
	require.NoError(t, repos.Organizers.Create(ctx, organizer))

	require.NoError(t, New(repos, WithOrganizerCount(0)).Run(ctx))

	found, err := repos.Organizers.GetByEmail(ctx, AdminEmail)
	require.NoError(t, err)
	require.NotNil(t, found.OidcSub)
	assert.Equal(t, linked, *found.OidcSub,
		"a sub set by a real login must not be overwritten by the seed")
}

// TestSeeder_AssignsTeamsToDexUser checks the fixed team assignment, which is
// how the two Dex organizers end up in the "Infodesk" and "Netzwerk" teams.
func TestSeeder_AssignsTeamsToDexUser(t *testing.T) {
	repos := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, New(repos, WithOrganizerCount(0)).Run(ctx))

	admin, err := repos.Organizers.GetByEmail(ctx, AdminEmail)
	require.NoError(t, err)

	participation, err := repos.Participations.GetByOrganizerAndYear(ctx, admin.ID, DefaultActiveYear)
	require.NoError(t, err)
	require.NotEmpty(t, participation.Teams)

	names := make([]string, 0, len(participation.Teams))
	for _, team := range participation.Teams {
		names = append(names, team.Name)
	}

	assert.ElementsMatch(t, []string{"Infodesk", "Kasse"}, names)
}

// TestSeeder_BlockNumbersStayInWindow guards the shift grid against values
// the frontend could not map back to a time.
func TestSeeder_BlockNumbersStayInWindow(t *testing.T) {
	repos := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, New(repos, WithOrganizerCount(15), WithSeed(11)).Run(ctx))

	participations, total, err := repos.Participations.List(
		ctx, listAll(), repository.ParticipationListFilters{},
	)
	require.NoError(t, err)
	require.Positive(t, total)

	totalBlocks := partyDays * shiftBlocksPerDay

	for _, participation := range participations {
		assert.NotEqual(t, domain.ParticipationStatus(""), participation.Status)
		assert.Contains(t, shirtSizes, participation.TShirtSize)

		for _, block := range participation.ShiftPreferredTimes {
			assert.GreaterOrEqual(t, block, 0, "negative block number")
			assert.Less(t, block, totalBlocks, "block outside the party window")
		}

		for _, block := range participation.ShiftUnavailableTimes {
			assert.GreaterOrEqual(t, block, 0)
			assert.Less(t, block, totalBlocks)
		}
	}
}

func TestBuildShiftTimes_ReturnsUniqueBlocks(t *testing.T) {
	for range 100 {
		blocks := buildShiftTimes()
		require.NotEmpty(t, blocks)

		seen := map[int]bool{}
		for _, block := range blocks {
			require.False(t, seen[block], "duplicate block %d", block)
			seen[block] = true
		}
	}
}

func TestPickTeams_NamedAndRandomSelection(t *testing.T) {
	teams := []domain.Team{{Name: "Infodesk"}, {Name: "Kasse"}, {Name: "Netzwerk"}}

	named := pickTeams(teams, []string{"Infodesk", "Kasse"}, false)
	assert.Equal(t, []domain.Team{teams[0], teams[1]}, named)

	assert.Empty(t, pickTeams(teams, []string{"DoesNotExist"}, false),
		"unknown team names are skipped")

	assert.Empty(t, pickTeams(teams, nil, false),
		"Dex users without a fixed team list get no random selection")

	random := pickTeams(teams, nil, true)
	require.NotEmpty(t, random)

	for _, team := range random {
		assert.Contains(t, teams, team)
	}
}

func TestDexTeamsFor_OnlyDexUsers(t *testing.T) {
	names, ok := dexTeamsFor(AdminEmail)
	assert.True(t, ok)
	assert.NotEmpty(t, names)

	_, ok = dexTeamsFor("random@example.com")
	assert.False(t, ok)
}

// newTestStore returns a migrated in-memory store plus its repository set.
func newTestStore(t *testing.T) Repositories {
	t.Helper()

	store, err := gorm.NewSqliteInMemoryStore()
	require.NoError(t, err, "store should open")

	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	// The in-memory store uses a shared cache, so every test starts from scratch.
	require.NoError(t, store.PurgeAll(), "store should migrate cleanly")

	return Repositories{
		Organizers:     store.NewOrganizerRepository(),
		Teams:          store.NewTeamRepository(),
		EventYears:     store.NewEventYearRepository(),
		Participations: store.NewParticipationRepository(),
		CodeOfConducts: store.NewCodeOfConductRepository(),
	}
}

// listAll disables sorting surprises by using the default column and no limit.
