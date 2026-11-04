package hub

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/mailer"
	"github.com/potibm/crewapparat/internal/app/repository"
)

func (s *Server) ListInvites(c *gin.Context) {
	params := parseInviteListParams(c)
	filters := parseInviteFilters(c)

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		organizer, err := s.organizerRepo.GetByOidcSub(c.Request.Context(), uid)
		if err == nil && organizer != nil {
			id := organizer.ID
			filters.InvitedByID = &id
		}
	}

	items, total, err := s.inviteRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list invites: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

func (s *Server) GetInvite(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	invite, err := s.inviteRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, inviteNotFound)

			return
		}

		respondWithInternalServerProblem(c, "Failed to get invite: "+err.Error())

		return
	}

	c.JSON(http.StatusOK, invite)
}

func (s *Server) CreateInvite(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)

	organizer, err := s.organizerRepo.GetByOidcSub(c.Request.Context(), uid)
	if err != nil {
		respondWithForbiddenProblem(c, "Only linked organizers can create invites")

		return
	}

	var invite domain.Invite
	if err := c.ShouldBindJSON(&invite); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	if strings.TrimSpace(invite.Email) == "" {
		respondWithBadRequestProblem(c, "email is required")

		return
	}

	if strings.TrimSpace(invite.Handle) == "" {
		respondWithBadRequestProblem(c, "handle is required")

		return
	}

	invite.InvitedByID = organizer.ID

	// These fields are filled in by the invitee through the magic link, never
	// by the inviting organizer.
	invite.ProvidedName = ""
	invite.ProvidedPhone = ""
	invite.Notes = ""
	invite.CoCAccepted = false
	invite.CoCAcceptedAt = nil
	invite.ReviewNote = ""

	isAdmin := s.isAdmin(c, uid)
	if !isAdmin {
		invite.Status = domain.InviteStatusPendingApproval
	} else if invite.Status == "" {
		invite.Status = domain.InviteStatusSent
	}

	if err := s.inviteRepo.Create(c.Request.Context(), &invite); err != nil {
		respondWithInternalServerProblem(c, "Failed to create invite: "+err.Error())

		return
	}

	if !isAdmin {
		go func() {
			link := s.mailer.FrontendBaseURL() + "/admin/#/invites/" + invite.ID.String()

			data := mailer.InvitePendingApprovalData{
				Email:         invite.Email,
				InvitedByName: organizer.Realname,
				Link:          link,
			}
			for _, recipient := range s.adminRecipients {
				if err := s.mailer.SendInvitePendingApproval(recipient, data); err != nil {
					slog.Warn("failed to send invite approval email",
						"invite_id", invite.ID,
						"recipient", recipient,
						"error", err,
					)
				}
			}
		}()
	}

	slog.Info("Create Invite: Successfully created invite", "id", invite.ID)
	c.JSON(http.StatusCreated, invite)
}

func (s *Server) UpdateInvite(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	isAdmin := s.isAdmin(c, uid)

	existing, err := s.inviteRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, inviteNotFound)

			return
		}

		respondWithInternalServerProblem(c, "Failed to get invite: "+err.Error())

		return
	}

	if !isAdmin {
		organizer, oerr := s.organizerRepo.GetByOidcSub(c.Request.Context(), uid)
		if oerr != nil || organizer == nil || existing.InvitedByID != organizer.ID ||
			existing.Status != domain.InviteStatusPendingApproval {
			respondWithForbiddenProblem(c, "You can only update your own pending invites")

			return
		}
	}

	var invite domain.Invite
	if err := c.ShouldBindJSON(&invite); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	invite.ID = id

	// Preserve server-managed fields that are not part of the edit form.
	invite.InvitedByID = existing.InvitedByID
	invite.Token = existing.Token
	invite.ExpiresAt = existing.ExpiresAt

	// Preserve the fields submitted by the invitee through the magic link.
	invite.ProvidedName = existing.ProvidedName
	invite.ProvidedPhone = existing.ProvidedPhone
	invite.Notes = existing.Notes
	invite.CoCAccepted = existing.CoCAccepted
	invite.CoCAcceptedAt = existing.CoCAcceptedAt

	if !isAdmin {
		// Non-admins cannot change the status away from pending_approval.
		invite.Status = domain.InviteStatusPendingApproval
		invite.ReviewNote = existing.ReviewNote
	} else {
		userID, _ := c.Get("userID")

		uid, _ := userID.(string)
		if organizer, err := s.organizerRepo.GetByOidcSub(c.Request.Context(), uid); err == nil && organizer != nil {
			reviewer := organizer.ID
			invite.ReviewedByID = &reviewer
		}
	}

	if err := s.inviteRepo.Update(c.Request.Context(), &invite); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, inviteNotFound)

			return
		}

		respondWithInternalServerProblem(c, "Failed to update invite: "+err.Error())

		return
	}

	slog.Info("Update Invite: Successfully updated invite", "id", invite.ID)
	c.JSON(http.StatusOK, invite)
}

func (s *Server) DeleteInvite(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		existing, gerr := s.inviteRepo.GetByID(c.Request.Context(), id)
		if gerr != nil {
			respondWithNotFoundProblem(c, inviteNotFound)

			return
		}

		organizer, oerr := s.organizerRepo.GetByOidcSub(c.Request.Context(), uid)
		if oerr != nil || organizer == nil || existing.InvitedByID != organizer.ID ||
			existing.Status != domain.InviteStatusPendingApproval {
			respondWithForbiddenProblem(c, "You can only delete your own pending invites")

			return
		}
	}

	if err := s.inviteRepo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, inviteNotFound)

			return
		}

		respondWithInternalServerProblem(c, "Failed to delete invite: "+err.Error())

		return
	}

	slog.Info("Delete Invite: Successfully deleted invite", "id", id)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func parseInviteListParams(c *gin.Context) repository.InviteListParams {
	return repository.InviteListParams{
		Offset: parseIntQueryParam(c, "_start", 0),
		Limit:  parseIntQueryParam(c, "_end", 0) - parseIntQueryParam(c, "_start", 0),
		Sort:   c.Query("_sort"),
		Order:  c.Query("_order"),
	}
}

func parseInviteFilters(c *gin.Context) repository.InviteListFilters {
	filters := repository.InviteListFilters{}

	if q := c.Query("q"); q != "" {
		filters.Query = &q
	}

	return filters
}
