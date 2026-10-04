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

func (s *Server) ListTeams(c *gin.Context) {
	params := parseTeamListParams(c)
	filters := parseTeamFilters(c)

	items, total, err := s.teamRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list teams: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

func (s *Server) GetTeam(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	team, err := s.teamRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Team not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get team: "+err.Error())

		return
	}

	c.JSON(http.StatusOK, team)
}

func (s *Server) CreateTeam(c *gin.Context) {
	var team domain.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	if err := s.teamRepo.Create(c.Request.Context(), &team); err != nil {
		respondWithInternalServerProblem(c, "Failed to create team: "+err.Error())

		return
	}

	slog.Info("Create Team: Successfully created team", "id", team.ID)
	c.JSON(http.StatusCreated, team)
}

func (s *Server) UpdateTeam(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	var team domain.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	team.ID = id

	if err := s.teamRepo.Update(c.Request.Context(), &team); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Team not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to update team: "+err.Error())

		return
	}

	slog.Info("Update Team: Successfully updated team", "id", team.ID)
	c.JSON(http.StatusOK, team)
}

func (s *Server) DeleteTeam(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	if err := s.teamRepo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Team not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to delete team: "+err.Error())

		return
	}

	slog.Info("Delete Team: Successfully deleted team", "id", id)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func parseTeamListParams(c *gin.Context) repository.TeamListParams {
	return repository.TeamListParams{
		Offset: parseIntQueryParam(c, "_start", 0),
		Limit:  parseIntQueryParam(c, "_end", 0) - parseIntQueryParam(c, "_start", 0),
		Sort:   c.Query("_sort"),
		Order:  c.Query("_order"),
	}
}

func parseTeamFilters(c *gin.Context) repository.TeamListFilters {
	filters := repository.TeamListFilters{}

	if q := c.Query("q"); q != "" {
		filters.Query = &q
	}

	return filters
}
