package hub

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) ListTimetableEvents(c *gin.Context) {
	start := time.Now().UTC().Add(-24 * time.Hour)
	end := time.Now().UTC().Add(7 * 24 * time.Hour)

	if v := c.Query("start"); v != "" {
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			start = parsed
		}
	}

	if v := c.Query("end"); v != "" {
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			end = parsed
		}
	}

	items, err := s.timetableEventRepo.ListByRange(c.Request.Context(), start, end)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list timetable events: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.Itoa(len(items)))
	c.JSON(http.StatusOK, items)
}
