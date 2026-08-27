package transport

import (
	"encoding/json"
	"example.com/reproducible-build-farm/internal/domain"
	"fmt"
	"net/http"
	"time"
)

type executionResponse struct {
	ID            string                `json:"id"`
	State         domain.ExecutionState `json:"state"`
	ActionKey     string                `json:"action_key"`
	ResultDigest  string                `json:"result_digest,omitempty"`
	AttestationID string                `json:"attestation_id,omitempty"`
	Error         string                `json:"error,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
	FinishedAt    *time.Time            `json:"finished_at,omitempty"`
}

func writeExecution(w http.ResponseWriter, e domain.Execution) {
	var finish *time.Time
	if !e.FinishedAt.IsZero() {
		finish = &e.FinishedAt
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(executionResponse{ID: e.ID, State: e.State, ActionKey: e.ActionKey.String(), ResultDigest: e.ResultDigest.String(), AttestationID: e.AttestationID, Error: e.Error, CreatedAt: e.CreatedAt, FinishedAt: finish})
}
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func method(w http.ResponseWriter, r *http.Request, expected string) bool {
	if r.Method == expected {
		return true
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method must be " + expected})
	return false
}
func parseLimit(r *http.Request) int {
	n := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
			n = 50
		}
	}
	if n < 1 {
		n = 1
	}
	if n > 500 {
		n = 500
	}
	return n
}
