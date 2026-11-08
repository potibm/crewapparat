package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/potibm/crewapparat/internal/app/domain"
	store "github.com/potibm/crewapparat/internal/app/store/gorm"
)

func TestListAndGetShiftPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	st, err := store.NewSqliteInMemoryStore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	require.NoError(t, st.PurgeAll())

	repo := st.NewShiftPlanRepository()
	plan := &domain.ShiftPlan{
		EventYearID: "2026",
		Version:     "1.0.0",
		Status:      domain.ShiftPlanStatusPublished,
		Slots: []domain.ShiftPlanSlot{
			{Slot: 0, Organizers: []string{}},
		},
	}
	require.NoError(t, repo.Create(t.Context(), plan))

	server, err := NewServer(Config{ShiftPlanRepo: repo})
	require.NoError(t, err)

	r := gin.New()
	admin := r.Group("/api/admin")
	admin.GET("/shift-plans", server.ListShiftPlans)
	admin.GET("/shift-plans/:id", server.GetShiftPlan)

	// List
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/admin/shift-plans", http.NoBody)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var list []domain.ShiftPlan
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list, 1)
	require.Equal(t, "1.0.0", list[0].Version)

	// Get
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/admin/shift-plans/"+plan.ID.String(), http.NoBody)
	r.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusOK, w2.Code)

	// 404
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(
		http.MethodGet,
		"/api/admin/shift-plans/00000000-0000-0000-0000-000000000000",
		http.NoBody,
	)
	r.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusNotFound, w3.Code)
}
