package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/reproducible-build-farm/internal/application"
	"example.com/reproducible-build-farm/internal/cache"
	"example.com/reproducible-build-farm/internal/infrastructure"
	"example.com/reproducible-build-farm/internal/repository"
)

func TestSubmitUnknownDefinitionReturns404(t *testing.T) {
	store := repository.NewMemory()
	app := application.New(store, cache.NewMemory(1000), infrastructure.NewSimulatedExecutor())
	srv := New(app)
	body, _ := json.Marshal(map[string]string{"id": "ex-1", "definitionID": "missing-def", "idempotencyKey": "k1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}
