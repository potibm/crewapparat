package hub

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/potibm/crewapparat/internal/app/domain"
)

type meResponse struct {
	ID        string `json:"id"`
	Handle    string `json:"handle"`
	Realname  string `json:"realname"`
	Email     string `json:"email"`
	Pronouns  string `json:"pronouns,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	IsAdmin   bool   `json:"is_admin"`

	// CodeOfConduct is the currently published version and
	// CodeOfConductAcceptedAt is the moment this user accepted it. Both are
	// null when there is no published version or no acceptance yet.
	CodeOfConduct           *domain.CodeOfConduct `json:"code_of_conduct"`
	CodeOfConductAcceptedAt *time.Time            `json:"code_of_conduct_accepted_at"`
}

func (s *Server) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		respondWithUnauthorizedProblem(c, "User ID not found in context")

		return
	}

	organizer, err := s.organizerRepo.GetByOidcSub(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Organizer not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get organizer: "+err.Error())

		return
	}

	response := toMeResponse(organizer, s.isAdmin(c, userID.(string)))

	coc, acceptedAt, err := s.myCodeOfConductStatus(c, organizer.ID)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to get code of conduct: "+err.Error())

		return
	}

	response.CodeOfConduct = coc
	response.CodeOfConductAcceptedAt = acceptedAt

	c.JSON(http.StatusOK, response)
}

// myCodeOfConductStatus returns the current published version together with
// the acceptance of that version for the given organizer.
func (s *Server) myCodeOfConductStatus(
	c *gin.Context,
	organizerID uuid.UUID,
) (*domain.CodeOfConduct, *time.Time, error) {
	coc, err := s.codeOfConductRepo.GetLatestPublished(c.Request.Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}

		return nil, nil, err
	}

	acceptance, err := s.acceptanceRepo.GetByOrganizerAndCodeOfConduct(
		c.Request.Context(),
		organizerID,
		coc.ID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return coc, nil, nil
		}

		return nil, nil, err
	}

	acceptedAt := acceptance.AcceptedAt

	return coc, &acceptedAt, nil
}

// LinkMe implements the account-linking step from concept §4.1.5: a user whose
// organizer record exists (created via invite approval) but has no oidc_sub yet
// is linked to their IdP identity on first login. GET /me stays read-only;
// this POST is the only mutating entry point.
//
// Idempotent: an already-linked organizer with the same sub returns 200.
func (s *Server) LinkMe(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		respondWithUnauthorizedProblem(c, "User ID not found in context")

		return
	}

	userID := userIDVal.(string)

	emailVal, _ := c.Get("userEmail")

	email, _ := emailVal.(string)
	if email == "" {
		respondWithBadRequestProblem(c, "Email claim missing in token")

		return
	}

	if verifiedVal, ok := c.Get("userEmailVerified"); ok {
		if verified, ok := verifiedVal.(*bool); ok && verified != nil && !*verified {
			respondWithForbiddenProblem(c, "Email is not verified")

			return
		}
	}

	organizer, err := s.organizerRepo.GetByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithNotFoundProblem(c, "Organizer not found")

			return
		}

		respondWithInternalServerProblem(c, "Failed to get organizer: "+err.Error())

		return
	}

	if organizer.OidcSub != nil {
		if *organizer.OidcSub == userID {
			c.JSON(http.StatusOK, toMeResponse(organizer, s.isAdmin(c, userID)))

			return
		}

		respondWithConflictProblem(c, "Organizer is already linked to a different account")

		return
	}

	sub := userID
	organizer.OidcSub = &sub

	if err := s.organizerRepo.Update(c.Request.Context(), organizer); err != nil {
		respondWithInternalServerProblem(c, "Failed to link organizer: "+err.Error())

		return
	}

	slog.Info("LinkMe: linked organizer to user", "organizer_id", organizer.ID, "user_id", userID)
	c.JSON(http.StatusOK, toMeResponse(organizer, s.isAdmin(c, userID)))
}

func (s *Server) isAdmin(c *gin.Context, userID string) bool {
	if s.cfg.Auth == nil {
		return false
	}

	if groupsVal, exists := c.Get("userGroups"); exists {
		if groups, ok := groupsVal.([]string); ok {
			for _, g := range groups {
				if g == s.cfg.Auth.OidcAdminGroup {
					return true
				}
			}
		}
	}

	for _, admin := range s.cfg.Auth.ProxyAdmins {
		if admin == userID {
			return true
		}
	}

	return false
}

func toMeResponse(organizer *domain.Organizer, isAdmin bool) meResponse {
	return meResponse{
		ID:        organizer.ID.String(),
		Handle:    organizer.Handle,
		Realname:  organizer.Realname,
		Email:     organizer.Email,
		Pronouns:  organizer.Pronouns,
		AvatarURL: organizer.AvatarURL,
		IsAdmin:   isAdmin,
	}
}
