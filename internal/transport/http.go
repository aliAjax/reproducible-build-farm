package transport

import (
	"encoding/json"
	"example.com/reproducible-build-farm/internal/application"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/api"
	"io"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	App *application.Service
	Mux *http.ServeMux
}

func New(app *application.Service) *Server {
	s := &Server{App: app, Mux: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) routes() {
	s.Mux.HandleFunc("/healthz", s.health)
	s.Mux.HandleFunc("/readyz", s.health)
	s.Mux.HandleFunc("/api/v1/build-definitions", s.definitions)
	s.Mux.HandleFunc("/api/v1/executions", s.executions)
	s.Mux.HandleFunc("/api/v1/executions/", s.execution)
	s.Mux.HandleFunc("/api/v1/projects", s.projects)
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
func (s *Server) projects(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.WriteError(w, 405, "method_not_allowed", "POST required", "")
		return
	}
	var p domain.Project
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&p); err != nil {
		api.WriteError(w, 400, "invalid_json", err.Error(), "")
		return
	}
	if p.ID == "" {
		p.ID = "project-" + time.Now().Format("150405.000")
	}
	p.CreatedAt = time.Now().UTC()
	if err := s.App.Store.CreateProject(r.Context(), p); err != nil {
		api.WriteError(w, 409, "conflict", err.Error(), "")
		return
	}
	json.NewEncoder(w).Encode(p)
}
func (s *Server) definitions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.WriteError(w, 405, "method_not_allowed", "POST required", "")
		return
	}
	var req struct {
		ID, ProjectID string
		DSL           json.RawMessage
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 2<<20)).Decode(&req); err != nil {
		api.WriteError(w, 400, "invalid_json", err.Error(), "")
		return
	}
	d, err := s.App.CreateDefinition(r.Context(), req.ProjectID, req.ID, req.DSL)
	if err != nil {
		api.WriteError(w, 400, "invalid_definition", err.Error(), "")
		return
	}
	json.NewEncoder(w).Encode(d)
}
func (s *Server) executions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.WriteError(w, 405, "method_not_allowed", "POST required", "")
		return
	}
	var req struct{ ID, DefinitionID, IdempotencyKey string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, 400, "invalid_json", err.Error(), "")
		return
	}
	e, err := s.App.Submit(r.Context(), req.ID, req.DefinitionID, req.IdempotencyKey)
	if err != nil {
		api.WriteError(w, 400, "submit_failed", err.Error(), "")
		return
	}
	w.WriteHeader(202)
	json.NewEncoder(w).Encode(e)
}
func (s *Server) execution(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/executions/")
	if strings.HasSuffix(id, "/attestation") {
		id = strings.TrimSuffix(id, "/attestation")
		e, err := s.App.GetExecution(r.Context(), id)
		if err != nil {
			api.WriteError(w, 404, "not_found", err.Error(), "")
			return
		}
		a, err := s.App.GetAttestation(r.Context(), e.AttestationID)
		if err != nil {
			api.WriteError(w, 404, "not_found", err.Error(), "")
			return
		}
		json.NewEncoder(w).Encode(a)
		return
	}
	e, err := s.App.GetExecution(r.Context(), id)
	if err != nil {
		api.WriteError(w, 404, "not_found", err.Error(), "")
		return
	}
	json.NewEncoder(w).Encode(e)
}
