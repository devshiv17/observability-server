package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/observio/backend/internal/config"
	"github.com/observio/backend/internal/database"
)


// LogsHandler serves log data
type LogsHandler struct {
	cfg *config.Config
	logger *log.Logger
	db *database.ClickHouseClient
}


// NewLogsHandler creates a new handler for logs
func NewLogsHandler(cfg *config.Config, logger *log.Logger, db *database.ClickHouseClient) http.Handler {
	h := &LogsHandler{
		cfg: cfg,
		logger: logger,
		db: db,
	}
	r := chi.NewRouter()
	r.Get("/", h.GetLogs)
	r.Get("/top100", h.GetTop100Logs)
	return r
}


// GetTop100Logs returns the top 100 logs
func (h *LogsHandler) GetTop100Logs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	logs, err := h.db.GetTop100Logs(ctx)
	if err != nil {
		h.logger.Printf("Error fetching logs from ClickHouse: %v", err)
		respondError(w, http.StatusInternalServerError, "Could not fetch logs")
		return
	}

	respondJSON(w, http.StatusOK, logs)
}

// GetLogs reads logs and returns them as JSON with filtering
func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Optional query params: level, component, pattern, limit, offset
	level := r.URL.Query().Get("level")
	component := r.URL.Query().Get("component")
	pattern := r.URL.Query().Get("pattern")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var limit int = 100
	var offset int = 0
	var err error
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 100
		}
	}
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}
	}

	logs, err := h.db.GetLogs(ctx, limit, offset, level, component, pattern)
	if err != nil {
		h.logger.Printf("Error fetching logs from ClickHouse: %v", err)
		respondError(w, http.StatusInternalServerError, "Could not fetch logs")
		return
	}

	respondJSON(w, http.StatusOK, logs)
}


// Helper functions for HTTP responses - these should be moved to a common utility package later
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}
