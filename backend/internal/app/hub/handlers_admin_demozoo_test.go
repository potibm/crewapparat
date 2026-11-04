package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/potibm/crewapparat/internal/app/demozoo"
)

func newDemozooTestServer() *Server {
	return &Server{
		demozooSceners: []demozoo.Scener{
			{ID: 1, Name: "Alpha", CountryCode: "DE", Groups: []string{"G1"}},
			{ID: 2, Name: "Beta", CountryCode: "US", Groups: nil},
			{ID: 6, Name: "Six", CountryCode: "", Groups: []string{}},
			{ID: 10, Name: "Alphabet", CountryCode: "FR", Groups: []string{"G2"}},
		},
	}
}

func TestListDemozooSceners_FilterAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := newDemozooTestServer()
	r := gin.New()
	r.GET("/demozoo-sceners", s.ListDemozooSceners)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/demozoo-sceners?q=alp&_start=0&_end=10", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	if w.Header().Get("X-Total-Count") != "2" {
		t.Fatalf("total=%s", w.Header().Get("X-Total-Count"))
	}

	var got []demozoo.Scener
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	// "Alpha..." should come before "Alphabet" only if starts-with priority;
	// both start with "alp", so alphabetical: Alpha < Alphabet.
	if got[0].Name != "Alpha" || got[1].Name != "Alphabet" {
		t.Fatalf("order=%v", got)
	}
}

func TestListDemozooSceners_NumericExactID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := newDemozooTestServer()
	r := gin.New()
	r.GET("/demozoo-sceners", s.ListDemozooSceners)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/demozoo-sceners?q=6", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}

	var got []demozoo.Scener
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || got[0].ID != 6 {
		t.Fatalf("got=%+v", got)
	}
}

func TestListDemozooSceners_GetManyByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := newDemozooTestServer()
	r := gin.New()
	r.GET("/demozoo-sceners", s.ListDemozooSceners)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/demozoo-sceners?id=2&id=10", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}

	var got []demozoo.Scener
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
}

func TestGetDemozooScener(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := newDemozooTestServer()
	r := gin.New()
	r.GET("/demozoo-sceners/:id", s.GetDemozooScener)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/demozoo-sceners/6", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var got demozoo.Scener
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if got.ID != 6 || got.Name != "Six" {
		t.Fatalf("got=%+v", got)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/demozoo-sceners/999", http.NoBody)
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Fatalf("status=%d", w2.Code)
	}
}
