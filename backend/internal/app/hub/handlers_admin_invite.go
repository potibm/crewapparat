package hub

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

func (s *Server) ListInvites(c *gin.Context) {
	params := parseInviteListParams(c)
	filters := parseInviteFilters(c)

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
			respondWithNotFoundProblem(c, "Invite not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get invite: "+err.Error())

		return
	}

	c.JSON(http.StatusOK, invite)
}

func (s *Server) CreateInvite(c *gin.Context) {
	var invite domain.Invite
	if err := c.ShouldBindJSON(&invite); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	if err := s.inviteRepo.Create(c.Request.Context(), &invite); err != nil {
		respondWithInternalServerProblem(c, "Failed to create invite: "+err.Error())

		return
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

	var invite domain.Invite
	if err := c.ShouldBindJSON(&invite); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	invite.ID = id

	if err := s.inviteRepo.Update(c.Request.Context(), &invite); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Invite not found")

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

	if err := s.inviteRepo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Invite not found")

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
