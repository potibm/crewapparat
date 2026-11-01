package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondWithInternalServerProblem_DoesNotLeakDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/organizers", http.NoBody)

	respondWithInternalServerProblem(c, "sql: no such table: organizers")

	var body ProblemDetails
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))

	assert.Equal(t, http.StatusInternalServerError, body.Status)
	assert.Equal(t, "Internal Server Error", body.Title)
	assert.Equal(t, "An internal server error occurred", body.Detail)
	assert.NotContains(t, body.Detail, "sql")
	assert.NotContains(t, body.Detail, "organizers")
}

func TestRespondWithAvatarStorageProblem_UserErrorsAreBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, err := range []error{ErrAvatarTooLarge, ErrAvatarUnsupported} {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/organizers", http.NoBody)

		respondWithAvatarStorageProblem(c, err)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	}
}

func TestRespondWithAvatarStorageProblem_InternalErrorsStayGeneric(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/organizers", http.NoBody)

	respondWithAvatarStorageProblem(c, assert.AnError)

	var body ProblemDetails
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))

	assert.Equal(t, http.StatusInternalServerError, body.Status)
	assert.Equal(t, "An internal server error occurred", body.Detail)
}
