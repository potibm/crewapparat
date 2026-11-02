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
	store "github.com/potibm/crewapparat/internal/app/store/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListOrganizers_StatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st, err := store.NewSqliteInMemoryStore()
	require.NoError(t, err)

	defer st.Close()

	require.NoError(t, st.PurgeAll())

	orgRepo := st.NewOrganizerRepository()
	partRepo := st.NewParticipationRepository()

	org1 := &domain.Organizer{Handle: "org1", Email: "org1@test.com"}
	org2 := &domain.Organizer{Handle: "org2", Email: "org2@test.com"}

	require.NoError(t, orgRepo.Create(context.Background(), org1))
	require.NoError(t, orgRepo.Create(context.Background(), org2))

	p1 := &domain.Participation{
		OrganizerID: org1.ID,
		EventYearID: "2026",
		Status:      domain.ParticipationStatusConfirmed,
	}
	p2 := &domain.Participation{
		OrganizerID: org2.ID,
		EventYearID: "2026",
		Status:      domain.ParticipationStatusPending,
	}

	require.NoError(t, partRepo.Create(context.Background(), p1))
	require.NoError(t, partRepo.Create(context.Background(), p2))

	server, err := NewServer(Config{Cfg: config.Config{}, OrganizerRepo: orgRepo})
	require.NoError(t, err)

	r := gin.New()
	admin := r.Group("/api/admin")
	admin.GET("/organizers", server.ListOrganizers)

	cases := []struct {
		status   string
		expected int
		handle   string
	}{
		{"confirmed", 1, "org1"},
		{"all", 2, ""},
		{"pending", 1, "org2"},
		{"declined", 0, ""},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		url := "/api/admin/organizers?event_year_id=2026&status=" + tc.status + "&_start=0&_end=10&_sort=id&_order=ASC"
		req, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "status=%s", tc.status)

		var result []map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		assert.Len(t, result, tc.expected, "status=%s", tc.status)

		if tc.handle != "" {
			assert.Equal(t, tc.handle, result[0]["handle"], "status=%s", tc.status)
		}
	}
}
