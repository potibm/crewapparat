package hub

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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

//nolint:nestif // multipart/JSON branching
func (s *Server) CreateOrganizer(c *gin.Context) {
	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		respondWithForbiddenProblem(c, "Only admins can create organizers")

		return
	}

	fileHeader, _ := c.FormFile("media_upload")
	isMultipart := fileHeader != nil || strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data")

	var organizer domain.Organizer

	if isMultipart {
		organizer.Handle = c.PostForm("handle")
		organizer.Realname = c.PostForm("realname")
		organizer.Email = c.PostForm("email")
		organizer.Phone = c.PostForm("phone")
		organizer.SlackHandle = c.PostForm("slack_handle")
		organizer.SlackProfileURL = c.PostForm("slack_profile_url")
		organizer.Pronouns = c.PostForm("pronouns")
		organizer.Languages = c.PostFormArray("languages")

		organizer.AvatarURL = c.PostForm("avatar_url")
		if v := c.PostForm("demozoo_id"); v != "" {
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				u := uint(n)
				organizer.DemozooID = &u
			}
		}
	} else {
		if err := c.ShouldBindJSON(&organizer); err != nil {
			respondWithFailedToParsePayloadProblem(c, err)

			return
		}
	}

	if fileHeader != nil {
		url, err := saveAvatarFile(fileHeader)
		if err != nil {
			respondWithAvatarStorageProblem(c, err)

			return
		}

		organizer.AvatarURL = url
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

//nolint:nestif,gocognit,cyclop // multipart/JSON branching
func (s *Server) UpdateOrganizer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	existing, err := s.organizerRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Organizer not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get organizer: "+err.Error())

		return
	}

	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if !s.isAdmin(c, uid) {
		if existing.OidcSub == nil || *existing.OidcSub != uid {
			respondWithForbiddenProblem(c, "You can only update your own organizer profile")

			return
		}
	}

	fileHeader, _ := c.FormFile("media_upload")
	isMultipart := fileHeader != nil || strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data")

	removeAvatar := false

	var organizer domain.Organizer

	if isMultipart {
		organizer.Handle = c.PostForm("handle")
		organizer.Realname = c.PostForm("realname")
		organizer.Email = c.PostForm("email")
		organizer.Phone = c.PostForm("phone")
		organizer.SlackHandle = c.PostForm("slack_handle")
		organizer.SlackProfileURL = c.PostForm("slack_profile_url")
		organizer.Pronouns = c.PostForm("pronouns")
		organizer.Languages = c.PostFormArray("languages")

		organizer.AvatarURL = c.PostForm("avatar_url")
		if c.PostForm("remove_avatar") == "true" {
			removeAvatar = true
		}

		if v := c.PostForm("demozoo_id"); v != "" {
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				u := uint(n)
				organizer.DemozooID = &u
			}
		}
	} else {
		var payload struct {
			domain.Organizer

			RemoveAvatar bool `json:"remove_avatar"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			respondWithFailedToParsePayloadProblem(c, err)

			return
		}

		organizer = payload.Organizer
		removeAvatar = payload.RemoveAvatar
	}

	organizer.ID = id

	if fileHeader != nil {
		url, err := saveAvatarFile(fileHeader)
		if err != nil {
			respondWithAvatarStorageProblem(c, err)

			return
		}

		organizer.AvatarURL = url
	}

	// Preserve immutable / system fields that are not part of the edit form.
	if organizer.OidcSub == nil {
		organizer.OidcSub = existing.OidcSub
	}

	if organizer.Slug == "" {
		organizer.Slug = existing.Slug
	}

	// A newly uploaded avatar always wins over a removal request, so a file is
	// never written to disk and then discarded.
	if fileHeader == nil && removeAvatar {
		organizer.AvatarURL = ""
	} else if organizer.AvatarURL == "" && fileHeader == nil {
		organizer.AvatarURL = existing.AvatarURL
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

	if v := c.Query("event_year_id"); v != "" {
		filters.EventYearID = &v
	}

	if v := c.Query("status"); v != "" {
		filters.Status = &v
	}

	if v := c.Query("team"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filters.TeamID = &id
		}
	}

	return filters
}
