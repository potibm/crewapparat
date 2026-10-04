package hub

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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

	c.JSON(http.StatusOK, toMeResponse(organizer, s.isAdmin(c, userID.(string))))
}

func (s *Server) isAdmin(c *gin.Context, userID string) bool {
	if groups, exists := c.Get("userGroups"); exists {
		for _, g := range groups.([]string) {
			if g == s.cfg.Auth.OidcAdminGroup {
				return true
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
