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

func (s *Server) ListParticipations(c *gin.Context) {
	params := parseParticipationListParams(c)
	filters := parseParticipationFilters(c)

	items, total, err := s.participationRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list participations: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

func (s *Server) GetParticipation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	participation, err := s.participationRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Participation not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get participation: "+err.Error())

		return
	}

	c.JSON(http.StatusOK, participation)
}

func (s *Server) CreateParticipation(c *gin.Context) {
	var participation domain.Participation
	if err := c.ShouldBindJSON(&participation); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	if err := s.participationRepo.Create(c.Request.Context(), &participation); err != nil {
		respondWithInternalServerProblem(c, "Failed to create participation: "+err.Error())

		return
	}

	slog.Info("Create Participation: Successfully created participation", "id", participation.ID)
	c.JSON(http.StatusCreated, participation)
}

func (s *Server) UpdateParticipation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	var participation domain.Participation
	if err := c.ShouldBindJSON(&participation); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	participation.ID = id

	if err := s.participationRepo.Update(c.Request.Context(), &participation); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Participation not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to update participation: "+err.Error())

		return
	}

	slog.Info("Update Participation: Successfully updated participation", "id", participation.ID)
	c.JSON(http.StatusOK, participation)
}

func (s *Server) DeleteParticipation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	if err := s.participationRepo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Participation not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to delete participation: "+err.Error())

		return
	}

	slog.Info("Delete Participation: Successfully deleted participation", "id", id)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func parseParticipationListParams(c *gin.Context) repository.ParticipationListParams {
	return repository.ParticipationListParams{
		Offset: parseIntQueryParam(c, "_start", 0),
		Limit:  parseIntQueryParam(c, "_end", 0) - parseIntQueryParam(c, "_start", 0),
		Sort:   c.Query("_sort"),
		Order:  c.Query("_order"),
	}
}

func parseParticipationFilters(c *gin.Context) repository.ParticipationListFilters {
	filters := repository.ParticipationListFilters{}

	if q := c.Query("q"); q != "" {
		filters.Query = &q
	}

	if v := c.Query("organizer_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filters.OrganizerID = &id
		}
	}

	if v := c.Query("event_year_id"); v != "" {
		filters.EventYearID = &v
	}

	if v := c.Query("status"); v != "" {
		status := domain.ParticipationStatus(v)
		filters.Status = &status
	}

	if v := c.Query("team"); v != "" {
		filters.TeamName = &v
	}

	return filters
}
