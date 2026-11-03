package hub

import (
	"bytes"
	"context"
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

func TestUpdateCodeOfConduct_PublishedIsImmutable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st, err := store.NewSqliteInMemoryStore()
	require.NoError(t, err)

	t.Cleanup(func() { _ = st.Close() })

	require.NoError(t, st.PurgeAll())

	repo := st.NewCodeOfConductRepository()

	publishedAt := time.Now().Add(-time.Hour)
	published := &domain.CodeOfConduct{
		Status:      domain.CodeOfConductStatusPublished,
		Text:        "# Code of Conduct\n\noriginal",
		PublishedAt: &publishedAt,
	}
	require.NoError(t, repo.Create(context.Background(), published))

	pending := &domain.CodeOfConduct{
		Status: domain.CodeOfConductStatusPending,
		Text:   "draft",
	}
	require.NoError(t, repo.Create(context.Background(), pending))

	deletable := &domain.CodeOfConduct{
		Status: domain.CodeOfConductStatusPending,
		Text:   "scratch",
	}
	require.NoError(t, repo.Create(context.Background(), deletable))

	server, err := NewServer(Config{
		Cfg:               config.Config{Auth: &config.AuthConfig{ProxyAdmins: []string{"tester"}}},
		CodeOfConductRepo: repo,
	})
	require.NoError(t, err)

	r := gin.New()
	admin := r.Group("/api/admin")
	admin.Use(func(c *gin.Context) {
		c.Set("userID", "tester")
		c.Next()
	})
	admin.PUT("/code-of-conducts/:id", server.UpdateCodeOfConduct)
	admin.DELETE("/code-of-conducts/:id", server.DeleteCodeOfConduct)

	update := func(id, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req, reqErr := http.NewRequest(
			http.MethodPut,
			"/api/admin/code-of-conducts/"+id,
			bytes.NewBufferString(body),
		)
		require.NoError(t, reqErr)
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		return w
	}

	del := func(id string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req, reqErr := http.NewRequest(http.MethodDelete, "/api/admin/code-of-conducts/"+id, http.NoBody)
		require.NoError(t, reqErr)

		r.ServeHTTP(w, req)

		return w
	}

	t.Run("rejects updates of a published code of conduct", func(t *testing.T) {
		w := update(published.ID.String(), `{"status":"published","text":"tampered"}`)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "immutable")

		stored, getErr := repo.GetByID(context.Background(), published.ID)
		require.NoError(t, getErr)
		assert.Equal(t, "# Code of Conduct\n\noriginal", stored.Text)
	})

	t.Run("allows updates of a pending code of conduct", func(t *testing.T) {
		w := update(pending.ID.String(), `{"status":"pending","text":"improved draft"}`)

		assert.Equal(t, http.StatusOK, w.Code)

		stored, getErr := repo.GetByID(context.Background(), pending.ID)
		require.NoError(t, getErr)
		assert.Equal(t, "improved draft", stored.Text)
	})

	t.Run("publishing is allowed once and then locked", func(t *testing.T) {
		w := update(pending.ID.String(), `{"status":"published","text":"improved draft"}`)
		assert.Equal(t, http.StatusOK, w.Code)

		stored, getErr := repo.GetByID(context.Background(), pending.ID)
		require.NoError(t, getErr)
		require.NotNil(t, stored.PublishedAt)

		w = update(pending.ID.String(), `{"status":"published","text":"changed again"}`)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("rejects deleting a published code of conduct", func(t *testing.T) {
		w := del(published.ID.String())

		assert.Equal(t, http.StatusConflict, w.Code)

		_, getErr := repo.GetByID(context.Background(), published.ID)
		assert.NoError(t, getErr)
	})

	t.Run("allows deleting a pending code of conduct", func(t *testing.T) {
		w := del(deletable.ID.String())

		assert.Equal(t, http.StatusOK, w.Code)

		_, getErr := repo.GetByID(context.Background(), deletable.ID)
		assert.Error(t, getErr)
	})
}
