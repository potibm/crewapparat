package hub

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

type eventYearResponse struct {
	ID       string `json:"id"`
	Year     string `json:"year"`
	IsActive bool   `json:"is_active"`
}

func toEventYearResponse(year domain.EventYear) eventYearResponse {
	return eventYearResponse{
		ID:       year.Year,
		Year:     year.Year,
		IsActive: year.IsActive,
	}
}

func (s *Server) ListEventYears(c *gin.Context) {
	params := parseEventYearListParams(c)
	filters := parseEventYearFilters(c)

	items, total, err := s.eventYearRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list event years: "+err.Error())

		return
	}

	response := make([]eventYearResponse, 0, len(items))
	for i := range items {
		response = append(response, toEventYearResponse(items[i]))
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, response)
}

func (s *Server) GetEventYear(c *gin.Context) {
	year := c.Param("year")

	eventYear, err := s.eventYearRepo.GetByYear(c.Request.Context(), year)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Event year not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get event year: "+err.Error())

		return
	}

	c.JSON(http.StatusOK, toEventYearResponse(*eventYear))
}

func (s *Server) CreateEventYear(c *gin.Context) {
	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can create event years")

		return
	}

	var eventYear domain.EventYear
	if err := c.ShouldBindJSON(&eventYear); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	if err := s.eventYearRepo.Create(c.Request.Context(), &eventYear); err != nil {
		respondWithInternalServerProblem(c, "Failed to create event year: "+err.Error())

		return
	}

	// Create a pending participation for every organizer in the new year.
	organizers, _, err := s.organizerRepo.List(
		c.Request.Context(),
		repository.ListParams{},
		repository.OrganizerListFilters{},
	)
	if err != nil {
		slog.Warn("Create EventYear: failed to list organizers for participation seeding", "error", err)
	} else {
		for i := range organizers {
			participation := &domain.Participation{
				OrganizerID: organizers[i].ID,
				EventYearID: eventYear.Year,
				Status:      domain.ParticipationStatusPending,
			}
			if createErr := s.participationRepo.Create(c.Request.Context(), participation); createErr != nil {
				slog.Warn("Create EventYear: failed to create participation",
					"organizer_id", organizers[i].ID,
					"error", createErr,
				)
			}
		}
	}

	slog.Info("Create EventYear: Successfully created event year", "year", eventYear.Year)
	c.JSON(http.StatusCreated, toEventYearResponse(eventYear))
}

func (s *Server) UpdateEventYear(c *gin.Context) {
	year := c.Param("year")

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can update event years")

		return
	}

	var eventYear domain.EventYear
	if err := c.ShouldBindJSON(&eventYear); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	eventYear.Year = year

	// Only allow activating a year via update.
	if eventYear.IsActive {
		if err := s.eventYearRepo.ActivateYear(c.Request.Context(), year); err != nil {
			respondWithInternalServerProblem(c, "Failed to activate event year: "+err.Error())

			return
		}
	}

	slog.Info("Update EventYear: Successfully updated event year", "year", eventYear.Year)
	c.JSON(http.StatusOK, toEventYearResponse(eventYear))
}

func (s *Server) DeleteEventYear(c *gin.Context) {
	year := c.Param("year")

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can delete event years")

		return
	}

	// Prevent deletion of the active year.
	eventYear, err := s.eventYearRepo.GetByYear(c.Request.Context(), year)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Event year not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get event year: "+err.Error())

		return
	}

	if eventYear.IsActive {
		respondWithConflictProblem(c, "Cannot delete the active event year. Activate another year first.")

		return
	}

	// Delete all participations for this year first.
	participations, _, err := s.participationRepo.List(
		c.Request.Context(),
		repository.ParticipationListParams{},
		repository.ParticipationListFilters{EventYearID: &year},
	)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list participations: "+err.Error())

		return
	}

	for i := range participations {
		if delErr := s.participationRepo.Delete(c.Request.Context(), participations[i].ID); delErr != nil {
			slog.Warn("Delete EventYear: failed to delete participation",
				"participation_id", participations[i].ID,
				"error", delErr,
			)
		}
	}

	if err := s.eventYearRepo.Delete(c.Request.Context(), year); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Event year not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to delete event year: "+err.Error())

		return
	}

	slog.Info("Delete EventYear: Successfully deleted event year", "year", year)
	c.JSON(http.StatusOK, gin.H{"year": year})
}

func parseEventYearListParams(c *gin.Context) repository.EventYearListParams {
	return repository.EventYearListParams{
		Offset: parseIntQueryParam(c, "_start", 0),
		Limit:  parseIntQueryParam(c, "_end", 0) - parseIntQueryParam(c, "_start", 0),
		Sort:   c.Query("_sort"),
		Order:  c.Query("_order"),
	}
}

func parseEventYearFilters(c *gin.Context) repository.EventYearListFilters {
	filters := repository.EventYearListFilters{}

	if q := c.Query("q"); q != "" {
		filters.Query = &q
	}

	return filters
}
