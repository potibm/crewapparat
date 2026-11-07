package hub

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	// Enrich with the inviting organizer's name (API-only field, not persisted).
	names := make(map[uuid.UUID]string)
	for i := range items {
		if _, ok := names[items[i].InvitedByID]; !ok {
			names[items[i].InvitedByID] = s.organizerDisplayName(
				c.Request.Context(),
				items[i].InvitedByID,
			)
		}

		items[i].InvitedByName = names[items[i].InvitedByID]
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

func canUpdateInvite(s *Server, c *gin.Context, existing *domain.Invite, uid string) bool {
	organizer, err := s.organizerRepo.GetByOidcSub(c.Request.Context(), uid)
	if err != nil || organizer == nil {
		return false
	}

	if existing.InvitedByID != organizer.ID {
		return false
	}

	return existing.Status == domain.InviteStatusPendingApproval
}

func triggerInviteLinkEmail(s *Server, invite domain.Invite) {
	go func() {
		link := strings.TrimRight(s.mailer.FrontendBaseURL(), "/") + "/invite/" + invite.Token

		data := mailer.InviteLinkData{
			EventName: s.eventName(),
			Link:      link,
			ExpiresAt: invite.ExpiresAt,
		}
		if err := s.mailer.SendInviteLink(invite.Email, data); err != nil {
			slog.Warn("failed to send invite link email",
				"invite_id", invite.ID,
				"error", err,
			)
		}
	}()
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

	// Magic-link credential and expiry are always generated server-side.
	invite.Token = domain.NewInviteToken()
	invite.ExpiresAt = time.Now().UTC().Add(14 * 24 * time.Hour)

	// These fields are filled in by the invitee through the magic link, never
	// by the inviting organizer.
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
			link := strings.TrimRight(s.mailer.FrontendBaseURL(), "/") + "/invites/" + invite.ID.String()

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
	} else if invite.Status == domain.InviteStatusSent {
		triggerInviteLinkEmail(s, invite)
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

	if !isAdmin && !canUpdateInvite(s, c, existing, uid) {
		respondWithForbiddenProblem(c, "You can only update your own pending invites")

		return
	}

	var invite domain.Invite
	if err := c.ShouldBindJSON(&invite); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	invite.ID = id

	preserveInviteFields(&invite, existing)

	s.applyInviteRoleFields(c, &invite, existing, isAdmin)

	if !s.maybeProvisionOrganizer(c, &invite, isAdmin) {
		return
	}

	if err := s.inviteRepo.Update(c.Request.Context(), &invite); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, inviteNotFound)

			return
		}

		respondWithInternalServerProblem(c, "Failed to update invite: "+err.Error())

		return
	}

	if isAdmin && invite.Status == domain.InviteStatusSent && existing.Status != domain.InviteStatusSent {
		triggerInviteLinkEmail(s, invite)
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
	sort := c.Query("_sort")
	if sort == "" {
		sort = "created_at"
	}

	order := c.Query("_order")
	if order == "" {
		order = "DESC"
	}

	return repository.InviteListParams{
		Offset: parseIntQueryParam(c, "_start", 0),
		Limit:  parseIntQueryParam(c, "_end", 0) - parseIntQueryParam(c, "_start", 0),
		Sort:   sort,
		Order:  order,
	}
}

// provisionOrganizerFromInvite creates an organizer record from an approved
// invite (plus a confirmed participation for the active event year, if any).
// It is idempotent: if an organizer with the invite's email already exists, no
// duplicate is created and the existing record is reused.
func (s *Server) organizerDisplayName(ctx context.Context, id uuid.UUID) string {
	if s.organizerRepo == nil {
		return ""
	}

	org, err := s.organizerRepo.GetByID(ctx, id)
	if err != nil || org == nil {
		return ""
	}

	if org.Realname != "" {
		return org.Realname
	}

	return org.Handle
}

func (s *Server) createParticipationIfActive(ctx context.Context, organizerID uuid.UUID) {
	if s.eventYearRepo == nil || s.participationRepo == nil {
		return
	}

	active, err := s.eventYearRepo.GetActive(ctx)
	if err != nil || active == nil {
		return
	}

	_, gerr := s.participationRepo.GetByOrganizerAndYear(ctx, organizerID, active.Year)
	if !errors.Is(gerr, gorm.ErrRecordNotFound) {
		return
	}

	participation := domain.Participation{
		OrganizerID: organizerID,
		EventYearID: active.Year,
		Status:      domain.ParticipationStatusConfirmed,
	}
	if cerr := s.participationRepo.Create(ctx, &participation); cerr != nil {
		slog.Warn("failed to create participation for approved invite",
			"organizer_id", organizerID,
			"event_year", active.Year,
			"error", cerr,
		)
	}
}

func (s *Server) provisionOrganizerFromInvite(ctx context.Context, invite *domain.Invite) error {
	organizer, err := s.organizerRepo.GetByEmail(ctx, invite.Email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		slug, slugErr := s.uniqueOrganizerSlug(ctx, invite.Handle)
		if slugErr != nil {
			return slugErr
		}

		newOrganizer := domain.Organizer{
			Handle:    invite.Handle,
			Email:     invite.Email,
			DemozooID: invite.DemozooID,
			Slug:      slug,
		}
		if createErr := s.organizerRepo.Create(ctx, &newOrganizer); createErr != nil {
			return createErr
		}

		organizer = &newOrganizer
	}

	s.createParticipationIfActive(ctx, organizer.ID)

	return nil
}

func parseInviteFilters(c *gin.Context) repository.InviteListFilters {
	filters := repository.InviteListFilters{}

	if q := c.Query("q"); q != "" {
		filters.Query = &q
	}

	if v := c.Query("status"); v != "" {
		status := domain.InviteStatus(v)
		filters.Status = &status
	}

	return filters
}

func preserveInviteFields(invite, existing *domain.Invite) {
	// Server-managed fields that are not part of the edit form.
	invite.InvitedByID = existing.InvitedByID
	invite.Token = existing.Token
	invite.ExpiresAt = existing.ExpiresAt

	// Fields submitted by the invitee through the magic link.
	invite.Notes = existing.Notes
	invite.CoCAccepted = existing.CoCAccepted
	invite.CoCAcceptedAt = existing.CoCAcceptedAt
	invite.CoCAcceptedVersion = existing.CoCAcceptedVersion

	// Invite-provided fields, kept when the edit form omits them.
	if invite.Email == "" {
		invite.Email = existing.Email
	}

	if strings.TrimSpace(invite.Handle) == "" {
		invite.Handle = existing.Handle
	}

	if invite.DemozooID == nil {
		invite.DemozooID = existing.DemozooID
	}

	if invite.Reason == "" {
		invite.Reason = existing.Reason
	}
}

func (s *Server) applyInviteRoleFields(
	c *gin.Context,
	invite, existing *domain.Invite,
	isAdmin bool,
) {
	if !isAdmin {
		// Non-admins cannot change the status away from pending_approval.
		invite.Status = domain.InviteStatusPendingApproval
		invite.ReviewNote = existing.ReviewNote
		invite.OpenIDAccountCreated = existing.OpenIDAccountCreated
		invite.SlackInvited = existing.SlackInvited

		return
	}

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if organizer, err := s.organizerRepo.GetByOidcSub(c.Request.Context(), uid); err == nil && organizer != nil {
		reviewer := organizer.ID
		invite.ReviewedByID = &reviewer
	}
}

func (s *Server) maybeProvisionOrganizer(
	c *gin.Context,
	invite *domain.Invite,
	isAdmin bool,
) bool {
	if !isAdmin || invite.Status != domain.InviteStatusApproved {
		return true
	}

	if strings.TrimSpace(invite.Handle) == "" {
		respondWithBadRequestProblem(c, "handle is required to approve an invite")

		return false
	}

	if err := s.provisionOrganizerFromInvite(c.Request.Context(), invite); err != nil {
		respondWithInternalServerProblem(c, "Failed to provision organizer: "+err.Error())

		return false
	}

	return true
}
