package hub

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

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
			respondWithNotFoundProblem(c, organizerNotFound)

			return
		}

		respondWithInternalServerProblem(c, organizerGetErrorPrefix+err.Error())

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
			respondWithNotFoundProblem(c, organizerNotFound)

			return
		}

		respondWithInternalServerProblem(c, organizerGetErrorPrefix+err.Error())

		return
	}

	// Populate the active year's participation so the show page can render the
	// current status and teams without a second lookup.
	if s.eventYearRepo != nil {
		active, activeErr := s.eventYearRepo.GetActive(c.Request.Context())
		if activeErr == nil && active != nil {
			for i := range organizer.Participations {
				if organizer.Participations[i].EventYearID != active.Year {
					continue
				}

				participation := organizer.Participations[i]
				organizer.CurrentParticipationStatus = string(participation.Status)
				organizer.CurrentParticipationTeams = participation.Teams

				break
			}
		}
	}

	c.JSON(http.StatusOK, organizer)
}

// uniqueOrganizerSlug builds a unique slug from the handle by appending a
// counter when the base slug is already taken.
func (s *Server) uniqueOrganizerSlug(ctx context.Context, handle string) (string, error) {
	base := domain.Slugify(handle)
	slug := base

	for i := 2; ; i++ {
		if _, err := s.organizerRepo.GetBySlug(ctx, slug); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return slug, nil
			}

			return "", err
		}

		slug = base + "-" + strconv.Itoa(i)
	}
}

// canEditOrganizer allows admins and the organizer themselves to edit a
// profile.
func (s *Server) canEditOrganizer(c *gin.Context, organizer *domain.Organizer) bool {
	userID, _ := c.Get("userID")

	uid, _ := userID.(string)
	if uid == "" {
		return false
	}

	if s.isAdmin(c, uid) {
		return true
	}

	return organizer.OidcSub != nil && *organizer.OidcSub == uid
}

// preserveOrganizerSystemFields keeps values that are not part of the edit
// form but must survive an update.
func preserveOrganizerSystemFields(organizer, existing *domain.Organizer) {
	if organizer.OidcSub == nil {
		organizer.OidcSub = existing.OidcSub
	}

	if organizer.Slug == "" {
		organizer.Slug = existing.Slug
	}
}

// resolveAvatarURL decides which avatar URL to persist: a newly uploaded file
// wins over everything, then an explicit removal, then an explicitly provided
// URL, and finally the previously stored one.
func resolveAvatarURL(
	fileHeader *multipart.FileHeader,
	providedURL, existingURL string,
	removeAvatar bool,
) (string, error) {
	if fileHeader != nil {
		return saveAvatarFile(fileHeader)
	}

	if removeAvatar {
		return "", nil
	}

	if providedURL != "" {
		return providedURL, nil
	}

	return existingURL, nil
}

// parseOrganizerFromMultipart extracts the organizer fields from a
// multipart/form-data request. URL and first-year validation stay in the
// handlers so this only does mechanical parsing.
func parseOrganizerFromMultipart(c *gin.Context) (domain.Organizer, error) {
	var organizer domain.Organizer

	organizer.Handle = c.PostForm("handle")
	organizer.Realname = c.PostForm("realname")
	organizer.Email = c.PostForm("email")
	organizer.Phone = c.PostForm("phone")
	organizer.SlackHandle = c.PostForm("slack_handle")
	organizer.SlackProfileURL = c.PostForm("slack_profile_url")
	organizer.MediaWikiProfileURL = c.PostForm("mediawiki_profile_url")
	organizer.Pronouns = c.PostForm("pronouns")
	organizer.Languages = c.PostFormArray("languages")
	organizer.AvatarURL = c.PostForm("avatar_url")

	if v := c.PostForm("demozoo_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			u := uint(n)
			organizer.DemozooID = &u
		}
	}

	if v := c.PostForm("first_year_as_organizer"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return organizer, fmt.Errorf("invalid first_year_as_organizer: must be an integer")
		}

		organizer.FirstYearAsOrganizer = &n
	}

	return organizer, nil
}

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
		parsed, err := parseOrganizerFromMultipart(c)
		if err != nil {
			respondWithBadRequestProblem(c, err.Error())

			return
		}

		organizer = parsed
	} else {
		if err := c.ShouldBindJSON(&organizer); err != nil {
			respondWithFailedToParsePayloadProblem(c, err)

			return
		}
	}

	avatarURL, err := resolveAvatarURL(fileHeader, organizer.AvatarURL, "", false)
	if err != nil {
		respondWithAvatarStorageProblem(c, err)

		return
	}

	organizer.AvatarURL = avatarURL

	if err := validateOrganizerProfileURLs(organizer.SlackProfileURL, organizer.MediaWikiProfileURL); err != nil {
		respondWithBadRequestProblem(c, err.Error())

		return
	}

	if err := validateFirstYearAsOrganizer(organizer.FirstYearAsOrganizer); err != nil {
		respondWithBadRequestProblem(c, err.Error())

		return
	}

	if organizer.Slug == "" {
		slug, err := s.uniqueOrganizerSlug(c.Request.Context(), organizer.Handle)
		if err != nil {
			respondWithInternalServerProblem(c, "Failed to check slug: "+err.Error())

			return
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

	existing, err := s.organizerRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, organizerNotFound)

			return
		}

		respondWithInternalServerProblem(c, organizerGetErrorPrefix+err.Error())

		return
	}

	if !s.canEditOrganizer(c, existing) {
		respondWithForbiddenProblem(c, "You can only update your own organizer profile")

		return
	}

	fileHeader, _ := c.FormFile("media_upload")
	isMultipart := fileHeader != nil || strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data")

	removeAvatar := false

	var organizer domain.Organizer

	if isMultipart {
		parsed, err := parseOrganizerFromMultipart(c)
		if err != nil {
			respondWithBadRequestProblem(c, err.Error())

			return
		}

		organizer = parsed

		if c.PostForm("remove_avatar") == "true" {
			removeAvatar = true
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

	// Preserve immutable / system fields that are not part of the edit form.
	preserveOrganizerSystemFields(&organizer, existing)

	// A newly uploaded avatar always wins over a removal request, so a file is
	// never written to disk and then discarded.
	avatarURL, err := resolveAvatarURL(
		fileHeader,
		organizer.AvatarURL,
		existing.AvatarURL,
		removeAvatar,
	)
	if err != nil {
		respondWithAvatarStorageProblem(c, err)

		return
	}

	organizer.AvatarURL = avatarURL

	if err := validateOrganizerProfileURLs(organizer.SlackProfileURL, organizer.MediaWikiProfileURL); err != nil {
		respondWithBadRequestProblem(c, err.Error())

		return
	}

	if err := validateFirstYearAsOrganizer(organizer.FirstYearAsOrganizer); err != nil {
		respondWithBadRequestProblem(c, err.Error())

		return
	}

	if err := s.organizerRepo.Update(c.Request.Context(), &organizer); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, organizerNotFound)

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
			respondWithNotFoundProblem(c, organizerNotFound)

			return
		}

		respondWithInternalServerProblem(c, "Failed to delete organizer: "+err.Error())

		return
	}

	slog.Info("Delete Organizer: Successfully deleted organizer", "id", id)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

var slackProfileURLPattern = regexp.MustCompile(`^https://[a-zA-Z0-9-]+\.slack\.com/team/[A-Za-z0-9]+$`)

func validateOrganizerProfileURLs(slackURL, wikiURL string) error {
	if slackURL != "" {
		if !slackProfileURLPattern.MatchString(slackURL) {
			return errors.New("invalid slack_profile_url: expected https://<workspace>.slack.com/team/<id>")
		}
	}

	if wikiURL != "" {
		parsed, err := url.Parse(wikiURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return errors.New("invalid mediawiki_profile_url: must be a valid http(s) URL")
		}

		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return errors.New("invalid mediawiki_profile_url: scheme must be http or https")
		}
	}

	return nil
}

func validateFirstYearAsOrganizer(year *int) error {
	if year == nil {
		return nil
	}

	currentYear := time.Now().Year()
	if *year < 1980 || *year > currentYear+1 {
		return fmt.Errorf("first_year_as_organizer must be between 1980 and %d", currentYear+1)
	}

	return nil
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
