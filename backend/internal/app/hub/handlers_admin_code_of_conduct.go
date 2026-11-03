package hub

import (
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
)

func (s *Server) ListCodeOfConducts(c *gin.Context) {
	params := parseCodeOfConductListParams(c)
	filters := parseCodeOfConductFilters(c)

	items, total, err := s.codeOfConductRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list code of conducts: "+err.Error())

		return
	}

	for i := range items {
		count, countErr := s.codeOfConductRepo.CountAcceptances(c.Request.Context(), items[i].ID)
		if countErr != nil {
			respondWithInternalServerProblem(c, "Failed to count acceptances: "+countErr.Error())

			return
		}

		items[i].AcceptedCount = count
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

func (s *Server) GetCodeOfConduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	coc, err := s.codeOfConductRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Code of conduct not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get code of conduct: "+err.Error())

		return
	}

	count, err := s.codeOfConductRepo.CountAcceptances(c.Request.Context(), coc.ID)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to count acceptances: "+err.Error())

		return
	}

	coc.AcceptedCount = count

	c.JSON(http.StatusOK, coc)
}

func (s *Server) CreateCodeOfConduct(c *gin.Context) {
	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can create code of conducts")

		return
	}

	var coc domain.CodeOfConduct
	if err := c.ShouldBindJSON(&coc); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	if !validateCodeOfConductVersion(c, coc.Version) {
		return
	}

	if coc.Status == domain.CodeOfConductStatusPublished && coc.PublishedAt == nil {
		now := time.Now()
		coc.PublishedAt = &now
	}

	if err := s.codeOfConductRepo.Create(c.Request.Context(), &coc); err != nil {
		respondWithInternalServerProblem(c, "Failed to create code of conduct: "+err.Error())

		return
	}

	slog.Info("Create CodeOfConduct: Successfully created", "id", coc.ID)
	c.JSON(http.StatusCreated, coc)
}

func (s *Server) UpdateCodeOfConduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can update code of conducts")

		return
	}

	existing, err := s.codeOfConductRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Code of conduct not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get code of conduct: "+err.Error())

		return
	}

	if existing.Status == domain.CodeOfConductStatusPublished {
		respondWithConflictProblem(c, "Published code of conducts are immutable. Create a new version instead.")

		return
	}

	var coc domain.CodeOfConduct
	if err := c.ShouldBindJSON(&coc); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	if !validateCodeOfConductVersion(c, coc.Version) {
		return
	}

	coc.ID = id

	if coc.Status == domain.CodeOfConductStatusPublished && coc.PublishedAt == nil {
		if existing.PublishedAt != nil {
			coc.PublishedAt = existing.PublishedAt
		} else {
			now := time.Now()
			coc.PublishedAt = &now
		}
	}

	if err := s.codeOfConductRepo.Update(c.Request.Context(), &coc); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Code of conduct not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to update code of conduct: "+err.Error())

		return
	}

	slog.Info("Update CodeOfConduct: Successfully updated", "id", coc.ID)
	c.JSON(http.StatusOK, coc)
}

func (s *Server) DeleteCodeOfConduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can delete code of conducts")

		return
	}

	existing, err := s.codeOfConductRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Code of conduct not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get code of conduct: "+err.Error())

		return
	}

	if existing.Status == domain.CodeOfConductStatusPublished {
		respondWithConflictProblem(c, "Published code of conducts cannot be deleted. Create a new version instead.")

		return
	}

	if err := s.codeOfConductRepo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Code of conduct not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to delete code of conduct: "+err.Error())

		return
	}

	slog.Info("Delete CodeOfConduct: Successfully deleted", "id", id)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func parseCodeOfConductListParams(c *gin.Context) repository.CodeOfConductListParams {
	return repository.CodeOfConductListParams{
		Offset: parseIntQueryParam(c, "_start", 0),
		Limit:  parseIntQueryParam(c, "_end", 0) - parseIntQueryParam(c, "_start", 0),
		Sort:   c.Query("_sort"),
		Order:  c.Query("_order"),
	}
}

func parseCodeOfConductFilters(c *gin.Context) repository.CodeOfConductListFilters {
	filters := repository.CodeOfConductListFilters{}

	if v := c.Query("status"); v != "" {
		status := domain.CodeOfConductStatus(v)
		filters.Status = &status
	}

	return filters
}

var codeOfConductVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// validateCodeOfConductVersion ensures the version is a plain semantic
// version. Uniqueness is not enforced: keeping this simple is intentional.
func validateCodeOfConductVersion(c *gin.Context, version string) bool {
	if !codeOfConductVersionPattern.MatchString(version) {
		respondWithBadRequestProblem(c, "Version must be a semantic version like 1.0.0")

		return false
	}

	return true
}
