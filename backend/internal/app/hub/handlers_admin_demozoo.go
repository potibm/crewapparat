package hub

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/potibm/crewapparat/internal/app/demozoo"
)

func filterDemozooSceners(
	sceners []demozoo.Scener,
	q string,
	numericQuery bool,
	parsedID int64,
	idSet map[int64]struct{},
) []demozoo.Scener {
	result := make([]demozoo.Scener, 0, len(sceners))
	for _, sc := range sceners {
		if len(idSet) > 0 {
			if _, ok := idSet[sc.ID]; !ok {
				continue
			}
		}

		if q != "" {
			nameMatch := strings.Contains(strings.ToLower(sc.Name), q)

			idMatch := numericQuery && sc.ID == parsedID
			if !nameMatch && !idMatch {
				continue
			}
		}

		result = append(result, sc)
	}

	return result
}

func sortDemozooSceners(
	sceners []demozoo.Scener,
	q string,
	numericQuery bool,
	parsedID int64,
	sortField, order string,
) {
	if q != "" {
		sort.Slice(sceners, func(i, j int) bool {
			exactI := numericQuery && sceners[i].ID == parsedID

			exactJ := numericQuery && sceners[j].ID == parsedID
			if exactI != exactJ {
				return exactI
			}

			startsI := strings.HasPrefix(strings.ToLower(sceners[i].Name), q)

			startsJ := strings.HasPrefix(strings.ToLower(sceners[j].Name), q)
			if startsI != startsJ {
				return startsI
			}

			return strings.ToLower(sceners[i].Name) < strings.ToLower(sceners[j].Name)
		})

		return
	}

	if sortField == "" {
		sortField = "name"
	}

	desc := strings.ToLower(order) == "desc"

	sort.Slice(sceners, func(i, j int) bool {
		var less bool

		switch sortField {
		case "id":
			less = sceners[i].ID < sceners[j].ID
		default:
			less = strings.ToLower(sceners[i].Name) < strings.ToLower(sceners[j].Name)
		}

		if desc {
			return !less
		}

		return less
	})
}

func (s *Server) ListDemozooSceners(c *gin.Context) {
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))

	idFilters := c.QueryArray("id")
	idSet := map[int64]struct{}{}

	for _, v := range idFilters {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			idSet[n] = struct{}{}
		}
	}

	numericQuery := false
	parsedID := int64(0)

	if q != "" && isAllDigits(q) {
		numericQuery = true

		if n, err := strconv.Atoi(q); err == nil {
			parsedID = int64(n)
		}
	}

	sceners := filterDemozooSceners(s.demozooSceners, q, numericQuery, parsedID, idSet)

	sortDemozooSceners(sceners, q, numericQuery, parsedID, c.Query("_sort"), c.Query("_order"))

	total := len(sceners)

	start := parseIntQueryParam(c, "_start", 0)
	end := parseIntQueryParam(c, "_end", 0)

	if start < 0 {
		start = 0
	}

	const defaultPageSize = 25

	if end <= start {
		end = start + defaultPageSize
	}

	if start > total {
		start = total
	}

	if end > total {
		end = total
	}

	c.Header("X-Total-Count", strconv.Itoa(total))
	c.JSON(http.StatusOK, sceners[start:end])
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func (s *Server) GetDemozooScener(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})

		return
	}

	for _, sc := range s.demozooSceners {
		if sc.ID == id {
			c.JSON(http.StatusOK, sc)

			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "scener not found"})
}
