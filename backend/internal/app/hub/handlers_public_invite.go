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

	cocText := ""
	cocVersion := ""

	if s.codeOfConductRepo != nil {
		if coc, err := s.codeOfConductRepo.GetLatestPublished(c.Request.Context()); err == nil && coc != nil {
			cocText = coc.Text
			cocVersion = coc.Version
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"email":                   invite.Email,
		"handle":                  invite.Handle,
		"reason":                  invite.Reason,
		"event_name":              s.eventName(),
		"expires_at":              invite.ExpiresAt,
		"code_of_conduct":         cocText,
		"code_of_conduct_version": cocVersion,
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
		Notes       string `json:"notes"`
		CoCAccepted bool   `json:"coc_accepted"`
		CoCVersion  string `json:"coc_version"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if !payload.CoCAccepted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code of conduct must be accepted"})

		return
	}

	cocVersion := payload.CoCVersion
	if cocVersion == "" && s.codeOfConductRepo != nil {
		if coc, err := s.codeOfConductRepo.GetLatestPublished(c.Request.Context()); err == nil && coc != nil {
			cocVersion = coc.Version
		}
	}

	if err := s.inviteWorkflowRepo.SubmitForReview(
		c.Request.Context(),
		invite.ID,
		payload.Notes,
		payload.CoCAccepted,
		cocVersion,
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
