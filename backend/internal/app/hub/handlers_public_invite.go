package hub

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/mailer"
)

func (s *Server) GetPublicInvite(c *gin.Context) {
	token := c.Param("token")

	invite, err := s.inviteRepo.GetByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})

		return
	}

	if invite.IsExpired(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "invite expired"})

		return
	}

	if invite.Status != domain.InviteStatusSent {
		c.JSON(http.StatusNotFound, gin.H{"error": "invite not available"})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":      invite.Email,
		"handle":     invite.Handle,
		"reason":     invite.Reason,
		"event_name": s.eventName(),
		"expires_at": invite.ExpiresAt,
	})
}

func (s *Server) SubmitPublicInvite(c *gin.Context) {
	token := c.Param("token")

	invite, err := s.inviteRepo.GetByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})

		return
	}

	if invite.IsExpired(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "invite expired"})

		return
	}

	if invite.Status != domain.InviteStatusSent {
		c.JSON(http.StatusConflict, gin.H{"error": "invite not available for submission"})

		return
	}

	var payload struct {
		ProvidedName  string `json:"provided_name"`
		ProvidedPhone string `json:"provided_phone"`
		Notes         string `json:"notes"`
		CoCAccepted   bool   `json:"coc_accepted"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if !payload.CoCAccepted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code of conduct must be accepted"})

		return
	}

	if err := s.inviteWorkflowRepo.SubmitForReview(
		c.Request.Context(),
		invite.ID,
		payload.ProvidedName,
		payload.ProvidedPhone,
		payload.Notes,
		payload.CoCAccepted,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit invite"})

		return
	}

	go func() {
		data := mailer.InviteSubmittedData{EventName: s.eventName()}
		if err := s.mailer.SendInviteSubmitted(invite.Email, data); err != nil {
			slog.Warn("failed to send invite submitted confirmation", "invite_id", invite.ID, "error", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "pending_review"})
}

func (s *Server) eventName() string {
	if s.cfg.App.EventName != "" {
		return s.cfg.App.EventName
	}

	return "Evoke"
}
