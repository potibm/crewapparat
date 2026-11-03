package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

func TestCodeOfConductAcceptanceRepository(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	orgRepo := store.NewOrganizerRepository()
	cocRepo := store.NewCodeOfConductRepository()
	acceptanceRepo := store.NewCodeOfConductAcceptanceRepository()

	organizer := &domain.Organizer{Handle: "tester", Realname: "Test Organizer", Email: "tester@example.com"}
	require.NoError(t, orgRepo.Create(ctx, organizer))

	publishedAt := time.Now().Add(-time.Hour)
	coc := &domain.CodeOfConduct{
		Version:     "1.0.0",
		Status:      domain.CodeOfConductStatusPublished,
		Text:        "# CoC",
		PublishedAt: &publishedAt,
	}
	require.NoError(t, cocRepo.Create(ctx, coc))

	acceptedAt := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	acceptance := &domain.CodeOfConductAcceptance{
		OrganizerID:     organizer.ID,
		CodeOfConductID: coc.ID,
		AcceptedAt:      acceptedAt,
	}
	require.NoError(t, acceptanceRepo.Create(ctx, acceptance))
	require.NotEmpty(t, acceptance.ID)

	t.Run("GetByOrganizerAndCodeOfConduct", func(t *testing.T) {
		loaded, err := acceptanceRepo.GetByOrganizerAndCodeOfConduct(ctx, organizer.ID, coc.ID)
		require.NoError(t, err)
		assert.True(t, acceptedAt.Equal(loaded.AcceptedAt))
	})

	t.Run("duplicate acceptance violates the unique constraint", func(t *testing.T) {
		duplicate := &domain.CodeOfConductAcceptance{
			OrganizerID:     organizer.ID,
			CodeOfConductID: coc.ID,
			AcceptedAt:      time.Now(),
		}

		require.Error(t, acceptanceRepo.Create(ctx, duplicate))
	})

	t.Run("ListByOrganizerID", func(t *testing.T) {
		items, err := acceptanceRepo.ListByOrganizerID(ctx, organizer.ID)
		require.NoError(t, err)
		require.Len(t, items, 1)
	})

	t.Run("List includes organizer data and filters by version", func(t *testing.T) {
		versionID := coc.ID

		items, total, err := acceptanceRepo.List(
			ctx,
			repository.CodeOfConductAcceptanceListParams{
				Sort:  "accepted",
				Order: "DESC",
			},
			repository.CodeOfConductAcceptanceListFilters{CodeOfConductID: &versionID},
		)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Len(t, items, 1)
		assert.Equal(t, "tester", items[0].Handle)
		assert.Equal(t, "Test Organizer", items[0].Realname)
	})

	t.Run("CountAcceptances", func(t *testing.T) {
		count, err := cocRepo.CountAcceptances(ctx, coc.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}
