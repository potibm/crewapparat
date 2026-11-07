package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/config"
	"github.com/potibm/crewapparat/internal/app/domain"
	store "github.com/potibm/crewapparat/internal/app/store/gorm"
)

func TestUpdateInvite_ApproveCreatesOrganizer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st, err := store.NewSqliteInMemoryStore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	require.NoError(t, st.PurgeAll())

	organizerRepo := st.NewOrganizerRepository()
	inviteRepo := st.NewInviteRepository()
	workflowRepo := st.NewInviteWorkflowRepository()
	eventYearRepo := st.NewEventYearRepository()
	participationRepo := st.NewParticipationRepository()

	inviter := &domain.Organizer{Handle: "inviter", Email: "inviter@example.com", Slug: "inviter"}
	require.NoError(t, organizerRepo.Create(context.Background(), inviter))

	require.NoError(t, eventYearRepo.Create(context.Background(), &domain.EventYear{Year: "2026", IsActive: true}))

	invite := &domain.Invite{
		Email:       "newbie@example.com",
		Status:      domain.InviteStatusPendingReview,
		Handle:      "newbie",
		Token:       domain.NewInviteToken(),
		ExpiresAt:   time.Now().UTC().Add(14 * 24 * time.Hour),
		InvitedByID: inviter.ID,
		CoCAccepted: true,
	}
	require.NoError(t, inviteRepo.Create(context.Background(), invite))

	server, err := NewServer(Config{
		Cfg:                config.Config{Auth: &config.AuthConfig{ProxyAdmins: []string{"tester"}}},
		OrganizerRepo:      organizerRepo,
		InviteRepo:         inviteRepo,
		InviteWorkflowRepo: workflowRepo,
		EventYearRepo:      eventYearRepo,
		ParticipationRepo:  participationRepo,
	})
	require.NoError(t, err)

	r := gin.New()
	admin := r.Group("/api/admin")
	admin.Use(func(c *gin.Context) {
		c.Set("userID", "tester")
		c.Next()
	})
	admin.PUT("/invites/:id", server.UpdateInvite)

	body, err := json.Marshal(map[string]any{
		"email":  invite.Email,
		"handle": invite.Handle,
		"status": "approved",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPut, "/api/admin/invites/"+invite.ID.String(), bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	loaded, err := inviteRepo.GetByID(context.Background(), invite.ID)
	require.NoError(t, err)
	require.Equal(t, domain.InviteStatusApproved, loaded.Status)

	org, err := organizerRepo.GetByEmail(context.Background(), invite.Email)
	require.NoError(t, err)
	require.Equal(t, invite.Handle, org.Handle)

	part, err := participationRepo.GetByOrganizerAndYear(context.Background(), org.ID, "2026")
	require.NoError(t, err)
	require.Equal(t, domain.ParticipationStatusConfirmed, part.Status)
}

func TestUpdateInvite_ApproveIsIdempotent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st, err := store.NewSqliteInMemoryStore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	require.NoError(t, st.PurgeAll())

	organizerRepo := st.NewOrganizerRepository()
	inviteRepo := st.NewInviteRepository()
	workflowRepo := st.NewInviteWorkflowRepository()
	eventYearRepo := st.NewEventYearRepository()
	participationRepo := st.NewParticipationRepository()

	inviter := &domain.Organizer{Handle: "inviter2", Email: "inviter2@example.com", Slug: "inviter2"}
	require.NoError(t, organizerRepo.Create(context.Background(), inviter))

	invite := &domain.Invite{
		Email:       "dup@example.com",
		Status:      domain.InviteStatusPendingReview,
		Handle:      "dup",
		Token:       domain.NewInviteToken(),
		ExpiresAt:   time.Now().UTC().Add(14 * 24 * time.Hour),
		InvitedByID: inviter.ID,
	}
	require.NoError(t, inviteRepo.Create(context.Background(), invite))

	// Pre-existing organizer with same email.
	require.NoError(t, organizerRepo.Create(context.Background(), &domain.Organizer{
		Handle: "existing", Email: invite.Email, Slug: "existing",
	}))

	server, err := NewServer(Config{
		Cfg:                config.Config{Auth: &config.AuthConfig{ProxyAdmins: []string{"tester"}}},
		OrganizerRepo:      organizerRepo,
		InviteRepo:         inviteRepo,
		InviteWorkflowRepo: workflowRepo,
		EventYearRepo:      eventYearRepo,
		ParticipationRepo:  participationRepo,
	})
	require.NoError(t, err)

	r := gin.New()
	admin := r.Group("/api/admin")
	admin.Use(func(c *gin.Context) {
		c.Set("userID", "tester")
		c.Next()
	})
	admin.PUT("/invites/:id", server.UpdateInvite)

	body, _ := json.Marshal(map[string]any{"status": "approved", "handle": invite.Handle, "email": invite.Email})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/admin/invites/"+invite.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// Idempotent: the pre-existing organizer is reused, not duplicated.
	org, err := organizerRepo.GetByEmail(context.Background(), invite.Email)
	require.NoError(t, err)
	require.Equal(t, "existing", org.Handle)
}
