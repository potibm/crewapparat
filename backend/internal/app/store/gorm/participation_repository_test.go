package gorm

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

func newOrganizerForParticipation(t *testing.T, store *Store) *domain.Organizer {
	t.Helper()

	o := newOrganizerFixture(t, nil)
	require.NoError(t, store.NewOrganizerRepository().Create(context.Background(), o))

	return o
}

func TestParticipationRepository_Roundtrip(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewParticipationRepository()
	ctx := context.Background()

	organizer := newOrganizerForParticipation(t, store)
	in := &domain.Participation{
		OrganizerID:           organizer.ID,
		EventYearID:           "2026",
		Status:                domain.ParticipationStatusConfirmed,
		TShirtSize:            "XL",
		AdditionalPaidShirt:   true,
		PublicWebsiteConsent:  true,
		ShiftPreferredTimes:   []int{0, 1, 2, 32, 33},
		ShiftUnavailableTimes: []int{5, 6},
		ShiftPreferredPartner: []string{"uuid-a", "uuid-b"},
	}

	require.NoError(t, repo.Create(ctx, in))
	require.NotEqual(t, uuid.Nil, in.ID)

	loaded, err := repo.GetByID(ctx, in.ID)
	require.NoError(t, err)
	require.Equal(t, organizer.ID, loaded.OrganizerID)
	require.Equal(t, "2026", loaded.EventYearID)
	require.Equal(t, domain.ParticipationStatusConfirmed, loaded.Status)
	require.Equal(t, "XL", loaded.TShirtSize)
	require.True(t, loaded.AdditionalPaidShirt)
	require.True(t, loaded.PublicWebsiteConsent)
	require.Equal(t, []int{0, 1, 2, 32, 33}, loaded.ShiftPreferredTimes)
	require.Equal(t, []int{5, 6}, loaded.ShiftUnavailableTimes)
	require.Equal(t, []string{"uuid-a", "uuid-b"}, loaded.ShiftPreferredPartner)
}

func TestParticipationRepository_UniquePerOrganizerAndYear(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewParticipationRepository()
	ctx := context.Background()

	organizer := newOrganizerForParticipation(t, store)
	base := &domain.Participation{
		OrganizerID: organizer.ID,
		EventYearID: "2026",
		Status:      domain.ParticipationStatusPending,
	}
	require.NoError(t, repo.Create(ctx, base))

	duplicate := &domain.Participation{
		OrganizerID: organizer.ID,
		EventYearID: "2026",
		Status:      domain.ParticipationStatusConfirmed,
	}
	require.Error(t, repo.Create(ctx, duplicate),
		"two participations for the same organizer-year must fail")

	otherYear := &domain.Participation{
		OrganizerID: organizer.ID,
		EventYearID: "2027",
		Status:      domain.ParticipationStatusPending,
	}
	require.NoError(t, repo.Create(ctx, otherYear), "anderes Jahr ist erlaubt")
}

func TestParticipationRepository_GetByOrganizerAndYear(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewParticipationRepository()
	ctx := context.Background()

	organizer := newOrganizerForParticipation(t, store)
	in := &domain.Participation{
		OrganizerID: organizer.ID,
		EventYearID: "2026",
		Status:      domain.ParticipationStatusConfirmed,
	}
	require.NoError(t, repo.Create(ctx, in))

	found, err := repo.GetByOrganizerAndYear(ctx, organizer.ID, "2026")
	require.NoError(t, err)
	require.Equal(t, in.ID, found.ID)
}

func TestParticipationRepository_SetTeams(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewParticipationRepository()
	teamRepo := store.NewTeamRepository()
	ctx := context.Background()

	organizer := newOrganizerForParticipation(t, store)
	p := &domain.Participation{OrganizerID: organizer.ID, EventYearID: "2026"}
	require.NoError(t, repo.Create(ctx, p))

	infodesk := &domain.Team{Name: "Infodesk"}
	network := &domain.Team{Name: "Network"}

	require.NoError(t, teamRepo.Create(ctx, infodesk))
	require.NoError(t, teamRepo.Create(ctx, network))

	require.NoError(t, repo.SetTeams(ctx, p.ID, []domain.Team{*infodesk, *network}))

	loaded, err := repo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	require.Len(t, loaded.Teams, 2)

	// Replace rather than append: the old join rows must disappear.
	require.NoError(t, repo.SetTeams(ctx, p.ID, []domain.Team{*network}))

	loaded, err = repo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	require.Len(t, loaded.Teams, 1)
	require.Equal(t, network.ID, loaded.Teams[0].ID)

	var joinRows int64
	require.NoError(t, store.db.Table("participation_teams").Count(&joinRows).Error)
	require.Equal(t, int64(1), joinRows, "join table must have exactly one row")
}

// TestParticipationRepository_SetTeamsLeavesNoOrphans covers the case of a
// soft-deleted team, where the join row survives as an orphan.
func TestParticipationRepository_SetTeamsLeavesNoOrphans(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewParticipationRepository()
	teamRepo := store.NewTeamRepository()
	ctx := context.Background()

	organizer := newOrganizerForParticipation(t, store)
	p := &domain.Participation{OrganizerID: organizer.ID, EventYearID: "2026"}
	require.NoError(t, repo.Create(ctx, p))

	team := &domain.Team{Name: "Infodesk"}
	require.NoError(t, teamRepo.Create(ctx, team))
	require.NoError(t, repo.SetTeams(ctx, p.ID, []domain.Team{*team}))

	require.NoError(t, teamRepo.Delete(ctx, team.ID))

	loaded, err := repo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	require.Empty(t, loaded.Teams, "soft-deleted team must not be loaded")

	var orphans int64
	require.NoError(t, store.db.Table("participation_teams").Count(&orphans).Error)
	require.Equal(t, int64(1), orphans,
		"join row stays in place - SetTeams clears it explicitly")
}

func TestParticipationRepository_ListFilter(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewParticipationRepository()
	ctx := context.Background()

	organizer := newOrganizerForParticipation(t, store)
	status := domain.ParticipationStatusConfirmed

	require.NoError(t, repo.Create(ctx, &domain.Participation{
		OrganizerID: organizer.ID, EventYearID: "2026",
		Status: domain.ParticipationStatusConfirmed, TShirtSize: "L",
	}))
	require.NoError(t, repo.Create(ctx, &domain.Participation{
		OrganizerID: organizer.ID, EventYearID: "2027",
		Status: domain.ParticipationStatusDeclined,
	}))

	list, total, err := repo.List(ctx,
		repository.ParticipationListParams{Limit: 10, Sort: "id", Order: "ASC"},
		repository.ParticipationListFilters{OrganizerID: &organizer.ID, Status: &status})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Equal(t, domain.ParticipationStatusConfirmed, list[0].Status)
}

func TestParticipationRepository_ListFilterByTeam(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewParticipationRepository()
	teamRepo := store.NewTeamRepository()
	ctx := context.Background()

	infodesk := &domain.Team{Name: "Infodesk", HasShifts: true}
	require.NoError(t, teamRepo.Create(ctx, infodesk))

	network := &domain.Team{Name: "Network"}
	require.NoError(t, teamRepo.Create(ctx, network))

	organizer := newOrganizerForParticipation(t, store)

	inInfodesk := &domain.Participation{
		OrganizerID: organizer.ID, EventYearID: "2026", Teams: []domain.Team{*infodesk},
	}
	require.NoError(t, repo.Create(ctx, inInfodesk))

	otherOrganizer := newOrganizerForParticipation(t, store)

	inNetwork := &domain.Participation{
		OrganizerID: otherOrganizer.ID, EventYearID: "2026", Teams: []domain.Team{*network},
	}
	require.NoError(t, repo.Create(ctx, inNetwork))

	teamName := "Infodesk"
	year := "2026"

	list, total, err := repo.List(ctx,
		repository.ParticipationListParams{Limit: 10, Sort: "id", Order: "ASC"},
		repository.ParticipationListFilters{TeamName: &teamName, EventYearID: &year})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Equal(t, organizer.ID, list[0].OrganizerID)
	require.Len(t, list[0].Teams, 1)
	require.True(t, list[0].Teams[0].HasShifts)
}

func TestParticipationRepository_SortWhitelist(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewParticipationRepository()

	_, _, err := repo.List(context.Background(),
		repository.ParticipationListParams{Limit: 10, Sort: "organizer_id"},
		repository.ParticipationListFilters{})
	require.Error(t, err, "organizer_id is not in the whitelist")
}

func TestParticipationRepository_DeleteClearsJoinTable(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewParticipationRepository()
	teamRepo := store.NewTeamRepository()
	ctx := context.Background()

	organizer := newOrganizerForParticipation(t, store)
	p := &domain.Participation{OrganizerID: organizer.ID, EventYearID: "2026"}
	require.NoError(t, repo.Create(ctx, p))

	team := &domain.Team{Name: "Infodesk"}
	require.NoError(t, teamRepo.Create(ctx, team))
	require.NoError(t, repo.SetTeams(ctx, p.ID, []domain.Team{*team}))

	require.NoError(t, repo.Delete(ctx, p.ID))

	var orphans int64
	require.NoError(t, store.db.Table("participation_teams").Count(&orphans).Error)
	require.Equal(t, int64(0), orphans, "delete must clean up the join rows")
}
