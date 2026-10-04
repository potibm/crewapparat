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
	var eventYear domain.EventYear
	if err := c.ShouldBindJSON(&eventYear); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	if err := s.eventYearRepo.Create(c.Request.Context(), &eventYear); err != nil {
		respondWithInternalServerProblem(c, "Failed to create event year: "+err.Error())

		return
	}

	slog.Info("Create EventYear: Successfully created event year", "year", eventYear.Year)
	c.JSON(http.StatusCreated, toEventYearResponse(eventYear))
}

func (s *Server) UpdateEventYear(c *gin.Context) {
	year := c.Param("year")

	var eventYear domain.EventYear
	if err := c.ShouldBindJSON(&eventYear); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	eventYear.Year = year

	if err := s.eventYearRepo.Update(c.Request.Context(), &eventYear); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Event year not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to update event year: "+err.Error())

		return
	}

	slog.Info("Update EventYear: Successfully updated event year", "year", eventYear.Year)
	c.JSON(http.StatusOK, toEventYearResponse(eventYear))
}

func (s *Server) DeleteEventYear(c *gin.Context) {
	year := c.Param("year")

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
