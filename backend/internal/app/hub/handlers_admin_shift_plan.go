package hub

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

func (s *Server) ListShiftPlans(c *gin.Context) {
	params := parseShiftPlanListParams(c)
	filters := parseShiftPlanFilters(c)

	items, total, err := s.shiftPlanRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list shift plans: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

func (s *Server) GetShiftPlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	plan, err := s.shiftPlanRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Shift plan not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get shift plan: "+err.Error())

		return
	}

	c.JSON(http.StatusOK, plan)
}

func parseShiftPlanListParams(c *gin.Context) repository.ShiftPlanListParams {
	return repository.ShiftPlanListParams{
		Offset: parseIntQueryParam(c, "_start", 0),
		Limit:  parseIntQueryParam(c, "_end", 0) - parseIntQueryParam(c, "_start", 0),
		Sort:   c.Query("_sort"),
		Order:  c.Query("_order"),
	}
}

func parseShiftPlanFilters(c *gin.Context) repository.ShiftPlanListFilters {
	filters := repository.ShiftPlanListFilters{}

	if v := c.Query("event_year_id"); v != "" {
		filters.EventYearID = &v
	}

	if v := c.Query("status"); v != "" {
		status := domain.ShiftPlanStatus(v)
		filters.Status = &status
	}

	return filters
}
