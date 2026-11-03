package hub

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// codeOfConductExemptPaths stay reachable without accepting the current Code
// of Conduct: the frontend needs them to render the acceptance screen.
var codeOfConductExemptPaths = map[string]bool{
	"/api/admin/me":                        true,
	"/api/admin/me/link":                   true,
	"/api/admin/me/code-of-conduct/accept": true,
}

// requireCodeOfConductAcceptance blocks all admin endpoints until the current
// user accepted the latest published Code of Conduct. Admins are not exempt:
// they are organizers too.
func (s *Server) requireCodeOfConductAcceptance() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions || codeOfConductExemptPaths[c.Request.URL.Path] {
			c.Next()

			return
		}

		if s.organizerRepo == nil || s.codeOfConductRepo == nil || s.acceptanceRepo == nil {
			c.Next()

			return
		}

		userIDVal, exists := c.Get("userID")
		if !exists {
			c.Next()

			return
		}

		userID, _ := userIDVal.(string)

		organizer, err := s.organizerRepo.GetByOidcSub(c.Request.Context(), userID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				respondWithInternalServerProblem(c, "Failed to get organizer: "+err.Error())

				return
			}

			// Unlinked users cannot accept anything; the handlers answer for
			// them and the frontend shows nothing useful anyway.
			c.Next()

			return
		}

		coc, err := s.codeOfConductRepo.GetLatestPublished(c.Request.Context())
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				respondWithInternalServerProblem(c, "Failed to get code of conduct: "+err.Error())

				return
			}

			// Without a published version there is nothing to accept.
			c.Next()

			return
		}

		_, err = s.acceptanceRepo.GetByOrganizerAndCodeOfConduct(c.Request.Context(), organizer.ID, coc.ID)
		if err == nil {
			c.Next()

			return
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithInternalServerProblem(c, "Failed to check code of conduct acceptance: "+err.Error())

			return
		}

		respondWithCodeOfConductRequiredProblem(c)
	}
}
