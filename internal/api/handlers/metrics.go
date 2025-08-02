package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/observio/backend/internal/config"
)

// MetricsHandler handles metrics-related API endpoints
type MetricsHandler struct {
	cfg    *config.Config
	logger *log.Logger
}

// MetricResponse represents a metric data point or series
type MetricResponse struct {
	Name      string      `json:"name"`
	Labels    map[string]string `json:"labels"`
	Value     float64     `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

// MetricQuery represents a query for metrics data
type MetricQuery struct {
	Query     string     `json:"query"`
	Start     time.Time  `json:"start"`
	End       time.Time  `json:"end"`
	Step      string     `json:"step"`
	DataSource string    `json:"dataSource"`
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(cfg *config.Config, logger *log.Logger) http.Handler {
	h := &MetricsHandler{
		cfg:    cfg,
		logger: logger,
	}

	r := chi.NewRouter()
	r.Get("/", h.GetMetrics)
	r.Post("/query", h.QueryMetrics)
	r.Get("/{name}", h.GetMetricByName)
	
	return r
}

// GetMetrics returns a list of available metrics
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would fetch metrics from a database or time series database
	metrics := []string{
		"cpu_usage",
		"memory_usage",
		"request_duration",
		"error_rate",
		"throughput",
	}

	respondJSON(w, http.StatusOK, metrics)
}

// QueryMetrics handles complex metric queries
func (h *MetricsHandler) QueryMetrics(w http.ResponseWriter, r *http.Request) {
	var query MetricQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// In a real implementation, this would query a time series database like Prometheus
	// For now, we'll return sample data
	now := time.Now()
	metrics := []MetricResponse{
		{
			Name:      "cpu_usage",
			Labels:    map[string]string{"host": "server1", "env": "prod"},
			Value:     42.5,
			Timestamp: now.Add(-5 * time.Minute),
		},
		{
			Name:      "cpu_usage",
			Labels:    map[string]string{"host": "server1", "env": "prod"},
			Value:     45.2,
			Timestamp: now.Add(-4 * time.Minute),
		},
		{
			Name:      "cpu_usage",
			Labels:    map[string]string{"host": "server1", "env": "prod"},
			Value:     47.8,
			Timestamp: now.Add(-3 * time.Minute),
		},
		{
			Name:      "cpu_usage",
			Labels:    map[string]string{"host": "server1", "env": "prod"},
			Value:     44.3,
			Timestamp: now.Add(-2 * time.Minute),
		},
		{
			Name:      "cpu_usage",
			Labels:    map[string]string{"host": "server1", "env": "prod"},
			Value:     43.1,
			Timestamp: now.Add(-1 * time.Minute),
		},
	}

	respondJSON(w, http.StatusOK, metrics)
}

// GetMetricByName returns data for a specific metric
func (h *MetricsHandler) GetMetricByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	
	// In a real implementation, this would fetch the specific metric from a database
	now := time.Now()
	metric := MetricResponse{
		Name:      name,
		Labels:    map[string]string{"host": "server1", "env": "prod"},
		Value:     42.5,
		Timestamp: now,
	}

	respondJSON(w, http.StatusOK, metric)
}
