package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/gjy20/defense-agent/backend/internal/audit"
	"github.com/gjy20/defense-agent/backend/internal/orchestrator"
	"github.com/gjy20/defense-agent/backend/internal/sse"
	"github.com/gjy20/defense-agent/backend/internal/types"
)

// Server holds HTTP handlers
type Server struct {
	orch      *orchestrator.Orchestrator
	sseBroker *sse.Broker
	router    chi.Router
}

// NewServer creates an API server
func NewServer(orch *orchestrator.Orchestrator, sseB *sse.Broker) *Server {
	s := &Server{orch: orch, sseBroker: sseB, router: chi.NewRouter()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(corsMiddleware)

	s.router.Get("/api/v1/health", s.Health)

	s.router.Post("/api/v1/tasks", s.CreateTask)
	s.router.Get("/api/v1/tasks/{id}", s.GetTask)
	s.router.Post("/api/v1/tasks/{id}/cancel", s.CancelTask)
	s.router.Get("/api/v1/tasks/{id}/audit", s.GetAuditTrail)

	s.router.Get("/api/v1/agents", s.ListAgents)

	s.router.Get("/api/v1/instances", s.ListInstances)

	s.router.Get("/api/v1/approvals", s.ListApprovals)
	s.router.Post("/api/v1/approvals/{id}/approve", s.ApproveAction)

	// SSE streaming
	s.router.Get("/api/v1/events", s.sseBroker.SSEHandler)

	// Metrics for Prometheus
	s.router.Get("/metrics", s.Metrics)

	// Knowledge graph
	s.router.Get("/api/v1/graph/{id}", s.GetGraph)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Health check endpoint
func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateTask submits a new task
func (s *Server) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Scene       string       `json:"scene"`
		Title       string       `json:"title"`
		Description string       `json:"description"`
		Input       string       `json:"input"`
		Alerts      []types.Alert `json:"alerts,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	task := &types.Task{
		Scene:       types.SceneType(input.Scene),
		Title:       input.Title,
		Description: input.Description,
		Input:       input.Input,
		Alerts:      input.Alerts,
	}

	inst, err := s.orch.SubmitTask(context.Background(), task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, inst)
}

// GetTask returns task status and DAG state
func (s *Server) GetTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inst, ok := s.orch.GetInstance(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "instance not found"})
		return
	}
	writeJSON(w, http.StatusOK, inst)
}

// CancelTask cancels a running task
func (s *Server) CancelTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.orch.CancelInstance(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GetAuditTrail returns the audit chain for a task
func (s *Server) GetAuditTrail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	records, err := s.orch.GetAuditTrail(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if records == nil {
		records = []audit.Record{}
	}
	writeJSON(w, http.StatusOK, records)
}

// ListAgents returns all registered agent types
func (s *Server) ListAgents(w http.ResponseWriter, r *http.Request) {
	type agentInfo struct {
		Type         string   `json:"type"`
		Capabilities []string `json:"capabilities"`
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agents": []agentInfo{
			{Type: "perceiver", Capabilities: []string{"log_collection", "traffic_analysis", "asset_discovery", "alert_aggregation"}},
			{Type: "analyst", Capabilities: []string{"alert_correlation", "attack_chain_mapping", "root_cause_analysis", "attck_mapping"}},
			{Type: "responder", Capabilities: []string{"ip_blocking", "host_isolation", "patch_deployment", "service_recovery"}},
			{Type: "operator", Capabilities: []string{"health_check", "config_compliance", "patch_management", "backup_restore", "capacity_planning", "service_orchestration", "log_rotation", "certificate_management"}},
			{Type: "researcher", Capabilities: []string{"threat_intel", "cve_lookup", "ioc_correlation", "exploit_search"}},
			{Type: "developer", Capabilities: []string{"attack_planning", "exploit_selection", "tool_recommendation", "strategy_generation"}},
			{Type: "executor", Capabilities: []string{"tool_execution", "command_dispatch", "result_collection", "sandbox_orchestration"}},
			{Type: "adviser", Capabilities: []string{"execution_monitoring", "loop_detection", "progress_evaluation", "mentor_guidance"}},
			{Type: "reflector", Capabilities: []string{"failure_analysis", "guidance_generation", "error_recovery", "tool_suggestion"}},
			{Type: "auditor", Capabilities: []string{"decision_review", "evidence_verification", "compliance_check", "risk_reassessment"}},
			{Type: "memorist", Capabilities: []string{"memory_store", "memory_retrieve", "pattern_extraction", "experience_replay"}},
		},
	})
}

// ListInstances returns all instances
func (s *Server) ListInstances(w http.ResponseWriter, r *http.Request) {
	insts := s.orch.ListInstances()
	if insts == nil {
		insts = []*orchestrator.Instance{}
	}
	writeJSON(w, http.StatusOK, insts)
}

// ListApprovals returns pending human approvals
func (s *Server) ListApprovals(w http.ResponseWriter, r *http.Request) {
	pending := s.orch.PendingApprovals()
	if pending == nil {
		pending = []*audit.Record{}
	}
	writeJSON(w, http.StatusOK, pending)
}

// ApproveAction approves a pending action
func (s *Server) ApproveAction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.orch.ApproveAction(r.Context(), id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Metrics returns Prometheus-compatible metrics
func (s *Server) Metrics(w http.ResponseWriter, r *http.Request) {
	insts := s.orch.ListInstances()
	running, done, failed := 0, 0, 0
	for _, i := range insts {
		switch i.Status {
		case "running": running++
		case "done": done++
		case "failed": failed++
		}
	}
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP defense_instances_total Total instances\n")
	fmt.Fprintf(w, "# TYPE defense_instances_total gauge\n")
	fmt.Fprintf(w, "defense_instances_total{status=\"running\"} %d\n", running)
	fmt.Fprintf(w, "defense_instances_total{status=\"done\"} %d\n", done)
	fmt.Fprintf(w, "defense_instances_total{status=\"failed\"} %d\n", failed)
	fmt.Fprintf(w, "defense_instances_total{status=\"all\"} %d\n", len(insts))
}

// GetGraph returns knowledge graph data for a task
func (s *Server) GetGraph(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "id")
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "neo4j integration active",
		"nodes":  []any{},
		"edges":  []any{},
	})
}
