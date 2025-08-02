package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/observio/backend/internal/config"
)

// DashboardHandler handles dashboard-related API endpoints
type DashboardHandler struct {
	cfg    *config.Config
	logger *log.Logger
}

// Dashboard represents a monitoring dashboard
type Dashboard struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Panels      []Panel   `json:"panels"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   string    `json:"createdBy"`
}

// Panel represents a visualization panel within a dashboard
type Panel struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // graph, singlestat, table, etc.
	Query       string                 `json:"query"`
	DataSource  string                 `json:"dataSource"`
	Position    map[string]int         `json:"position"` // x, y, w, h
	Options     map[string]interface{} `json:"options"`
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(cfg *config.Config, logger *log.Logger) http.Handler {
	h := &DashboardHandler{
		cfg:    cfg,
		logger: logger,
	}

	r := chi.NewRouter()
	r.Get("/", h.ListDashboards)
	r.Post("/", h.CreateDashboard)
	r.Get("/{id}", h.GetDashboard)
	r.Put("/{id}", h.UpdateDashboard)
	r.Delete("/{id}", h.DeleteDashboard)
	
	return r
}

// ListDashboards returns a list of all dashboards
func (h *DashboardHandler) ListDashboards(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would fetch dashboards from a database
	dashboards := []Dashboard{
		{
			ID:          "dashboard-1",
			Title:       "System Overview",
			Description: "Overview of system metrics",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-12 * time.Hour),
			CreatedBy:   "admin",
			Panels: []Panel{
				{
					ID:         "panel-1",
					Title:      "CPU Usage",
					Type:       "graph",
					Query:      "cpu_usage{host=~\".*\"}",
					DataSource: "prometheus",
					Position:   map[string]int{"x": 0, "y": 0, "w": 12, "h": 8},
				},
				{
					ID:         "panel-2",
					Title:      "Memory Usage",
					Type:       "graph",
					Query:      "memory_usage{host=~\".*\"}",
					DataSource: "prometheus",
					Position:   map[string]int{"x": 12, "y": 0, "w": 12, "h": 8},
				},
			},
		},
		{
			ID:          "dashboard-2",
			Title:       "Application Performance",
			Description: "Application performance metrics",
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
			CreatedBy:   "admin",
			Panels: []Panel{
				{
					ID:         "panel-3",
					Title:      "Request Duration",
					Type:       "graph",
					Query:      "request_duration{service=~\".*\"}",
					DataSource: "prometheus",
					Position:   map[string]int{"x": 0, "y": 0, "w": 24, "h": 8},
				},
				{
					ID:         "panel-4",
					Title:      "Error Rate",
					Type:       "graph",
					Query:      "error_rate{service=~\".*\"}",
					DataSource: "prometheus",
					Position:   map[string]int{"x": 0, "y": 8, "w": 12, "h": 8},
				},
				{
					ID:         "panel-5",
					Title:      "Throughput",
					Type:       "graph",
					Query:      "throughput{service=~\".*\"}",
					DataSource: "prometheus",
					Position:   map[string]int{"x": 12, "y": 8, "w": 12, "h": 8},
				},
			},
		},
	}

	respondJSON(w, http.StatusOK, dashboards)
}

// GetDashboard returns a specific dashboard by ID
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would fetch the dashboard from a database
	dashboard := Dashboard{
		ID:          id,
		Title:       "System Overview",
		Description: "Overview of system metrics",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-12 * time.Hour),
		CreatedBy:   "admin",
		Panels: []Panel{
			{
				ID:         "panel-1",
				Title:      "CPU Usage",
				Type:       "graph",
				Query:      "cpu_usage{host=~\".*\"}",
				DataSource: "prometheus",
				Position:   map[string]int{"x": 0, "y": 0, "w": 12, "h": 8},
			},
			{
				ID:         "panel-2",
				Title:      "Memory Usage",
				Type:       "graph",
				Query:      "memory_usage{host=~\".*\"}",
				DataSource: "prometheus",
				Position:   map[string]int{"x": 12, "y": 0, "w": 12, "h": 8},
			},
		},
	}

	respondJSON(w, http.StatusOK, dashboard)
}

// CreateDashboard creates a new dashboard
func (h *DashboardHandler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	var dashboard Dashboard
	if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// In a real implementation, this would save the dashboard to a database
	// For now, we'll just return the dashboard with a generated ID
	dashboard.ID = "new-dashboard-id"
	dashboard.CreatedAt = time.Now()
	dashboard.UpdatedAt = time.Now()
	dashboard.CreatedBy = "current-user" // In a real app, this would come from auth context

	respondJSON(w, http.StatusCreated, dashboard)
}

// UpdateDashboard updates an existing dashboard
func (h *DashboardHandler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var dashboard Dashboard
	if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// In a real implementation, this would update the dashboard in a database
	dashboard.ID = id
	dashboard.UpdatedAt = time.Now()

	respondJSON(w, http.StatusOK, dashboard)
}

// DeleteDashboard deletes a dashboard
func (h *DashboardHandler) DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would delete the dashboard from a database
	h.logger.Printf("Deleting dashboard with ID: %s", id)

	// Return a success message
	respondJSON(w, http.StatusOK, map[string]string{"message": "Dashboard deleted successfully"})
}
