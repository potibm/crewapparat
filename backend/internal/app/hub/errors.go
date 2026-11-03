package hub

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProblemDetails struct {
	Type     string `json:"type"`               // URI to error documentation (often "about:blank")
	Title    string `json:"title"`              // Short, human-readable summary
	Status   int    `json:"status"`             // The HTTP status code
	Detail   string `json:"detail,omitempty"`   // Specific info about this error
	Instance string `json:"instance,omitempty"` // URI of the specific request
}

func respondWithProblem(c *gin.Context, status int, title, detail string) {
	c.Header("Content-Type", "application/problem+json")

	c.AbortWithStatusJSON(status, ProblemDetails{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: c.Request.RequestURI,
	})
}

func respondWithInvalidIDFormatProblem(c *gin.Context) {
	respondWithProblem(c, http.StatusBadRequest, "Bad Request", "Invalid ID format")
}

// respondWithInternalServerProblem logs the detail - which may contain internal
// error information such as database messages - and returns a generic problem
// response so that implementation details never leak to the client.
func respondWithInternalServerProblem(c *gin.Context, detail string) {
	slog.Error("internal server error",
		"detail", detail,
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)

	respondWithProblem(c, http.StatusInternalServerError, "Internal Server Error", "An internal server error occurred")
}

func respondWithNotFoundProblem(c *gin.Context, detail string) {
	respondWithProblem(c, http.StatusNotFound, "Not Found", detail)
}

func respondWithFailedToParsePayloadProblem(c *gin.Context, err error) {
	respondWithBadRequestProblem(c, "Failed to parse payload: "+err.Error())
}

func respondWithBadRequestProblem(c *gin.Context, detail string) {
	respondWithProblem(c, http.StatusBadRequest, "Bad Request", detail)
}

func respondWithUnauthorizedProblem(c *gin.Context, detail string) {
	respondWithProblem(c, http.StatusUnauthorized, "Unauthorized", detail)
}

func respondWithForbiddenProblem(c *gin.Context, detail string) {
	respondWithProblem(c, http.StatusForbidden, "Forbidden", detail)
}

func respondWithConflictProblem(c *gin.Context, detail string) {
	respondWithProblem(c, http.StatusConflict, "Conflict", detail)
}

func respondWithCodeOfConductRequiredProblem(c *gin.Context) {
	respondWithProblem(
		c,
		http.StatusPreconditionRequired,
		"Code of Conduct acceptance required",
		"Accept the current Code of Conduct to continue",
	)
}
