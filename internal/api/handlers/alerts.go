package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/observio/backend/internal/config"
)

// AlertsHandler handles alert-related API endpoints
type AlertsHandler struct {
	cfg    *config.Config
	logger *log.Logger
}

// Alert represents a monitoring alert
type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Query       string            `json:"query"`
	Threshold   float64           `json:"threshold"`
	Operator    string            `json:"operator"` // >, <, ==, !=, >=, <=
	Severity    string            `json:"severity"` // critical, warning, info
	Status      string            `json:"status"`   // active, resolved, pending
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	LastFiredAt *time.Time        `json:"lastFiredAt,omitempty"`
}

// AlertRule represents a rule for generating alerts
type AlertRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Query       string            `json:"query"`
	Threshold   float64           `json:"threshold"`
	Operator    string            `json:"operator"` // >, <, ==, !=, >=, <=
	Severity    string            `json:"severity"` // critical, warning, info
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// NewAlertsHandler creates a new alerts handler
func NewAlertsHandler(cfg *config.Config, logger *log.Logger) http.Handler {
	h := &AlertsHandler{
		cfg:    cfg,
		logger: logger,
	}

	r := chi.NewRouter()
	// Alert endpoints
	r.Get("/", h.ListAlerts)
	r.Get("/{id}", h.GetAlert)
	r.Put("/{id}/resolve", h.ResolveAlert)
	
	// Alert rules endpoints
	r.Route("/rules", func(r chi.Router) {
		r.Get("/", h.ListAlertRules)
		r.Post("/", h.CreateAlertRule)
		r.Get("/{id}", h.GetAlertRule)
		r.Put("/{id}", h.UpdateAlertRule)
		r.Delete("/{id}", h.DeleteAlertRule)
		r.Put("/{id}/enable", h.EnableAlertRule)
		r.Put("/{id}/disable", h.DisableAlertRule)
	})
	
	return r
}

// ListAlerts returns a list of all active alerts
func (h *AlertsHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	status := r.URL.Query().Get("status")
	severity := r.URL.Query().Get("severity")
	
	h.logger.Printf("Listing alerts with status: %s, severity: %s", status, severity)
	
	// In a real implementation, this would fetch alerts from a database with filters
	now := time.Now()
	firedTime := now.Add(-15 * time.Minute)
	
	alerts := []Alert{
		{
			ID:          "alert-1",
			Name:        "High CPU Usage",
			Description: "CPU usage is above threshold",
			Query:       "cpu_usage{host=\"server1\"} > 90",
			Threshold:   90.0,
			Operator:    ">",
			Severity:    "critical",
			Status:      "active",
			Labels: map[string]string{
				"host":        "server1",
				"environment": "production",
			},
			Annotations: map[string]string{
				"summary":     "High CPU usage on server1",
				"description": "CPU usage is above 90% for more than 5 minutes",
			},
			CreatedAt:   now.Add(-30 * time.Minute),
			UpdatedAt:   now.Add(-15 * time.Minute),
			LastFiredAt: &firedTime,
		},
		{
			ID:          "alert-2",
			Name:        "High Memory Usage",
			Description: "Memory usage is above threshold",
			Query:       "memory_usage{host=\"server2\"} > 85",
			Threshold:   85.0,
			Operator:    ">",
			Severity:    "warning",
			Status:      "active",
			Labels: map[string]string{
				"host":        "server2",
				"environment": "production",
			},
			Annotations: map[string]string{
				"summary":     "High memory usage on server2",
				"description": "Memory usage is above 85% for more than 10 minutes",
			},
			CreatedAt:   now.Add(-45 * time.Minute),
			UpdatedAt:   now.Add(-20 * time.Minute),
			LastFiredAt: &firedTime,
		},
	}

	// Filter by status if provided
	if status != "" {
		var filtered []Alert
		for _, alert := range alerts {
			if alert.Status == status {
				filtered = append(filtered, alert)
			}
		}
		alerts = filtered
	}

	// Filter by severity if provided
	if severity != "" {
		var filtered []Alert
		for _, alert := range alerts {
			if alert.Severity == severity {
				filtered = append(filtered, alert)
			}
		}
		alerts = filtered
	}

	respondJSON(w, http.StatusOK, alerts)
}

// GetAlert returns a specific alert by ID
func (h *AlertsHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would fetch the alert from a database
	now := time.Now()
	firedTime := now.Add(-15 * time.Minute)
	
	alert := Alert{
		ID:          id,
		Name:        "High CPU Usage",
		Description: "CPU usage is above threshold",
		Query:       "cpu_usage{host=\"server1\"} > 90",
		Threshold:   90.0,
		Operator:    ">",
		Severity:    "critical",
		Status:      "active",
		Labels: map[string]string{
			"host":        "server1",
			"environment": "production",
		},
		Annotations: map[string]string{
			"summary":     "High CPU usage on server1",
			"description": "CPU usage is above 90% for more than 5 minutes",
		},
		CreatedAt:   now.Add(-30 * time.Minute),
		UpdatedAt:   now.Add(-15 * time.Minute),
		LastFiredAt: &firedTime,
	}

	respondJSON(w, http.StatusOK, alert)
}

// ResolveAlert marks an alert as resolved
func (h *AlertsHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would update the alert status in a database
	h.logger.Printf("Resolving alert with ID: %s", id)
	
	now := time.Now()
	firedTime := now.Add(-15 * time.Minute)
	
	alert := Alert{
		ID:          id,
		Name:        "High CPU Usage",
		Description: "CPU usage is above threshold",
		Query:       "cpu_usage{host=\"server1\"} > 90",
		Threshold:   90.0,
		Operator:    ">",
		Severity:    "critical",
		Status:      "resolved", // Status updated to resolved
		Labels: map[string]string{
			"host":        "server1",
			"environment": "production",
		},
		Annotations: map[string]string{
			"summary":     "High CPU usage on server1",
			"description": "CPU usage is above 90% for more than 5 minutes",
		},
		CreatedAt:   now.Add(-30 * time.Minute),
		UpdatedAt:   now, // Update time set to now
		LastFiredAt: &firedTime,
	}

	respondJSON(w, http.StatusOK, alert)
}

// ListAlertRules returns a list of all alert rules
func (h *AlertsHandler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would fetch alert rules from a database
	now := time.Now()
	
	rules := []AlertRule{
		{
			ID:          "rule-1",
			Name:        "High CPU Usage Rule",
			Description: "Alert when CPU usage is above threshold",
			Query:       "cpu_usage{host=~\".*\"} > 90",
			Threshold:   90.0,
			Operator:    ">",
			Severity:    "critical",
			Labels: map[string]string{
				"service": "system",
				"team":    "infrastructure",
			},
			Annotations: map[string]string{
				"summary":     "High CPU usage detected",
				"description": "CPU usage is above 90% for more than 5 minutes",
			},
			Enabled:   true,
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: now.Add(-12 * time.Hour),
		},
		{
			ID:          "rule-2",
			Name:        "High Memory Usage Rule",
			Description: "Alert when memory usage is above threshold",
			Query:       "memory_usage{host=~\".*\"} > 85",
			Threshold:   85.0,
			Operator:    ">",
			Severity:    "warning",
			Labels: map[string]string{
				"service": "system",
				"team":    "infrastructure",
			},
			Annotations: map[string]string{
				"summary":     "High memory usage detected",
				"description": "Memory usage is above 85% for more than 10 minutes",
			},
			Enabled:   true,
			CreatedAt: now.Add(-48 * time.Hour),
			UpdatedAt: now.Add(-24 * time.Hour),
		},
	}

	respondJSON(w, http.StatusOK, rules)
}

// GetAlertRule returns a specific alert rule by ID
func (h *AlertsHandler) GetAlertRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would fetch the alert rule from a database
	now := time.Now()
	
	rule := AlertRule{
		ID:          id,
		Name:        "High CPU Usage Rule",
		Description: "Alert when CPU usage is above threshold",
		Query:       "cpu_usage{host=~\".*\"} > 90",
		Threshold:   90.0,
		Operator:    ">",
		Severity:    "critical",
		Labels: map[string]string{
			"service": "system",
			"team":    "infrastructure",
		},
		Annotations: map[string]string{
			"summary":     "High CPU usage detected",
			"description": "CPU usage is above 90% for more than 5 minutes",
		},
		Enabled:   true,
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now.Add(-12 * time.Hour),
	}

	respondJSON(w, http.StatusOK, rule)
}

// CreateAlertRule creates a new alert rule
func (h *AlertsHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	var rule AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// In a real implementation, this would save the alert rule to a database
	rule.ID = "new-rule-id"
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	rule.Enabled = true

	respondJSON(w, http.StatusCreated, rule)
}

// UpdateAlertRule updates an existing alert rule
func (h *AlertsHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var rule AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// In a real implementation, this would update the alert rule in a database
	rule.ID = id
	rule.UpdatedAt = time.Now()

	respondJSON(w, http.StatusOK, rule)
}

// DeleteAlertRule deletes an alert rule
func (h *AlertsHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would delete the alert rule from a database
	h.logger.Printf("Deleting alert rule with ID: %s", id)

	respondJSON(w, http.StatusOK, map[string]string{"message": "Alert rule deleted successfully"})
}

// EnableAlertRule enables an alert rule
func (h *AlertsHandler) EnableAlertRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would enable the alert rule in a database
	h.logger.Printf("Enabling alert rule with ID: %s", id)

	// Return the updated rule
	now := time.Now()
	rule := AlertRule{
		ID:          id,
		Name:        "High CPU Usage Rule",
		Description: "Alert when CPU usage is above threshold",
		Query:       "cpu_usage{host=~\".*\"} > 90",
		Threshold:   90.0,
		Operator:    ">",
		Severity:    "critical",
		Labels: map[string]string{
			"service": "system",
			"team":    "infrastructure",
		},
		Annotations: map[string]string{
			"summary":     "High CPU usage detected",
			"description": "CPU usage is above 90% for more than 5 minutes",
		},
		Enabled:   true, // Set to enabled
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now, // Update time set to now
	}

	respondJSON(w, http.StatusOK, rule)
}

// DisableAlertRule disables an alert rule
func (h *AlertsHandler) DisableAlertRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would disable the alert rule in a database
	h.logger.Printf("Disabling alert rule with ID: %s", id)

	// Return the updated rule
	now := time.Now()
	rule := AlertRule{
		ID:          id,
		Name:        "High CPU Usage Rule",
		Description: "Alert when CPU usage is above threshold",
		Query:       "cpu_usage{host=~\".*\"} > 90",
		Threshold:   90.0,
		Operator:    ">",
		Severity:    "critical",
		Labels: map[string]string{
			"service": "system",
			"team":    "infrastructure",
		},
		Annotations: map[string]string{
			"summary":     "High CPU usage detected",
			"description": "CPU usage is above 90% for more than 5 minutes",
		},
		Enabled:   false, // Set to disabled
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now, // Update time set to now
	}

	respondJSON(w, http.StatusOK, rule)
}
