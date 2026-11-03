package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/config"
	"github.com/potibm/crewapparat/internal/app/domain"
	store "github.com/potibm/crewapparat/internal/app/store/gorm"
)

type codeOfConductTestSetup struct {
	server    *Server
	router    *gin.Engine
	organizer *domain.Organizer
	published *domain.CodeOfConduct
	pending   *domain.CodeOfConduct
	store     *store.Store
}

func newCodeOfConductTestSetup(t *testing.T) *codeOfConductTestSetup {
	t.Helper()

	gin.SetMode(gin.TestMode)

	st, err := store.NewSqliteInMemoryStore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	require.NoError(t, st.PurgeAll())

	organizerRepo := st.NewOrganizerRepository()
	cocRepo := st.NewCodeOfConductRepository()
	acceptanceRepo := st.NewCodeOfConductAcceptanceRepository()

	sub := "tester-sub"
	organizer := &domain.Organizer{
		OidcSub:  &sub,
		Handle:   "tester",
		Realname: "Test Organizer",
		Email:    "tester@example.com",
	}
	require.NoError(t, organizerRepo.Create(context.Background(), organizer))

	publishedAt := time.Now().Add(-time.Hour)
	published := &domain.CodeOfConduct{
		Version:     "1.0.0",
		Status:      domain.CodeOfConductStatusPublished,
		Text:        "# Code of Conduct\n\nBe nice.",
		PublishedAt: &publishedAt,
	}
	require.NoError(t, cocRepo.Create(context.Background(), published))

	pending := &domain.CodeOfConduct{
		Version: "1.1.0",
		Status:  domain.CodeOfConductStatusPending,
		Text:    "# Draft",
	}
	require.NoError(t, cocRepo.Create(context.Background(), pending))

	server, err := NewServer(Config{
		Cfg:               config.Config{Auth: &config.AuthConfig{ProxyAdmins: []string{"tester-sub"}}},
		OrganizerRepo:     organizerRepo,
		CodeOfConductRepo: cocRepo,
		AcceptanceRepo:    acceptanceRepo,
	})
	require.NoError(t, err)

	r := gin.New()
	admin := r.Group("/api/admin")
	admin.Use(func(c *gin.Context) {
		if user := c.GetHeader("X-Test-User"); user != "" {
			c.Set("userID", user)
		}

		c.Next()
	})
	admin.Use(server.requireCodeOfConductAcceptance())

	admin.GET("/me", server.GetMe)
	admin.POST("/me/code-of-conduct/accept", server.AcceptCodeOfConduct)
	admin.GET("/code-of-conduct-acceptances", server.ListCodeOfConductAcceptances)
	admin.POST("/code-of-conducts", server.CreateCodeOfConduct)
	admin.GET("/organizers", func(c *gin.Context) { c.Status(http.StatusOK) })

	return &codeOfConductTestSetup{
		server:    server,
		router:    r,
		organizer: organizer,
		published: published,
		pending:   pending,
		store:     st,
	}
}

func (s *codeOfConductTestSetup) request(method, path, user, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", user)
	s.router.ServeHTTP(w, req)

	return w
}

func TestCodeOfConductAcceptance_BlocksUntilAccepted(t *testing.T) {
	s := newCodeOfConductTestSetup(t)

	w := s.request(http.MethodGet, "/api/admin/organizers", "tester-sub", "")
	assert.Equal(t, http.StatusPreconditionRequired, w.Code)

	// The status endpoint is exempt so the frontend can render the prompt.
	w = s.request(http.MethodGet, "/api/admin/me", "tester-sub", "")
	require.Equal(t, http.StatusOK, w.Code)

	var status meResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &status))
	require.NotNil(t, status.CodeOfConduct)
	assert.Equal(t, "1.0.0", status.CodeOfConduct.Version)
	assert.Nil(t, status.CodeOfConductAcceptedAt)

	w = s.request(http.MethodPost, "/api/admin/me/code-of-conduct/accept", "tester-sub",
		`{"code_of_conduct_id":"`+s.published.ID.String()+`"}`)
	require.Equal(t, http.StatusOK, w.Code)

	w = s.request(http.MethodGet, "/api/admin/organizers", "tester-sub", "")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAcceptCodeOfConduct_RejectsStaleVersion(t *testing.T) {
	s := newCodeOfConductTestSetup(t)

	w := s.request(http.MethodPost, "/api/admin/me/code-of-conduct/accept", "tester-sub",
		`{"code_of_conduct_id":"`+s.pending.ID.String()+`"}`)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestAcceptCodeOfConduct_IsIdempotent(t *testing.T) {
	s := newCodeOfConductTestSetup(t)

	body := `{"code_of_conduct_id":"` + s.published.ID.String() + `"}`

	first := s.request(http.MethodPost, "/api/admin/me/code-of-conduct/accept", "tester-sub", body)
	require.Equal(t, http.StatusOK, first.Code)

	var firstResponse map[string]any
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstResponse))

	second := s.request(http.MethodPost, "/api/admin/me/code-of-conduct/accept", "tester-sub", body)
	require.Equal(t, http.StatusOK, second.Code)

	var secondResponse map[string]any
	require.NoError(t, json.Unmarshal(second.Body.Bytes(), &secondResponse))

	assert.Equal(t, firstResponse["accepted_at"], secondResponse["accepted_at"])
}

func TestListCodeOfConductAcceptances_AdminOnly(t *testing.T) {
	s := newCodeOfConductTestSetup(t)

	body := `{"code_of_conduct_id":"` + s.published.ID.String() + `"}`
	require.Equal(t, http.StatusOK,
		s.request(http.MethodPost, "/api/admin/me/code-of-conduct/accept", "tester-sub", body).Code)

	// Not an admin: forbidden.
	w := s.request(
		http.MethodGet,
		"/api/admin/code-of-conduct-acceptances?code_of_conduct_id="+s.published.ID.String(),
		"someone-else",
		"",
	)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Admin: list with organizer data.
	w = s.request(
		http.MethodGet,
		"/api/admin/code-of-conduct-acceptances?code_of_conduct_id="+s.published.ID.String(),
		"tester-sub",
		"",
	)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "1", w.Header().Get("X-Total-Count"))

	var items []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	require.Len(t, items, 1)
	assert.Equal(t, "tester", items[0]["handle"])
	assert.Equal(t, "Test Organizer", items[0]["realname"])
}

func TestCreateCodeOfConduct_ValidatesVersion(t *testing.T) {
	s := newCodeOfConductTestSetup(t)

	// Accept first so the middleware does not interfere.
	acceptBody := `{"code_of_conduct_id":"` + s.published.ID.String() + `"}`
	require.Equal(t, http.StatusOK,
		s.request(http.MethodPost, "/api/admin/me/code-of-conduct/accept", "tester-sub", acceptBody).Code)

	w := s.request(http.MethodPost, "/api/admin/code-of-conducts", "tester-sub",
		`{"version":"not-a-version","status":"pending","text":"x"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = s.request(http.MethodPost, "/api/admin/code-of-conducts", "tester-sub",
		`{"version":"2.0.0","status":"pending","text":"x"}`)
	assert.Equal(t, http.StatusCreated, w.Code)
}
