package hub

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/potibm/crewapparat/internal/app/config"
	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
	store "github.com/potibm/crewapparat/internal/app/store/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEnv struct {
	server *Server
	store  *store.Store
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	st, err := store.NewSqliteInMemoryStore()
	require.NoError(t, err)
	require.NoError(t, st.PurgeAll())

	server, err := NewServer(Config{
		Cfg:           config.Config{},
		OrganizerRepo: st.NewOrganizerRepository(),
	})
	require.NoError(t, err)

	t.Cleanup(func() { _ = st.Close() })

	return &testEnv{server: server, store: st}
}

func setupMeRouter(env *testEnv, userID, email string, verified *bool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	admin := r.Group("/api/admin")
	admin.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("userEmail", email)
		c.Set("userEmailVerified", verified)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, domain.UserIDKey, userID)
		ctx = context.WithValue(ctx, domain.UserEmailKey, email)
		ctx = context.WithValue(ctx, domain.UserEmailVerifiedKey, verified)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	admin.GET("/me", env.server.GetMe)
	admin.POST("/me/link", env.server.LinkMe)

	return r
}

func seedOrganizer(t *testing.T, env *testEnv, email string, sub *string) {
	t.Helper()

	org := &domain.Organizer{
		Email:    email,
		Handle:   "h",
		Realname: "r",
		OidcSub:  sub,
	}
	require.NoError(t, env.store.NewOrganizerRepository().Create(context.Background(), org))
}

func TestLinkMe_HappyPath(t *testing.T) {
	env := newTestEnv(t)

	const userID = "user-123"

	seedOrganizer(t, env, "user@example.com", nil)

	r := setupMeRouter(env, userID, "user@example.com", nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/admin/me/link", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "user@example.com", body["email"])

	// Verify sub was persisted.
	orgs, _, err := env.store.NewOrganizerRepository().
		List(context.Background(), repository.OrganizerListParams{}, repository.OrganizerListFilters{})
	require.NoError(t, err)
	require.Len(t, orgs, 1)
	require.NotNil(t, orgs[0].OidcSub)
	assert.Equal(t, userID, *orgs[0].OidcSub)
}

func TestLinkMe_Idempotent(t *testing.T) {
	env := newTestEnv(t)

	const userID = "user-123"

	sub := userID
	seedOrganizer(t, env, "user@example.com", &sub)

	r := setupMeRouter(env, userID, "user@example.com", nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/admin/me/link", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLinkMe_Conflict(t *testing.T) {
	env := newTestEnv(t)

	const userID = "user-123"

	other := "other-456"
	seedOrganizer(t, env, "user@example.com", &other)

	r := setupMeRouter(env, userID, "user@example.com", nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/admin/me/link", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestLinkMe_NotFound(t *testing.T) {
	env := newTestEnv(t)
	r := setupMeRouter(env, "user-123", "nobody@example.com", nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/admin/me/link", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLinkMe_MissingEmail(t *testing.T) {
	env := newTestEnv(t)
	r := setupMeRouter(env, "user-123", "", nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/admin/me/link", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLinkMe_EmailNotVerified(t *testing.T) {
	env := newTestEnv(t)
	f := false
	r := setupMeRouter(env, "user-123", "user@example.com", &f)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/admin/me/link", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetMe_Unlinked(t *testing.T) {
	env := newTestEnv(t)
	r := setupMeRouter(env, "user-123", "user@example.com", nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/admin/me", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
