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

func (s *Server) ListOrganizers(c *gin.Context) {
	params := parseOrganizerListParams(c)
	filters := parseOrganizerFilters(c)

	items, total, err := s.organizerRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list organizers: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, items)
}

func (s *Server) GetOrganizerBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	organizer, err := s.organizerRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Organizer not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get organizer: "+err.Error())

		return
	}

	c.JSON(http.StatusOK, organizer)
}

func (s *Server) GetOrganizer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	organizer, err := s.organizerRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Organizer not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get organizer: "+err.Error())

		return
	}

	c.JSON(http.StatusOK, organizer)
}

func (s *Server) CreateOrganizer(c *gin.Context) {
	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can create organizers")

		return
	}

	var organizer domain.Organizer
	if err := c.ShouldBindJSON(&organizer); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	if organizer.Slug == "" {
		base := domain.Slugify(organizer.Handle)

		slug := base
		for i := 2; ; i++ {
			if _, err := s.organizerRepo.GetBySlug(c.Request.Context(), slug); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					break
				}

				respondWithInternalServerProblem(c, "Failed to check slug: "+err.Error())

				return
			}

			slug = base + "-" + strconv.Itoa(i)
		}

		organizer.Slug = slug
	}

	if err := s.organizerRepo.Create(c.Request.Context(), &organizer); err != nil {
		respondWithInternalServerProblem(c, "Failed to create organizer: "+err.Error())

		return
	}

	slog.Info("Create Organizer: Successfully created organizer", "id", organizer.ID)
	c.JSON(http.StatusCreated, organizer)
}

func (s *Server) UpdateOrganizer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		existing, err := s.organizerRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				respondWithNotFoundProblem(c, "Organizer not found")

				return
			}

			respondWithInternalServerProblem(c, "Failed to get organizer: "+err.Error())

			return
		}

		if existing.OidcSub == nil || *existing.OidcSub != uid {
			respondWithForbiddenProblem(c, "You can only update your own organizer profile")

			return
		}
	}

	var organizer domain.Organizer
	if err := c.ShouldBindJSON(&organizer); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	organizer.ID = id

	// Slug is immutable: preserve the stored value regardless of the payload.
	if existing, err := s.organizerRepo.GetByID(c.Request.Context(), id); err == nil && existing.Slug != "" {
		organizer.Slug = existing.Slug
	}

	if err := s.organizerRepo.Update(c.Request.Context(), &organizer); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Organizer not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to update organizer: "+err.Error())

		return
	}

	slog.Info("Update Organizer: Successfully updated organizer", "id", organizer.ID)
	c.JSON(http.StatusOK, organizer)
}

func (s *Server) DeleteOrganizer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can delete organizers")

		return
	}

	if err := s.organizerRepo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Organizer not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to delete organizer: "+err.Error())

		return
	}

	slog.Info("Delete Organizer: Successfully deleted organizer", "id", id)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func parseOrganizerListParams(c *gin.Context) repository.OrganizerListParams {
	return repository.OrganizerListParams{
		Offset: parseIntQueryParam(c, "_start", 0),
		Limit:  parseIntQueryParam(c, "_end", 0) - parseIntQueryParam(c, "_start", 0),
		Sort:   c.Query("_sort"),
		Order:  c.Query("_order"),
	}
}

func parseOrganizerFilters(c *gin.Context) repository.OrganizerListFilters {
	filters := repository.OrganizerListFilters{}

	if q := c.Query("q"); q != "" {
		filters.Query = &q
	}

	if slug := c.Query("slug"); slug != "" {
		filters.Slug = &slug
	}

	return filters
}
