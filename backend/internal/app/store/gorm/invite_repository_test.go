package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

func newInviteFixture(t *testing.T, store *Store, mutate func(i *domain.Invite)) *domain.Invite {
	t.Helper()

	organizer := newOrganizerForParticipation(t, store)

	invite := &domain.Invite{
		Email:       fmtUnique("invitee") + "@example.com",
		Status:      domain.InviteStatusSent,
		Token:       domain.NewInviteToken(),
		ExpiresAt:   time.Now().UTC().Add(14 * 24 * time.Hour),
		InvitedByID: organizer.ID,
	}

	if mutate != nil {
		mutate(invite)
	}

	return invite
}

// TestInviteRepository_GetByToken covers the magic link: the token is sent as
// part of the URL and resolved again when the link is opened.
func TestInviteRepository_GetByToken(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewInviteRepository()
	ctx := context.Background()

	token := domain.NewInviteToken()
	require.Len(t, token, 36, "expected a UUIDv4")

	invite := &domain.Invite{
		Email:       fmtUnique("invitee") + "@example.com",
		Status:      domain.InviteStatusSent,
		Token:       token,
		ExpiresAt:   time.Now().UTC().Add(time.Hour),
		InvitedByID: uuid.New(),
	}
	require.NoError(t, repo.Create(ctx, invite))

	var stored string
	require.NoError(t, store.db.Raw(
		"SELECT token FROM invites WHERE id = ?", invite.ID,
	).Scan(&stored).Error)
	require.Equal(t, token, stored)

	found, err := repo.GetByToken(ctx, token)
	require.NoError(t, err)
	require.Equal(t, invite.ID, found.ID)
	require.Equal(t, token, found.Token)
}

func TestInviteRepository_GetByTokenNotFound(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewInviteRepository()

	_, err := repo.GetByToken(context.Background(), domain.NewInviteToken())
	require.True(t, IsNotFound(err))
}

func TestInviteRepository_SubmitForReview(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewInviteRepository()
	workflowRepo := store.NewInviteWorkflowRepository()
	ctx := context.Background()

	invite := newInviteFixture(t, store, nil)
	require.NoError(t, repo.Create(ctx, invite))

	require.NoError(t, workflowRepo.SubmitForReview(ctx, invite.ID, "hier notiert", true))

	loaded, err := repo.GetByID(ctx, invite.ID)
	require.NoError(t, err)
	require.Equal(t, domain.InviteStatusPendingReview, loaded.Status)
	require.Equal(t, "hier notiert", loaded.Notes)
	require.True(t, loaded.CoCAccepted)
	require.NotNil(t, loaded.CoCAcceptedAt)
}

func TestInviteRepository_Review(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewInviteRepository()
	workflowRepo := store.NewInviteWorkflowRepository()
	ctx := context.Background()

	invite := newInviteFixture(t, store, nil)
	require.NoError(t, repo.Create(ctx, invite))
	require.NoError(t, workflowRepo.SubmitForReview(ctx, invite.ID, "", true))

	reviewer := newOrganizerForParticipation(t, store)
	require.NoError(t, workflowRepo.Review(ctx, invite.ID, domain.InviteStatusApproved, reviewer.ID))

	loaded, err := repo.GetByID(ctx, invite.ID)
	require.NoError(t, err)
	require.Equal(t, domain.InviteStatusApproved, loaded.Status)
	require.NotNil(t, loaded.ReviewedByID)
	require.Equal(t, reviewer.ID, *loaded.ReviewedByID)
}

func TestInviteRepository_EmailExists(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewInviteRepository()
	workflowRepo := store.NewInviteWorkflowRepository()
	organizerRepo := store.NewOrganizerRepository()
	ctx := context.Background()

	organizer := newOrganizerFixture(t, nil)
	require.NoError(t, organizerRepo.Create(ctx, organizer))

	exists, err := workflowRepo.EmailExists(ctx, organizer.Email)
	require.NoError(t, err)
	require.True(t, exists, "email from the crew directory must count as taken")

	invite := newInviteFixture(t, store, nil)
	require.NoError(t, repo.Create(ctx, invite))

	exists, err = workflowRepo.EmailExists(ctx, invite.Email)
	require.NoError(t, err)
	require.True(t, exists, "email from an open invitation must count as taken")

	exists, err = workflowRepo.EmailExists(ctx, "niemand@example.com")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestInviteRepository_PurgeExpired(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewInviteRepository()
	workflowRepo := store.NewInviteWorkflowRepository()
	ctx := context.Background()

	past := time.Now().UTC().Add(-48 * time.Hour)
	future := time.Now().UTC().Add(48 * time.Hour)

	expired := newInviteFixture(t, store, func(i *domain.Invite) { i.ExpiresAt = past })
	require.NoError(t, repo.Create(ctx, expired))

	valid := newInviteFixture(t, store, func(i *domain.Invite) { i.ExpiresAt = future })
	require.NoError(t, repo.Create(ctx, valid))

	deleted, err := workflowRepo.PurgeExpired(ctx, time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, int64(1), deleted)

	_, err = repo.GetByID(ctx, expired.ID)
	require.True(t, IsNotFound(err))

	_, err = repo.GetByID(ctx, valid.ID)
	require.NoError(t, err)
}

func TestInviteRepository_PurgeExpiredBehaeltEingereichte(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewInviteRepository()
	workflowRepo := store.NewInviteWorkflowRepository()
	ctx := context.Background()

	past := time.Now().UTC().Add(-48 * time.Hour)
	submitted := newInviteFixture(t, store, func(i *domain.Invite) { i.ExpiresAt = past })
	require.NoError(t, repo.Create(ctx, submitted))
	require.NoError(t, workflowRepo.SubmitForReview(ctx, submitted.ID, "", true))

	deleted, err := workflowRepo.PurgeExpired(ctx, time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, int64(0), deleted, "eingereichte Einladungen bleiben erhalten")

	_, err = repo.GetByID(ctx, submitted.ID)
	require.NoError(t, err)
}

func TestInvite_IsExpired(t *testing.T) {
	now := time.Now().UTC()

	require.True(t, domain.Invite{ExpiresAt: now.Add(-time.Hour)}.IsExpired(now))
	require.False(t, domain.Invite{ExpiresAt: now.Add(time.Hour)}.IsExpired(now))
	require.False(t, domain.Invite{}.IsExpired(now), "ohne Ablaufdatum nie abgelaufen")
}

func TestInviteRepository_UniqueEmail(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewInviteRepository()
	ctx := context.Background()

	first := newInviteFixture(t, store, nil)
	require.NoError(t, repo.Create(ctx, first))

	second := newInviteFixture(t, store, func(i *domain.Invite) { i.Email = first.Email })
	require.Error(t, repo.Create(ctx, second), "duplicate email must fail")
}

func TestInviteRepository_SortWhitelist(t *testing.T) {
	store := newTestStore(t)
	repo := store.NewInviteRepository()

	_, _, err := repo.List(context.Background(),
		repository.InviteListParams{Limit: 10, Sort: "token_hash"},
		repository.InviteListFilters{})
	require.Error(t, err, "token_hash is not in the whitelist")
}
