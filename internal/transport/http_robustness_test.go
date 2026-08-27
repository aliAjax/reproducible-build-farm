package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/reproducible-build-farm/internal/application"
	"example.com/reproducible-build-farm/internal/cache"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/infrastructure"
	"example.com/reproducible-build-farm/internal/repository"
)

func TestMalformedBodyNo500(t *testing.T) {
	store := repository.NewMemory()
	app := application.New(store, cache.NewMemory(1000), infrastructure.NewSimulatedExecutor())
	srv := New(app)
	body, _ := json.Marshal(map[string]interface{}{"id": "d1", "projectID": "p1", "DSL": map[string]interface{}{"name": "x", "toolchain_id": "tc"}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/build-definitions", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed definition body must be rejected with 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestExecutionResponseOmitsZeroFinishedAt(t *testing.T) {
	rec := httptest.NewRecorder()
	writeExecution(rec, domain.Execution{ID: "ex-1", State: domain.StateQueued})
	if strings.Contains(rec.Body.String(), "finished_at") {
		t.Fatalf("unfinished execution must not report a finished_at timestamp: %s", rec.Body.String())
	}
}

func TestRecoverExposesPanicMessage(t *testing.T) {
	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom detail")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "boom detail") {
		t.Fatalf("recover must expose the panic detail, got %q", rec.Body.String())
	}
}
