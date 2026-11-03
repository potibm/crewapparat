package hub

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

type acceptCodeOfConductRequest struct {
	CodeOfConductID uuid.UUID `json:"code_of_conduct_id"`
}

// AcceptCodeOfConduct records that the authenticated user accepts the current
// published version. Idempotent: accepting twice returns the original
// timestamp.
func (s *Server) AcceptCodeOfConduct(c *gin.Context) {
	organizer, ok := s.currentOrganizer(c)
	if !ok {
		return
	}

	var req acceptCodeOfConductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	coc, err := s.codeOfConductRepo.GetLatestPublished(c.Request.Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "No published code of conduct")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get code of conduct: "+err.Error())

		return
	}

	if coc.ID != req.CodeOfConductID {
		respondWithConflictProblem(c, "The provided code of conduct is not the current published version")

		return
	}

	existing, err := s.acceptanceRepo.GetByOrganizerAndCodeOfConduct(
		c.Request.Context(),
		organizer.ID,
		coc.ID,
	)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"accepted_at": existing.AcceptedAt})

		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		respondWithInternalServerProblem(c, "Failed to get code of conduct acceptance: "+err.Error())

		return
	}

	acceptance := &domain.CodeOfConductAcceptance{
		OrganizerID:     organizer.ID,
		CodeOfConductID: coc.ID,
		AcceptedAt:      time.Now().UTC(),
	}

	if err := s.acceptanceRepo.Create(c.Request.Context(), acceptance); err != nil {
		// A concurrent request may have won the unique constraint race.
		existing, getErr := s.acceptanceRepo.GetByOrganizerAndCodeOfConduct(
			c.Request.Context(),
			organizer.ID,
			coc.ID,
		)
		if getErr == nil {
			c.JSON(http.StatusOK, gin.H{"accepted_at": existing.AcceptedAt})

			return
		}

		respondWithInternalServerProblem(c, "Failed to create code of conduct acceptance: "+err.Error())

		return
	}

	slog.Info(
		"Accept CodeOfConduct: recorded acceptance",
		"organizer_id", organizer.ID,
		"code_of_conduct_id", coc.ID,
		"version", coc.Version,
	)

	c.JSON(http.StatusOK, gin.H{"accepted_at": acceptance.AcceptedAt})
}

// ListCodeOfConductAcceptances returns who accepted which version. Admin only.
func (s *Server) ListCodeOfConductAcceptances(c *gin.Context) {
	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can list code of conduct acceptances")

		return
	}

	params := repository.CodeOfConductAcceptanceListParams{
		Offset: parseIntQueryParam(c, "_start", 0),
		Limit:  parseIntQueryParam(c, "_end", 0) - parseIntQueryParam(c, "_start", 0),
		Sort:   c.Query("_sort"),
		Order:  c.Query("_order"),
	}

	filters := repository.CodeOfConductAcceptanceListFilters{}

	if v := c.Query("code_of_conduct_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			respondWithInvalidIDFormatProblem(c)

			return
		}

		filters.CodeOfConductID = &id
	}

	if v := c.Query("organizer_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			respondWithInvalidIDFormatProblem(c)

			return
		}

		filters.OrganizerID = &id
	}

	items, total, err := s.acceptanceRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list code of conduct acceptances: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

// currentOrganizer resolves the organizer of the authenticated user and
// responds with the matching problem when that is not possible.
func (s *Server) currentOrganizer(c *gin.Context) (*domain.Organizer, bool) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		respondWithUnauthorizedProblem(c, "User ID not found in context")

		return nil, false
	}

	userID, _ := userIDVal.(string)

	organizer, err := s.organizerRepo.GetByOidcSub(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Organizer not found")

			return nil, false
		}

		respondWithInternalServerProblem(c, "Failed to get organizer: "+err.Error())

		return nil, false
	}

	return organizer, true
}
