package transport

import (
	"context"
	"encoding/json"
	"example.com/reproducible-build-farm/internal/application"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/observability"
	"example.com/reproducible-build-farm/pkg/api"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	App    *application.Service
	Mux    *http.ServeMux
	Logger *observability.Logger
}

func New(app *application.Service) *Server {
	s := &Server{App: app, Mux: http.NewServeMux(), Logger: observability.NewLogger()}
	s.routes()
	return s
}

// Handler returns the root handler with request middleware applied.
// ServeMux is exposed for tests and direct wiring, but production should
// route through Handler so request IDs, access logs and panic recovery run.
func (s *Server) Handler() http.Handler {
	h := http.Handler(s.Mux)
	h = s.recoverPanic(h)
	h = s.accessLog(h)
	h = s.requestID(h)
	return h
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
		s.fail(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "POST required")
		return
	}
	var p domain.Project
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&p); err != nil {
		s.fail(w, r, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if p.ID == "" {
		p.ID = "project-" + time.Now().Format("150405.000")
	}
	p.CreatedAt = time.Now().UTC()
	if err := s.App.Store.CreateProject(r.Context(), p); err != nil {
		status := http.StatusConflict
		code := "conflict"
		if !errors.Is(err, domain.ErrConflict) {
			status = http.StatusBadRequest
			code = "invalid_project"
		}
		s.fail(w, r, status, code, err.Error())
		return
	}
	s.write(w, http.StatusOK, p)
}

func (s *Server) definitions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.fail(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "POST required")
		return
	}
	var req struct {
		ID, ProjectID string
		DSL           json.RawMessage
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 2<<20)).Decode(&req); err != nil {
		s.fail(w, r, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	d, err := s.App.CreateDefinition(r.Context(), req.ProjectID, req.ID, req.DSL)
	if err != nil {
		s.fail(w, r, http.StatusBadRequest, "invalid_definition", err.Error())
		return
	}
	s.write(w, http.StatusOK, d)
}

func (s *Server) executions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.fail(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "POST required")
		return
	}
	var req struct{ ID, DefinitionID, IdempotencyKey string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.fail(w, r, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	e, err := s.App.Submit(r.Context(), req.ID, req.DefinitionID, req.IdempotencyKey)
	if err != nil {
		// Distinguish "definition does not exist" from a genuine submit failure so
		// callers get a meaningful status and operators get a log line they can trace.
		switch {
		case errors.Is(err, domain.ErrNotFound):
			s.fail(w, r, http.StatusNotFound, "definition_not_found", err.Error())
		case errors.Is(err, domain.ErrInvalid):
			s.fail(w, r, http.StatusBadRequest, "invalid_execution", err.Error())
		default:
			s.fail(w, r, http.StatusBadRequest, "submit_failed", err.Error())
		}
		return
	}
	s.Logger.Info("execution submitted", map[string]interface{}{
		"execution_id":  e.ID,
		"definition_id": req.DefinitionID,
	})
	writeExecution(w, e)
}

func (s *Server) execution(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/executions/")
	if strings.HasSuffix(id, "/attestation") {
		id = strings.TrimSuffix(id, "/attestation")
		e, err := s.App.GetExecution(r.Context(), id)
		if err != nil {
			s.fail(w, r, http.StatusNotFound, "not_found", err.Error())
			return
		}
		a, err := s.App.GetAttestation(r.Context(), e.AttestationID)
		if err != nil {
			s.fail(w, r, http.StatusNotFound, "not_found", err.Error())
			return
		}
		s.write(w, http.StatusOK, a)
		return
	}
	e, err := s.App.GetExecution(r.Context(), id)
	if err != nil {
		s.fail(w, r, http.StatusNotFound, "not_found", err.Error())
		return
	}
	s.write(w, http.StatusOK, e)
}

// requestID propagates an incoming X-Request-ID (generating one if absent) and
// stashes it on the request context so handlers can attach it to errors/logs.
func (s *Server) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = randID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, withRequestID(r, id))
	})
}

func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		s.Logger.Info("http", map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      rw.status,
			"duration_ms": time.Since(start).Milliseconds(),
			"request_id":  requestIDOf(r),
		})
	})
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.Logger.Error("panic", map[string]interface{}{
					"request_id": requestIDOf(r),
					"recover":    rec,
				})
				api.WriteError(w, http.StatusInternalServerError, "internal_error", "internal error", requestIDOf(r))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// fail logs the error with request context and writes a structured JSON error
// response that carries the request id, so operators can correlate a failing
// submit back to its log line.
func (s *Server) fail(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	rid := requestIDOf(r)
	s.Logger.Error("request failed", map[string]interface{}{
		"status":     status,
		"code":       code,
		"message":    message,
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": rid,
	})
	api.WriteError(w, status, code, message, rid)
}

func (s *Server) write(w http.ResponseWriter, status int, v interface{}) {
	writeJSON(w, status, v)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

type ctxKey struct{}

func withRequestID(r *http.Request, id string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, id))
}

func requestIDOf(r *http.Request) string {
	if v, ok := r.Context().Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}
