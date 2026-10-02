package gorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

func TestTeamRepository_Roundtrip(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewTeamRepository()
	ctx := context.Background()

	team := &domain.Team{Name: "Infodesk"}
	require.NoError(t, repo.Create(ctx, team))
	require.NotEmpty(t, team.ID)

	loaded, err := repo.GetByID(ctx, team.ID)
	require.NoError(t, err)
	require.Equal(t, "Infodesk", loaded.Name)

	byName, err := repo.GetByName(ctx, "Infodesk")
	require.NoError(t, err)
	require.Equal(t, team.ID, byName.ID)
}

func TestTeamRepository_UniqueName(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewTeamRepository()
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &domain.Team{Name: "Infodesk"}))
	require.Error(t, repo.Create(ctx, &domain.Team{Name: "Infodesk"}),
		"duplicate team name must fail")
}

func TestTeamRepository_UpdateUndDelete(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewTeamRepository()
	ctx := context.Background()

	team := &domain.Team{Name: "Infodesk"}
	require.NoError(t, repo.Create(ctx, team))

	team.Name = "Network"
	require.NoError(t, repo.Update(ctx, team))

	loaded, err := repo.GetByID(ctx, team.ID)
	require.NoError(t, err)
	require.Equal(t, "Network", loaded.Name)

	require.NoError(t, repo.Delete(ctx, team.ID))

	_, err = repo.GetByID(ctx, team.ID)
	require.True(t, IsNotFound(err))
}

func TestEventYearRepository_AnlegenUndAktivieren(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewEventYearRepository()
	ctx := context.Background()

	for _, year := range []string{"2025", "2026", "2027"} {
		require.NoError(t, repo.Create(ctx, &domain.EventYear{Year: year}))
	}

	active, err := repo.GetActive(ctx)
	require.True(t, IsNotFound(err), "no year may be active before ActivateYear")
	require.Nil(t, active)

	require.NoError(t, repo.ActivateYear(ctx, "2026"))

	active, err = repo.GetActive(ctx)
	require.NoError(t, err)
	require.Equal(t, "2026", active.Year)
}

// TestEventYearRepository_NurEinAktivesJahr guarantees uniqueness, because
// SQLite has no partial unique indexes.
func TestEventYearRepository_NurEinAktivesJahr(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewEventYearRepository()
	ctx := context.Background()

	for _, year := range []string{"2025", "2026"} {
		require.NoError(t, repo.Create(ctx, &domain.EventYear{Year: year}))
	}

	require.NoError(t, repo.ActivateYear(ctx, "2025"))
	require.NoError(t, repo.ActivateYear(ctx, "2026"))

	years, total, err := repo.List(ctx,
		repository.EventYearListParams{Limit: 10, Sort: "year", Order: "ASC"},
		repository.EventYearListFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)

	var activeCount int

	for _, y := range years {
		if y.IsActive {
			activeCount++
		}
	}

	require.Equal(t, 1, activeCount, "exactly one year must be active")
}

func TestEventYearRepository_ActivateYearUnbekannt(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewEventYearRepository()

	err := repo.ActivateYear(context.Background(), "2099")
	require.True(t, IsNotFound(err), "unknown year must not be activatable")
}

func TestEventYearRepository_Delete(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewEventYearRepository()
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &domain.EventYear{Year: "2026"}))
	require.NoError(t, repo.Delete(ctx, "2026"))

	_, err := repo.GetByYear(ctx, "2026")
	require.True(t, IsNotFound(err))
}
