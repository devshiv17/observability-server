package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/observio/backend/internal/config"
	"github.com/observio/backend/internal/database"
)

// ExploreHandler serves explore data for query builder
type ExploreHandler struct {
	cfg    *config.Config
	logger *log.Logger
	db     *database.ClickHouseClient
}

// DatabaseResponse represents the response structure for databases
type DatabaseResponse struct {
	Databases []string `json:"databases"`
}

// TablesResponse represents the response structure for tables
type TablesResponse struct {
	Tables []string `json:"tables"`
}

// TableFieldsResponse represents the response structure for table fields
type TableFieldsResponse struct {
	Fields []database.TableField `json:"fields"`
}

// NewExploreHandler creates a new handler for explore endpoints
func NewExploreHandler(cfg *config.Config, logger *log.Logger, db *database.ClickHouseClient) http.Handler {
	h := &ExploreHandler{
		cfg:    cfg,
		logger: logger,
		db:     db,
	}
	
	r := chi.NewRouter()
	r.Get("/databases", h.GetDatabases)
	r.Get("/databases/{database}/tables", h.GetTables)
	r.Get("/databases/{database}/tables/{table}/fields", h.GetTableFields)
	r.Post("/query", h.ExecuteQuery)
	
	return r
}

// GetDatabases retrieves all available databases
func (h *ExploreHandler) GetDatabases(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	h.logger.Printf("Fetching databases from ClickHouse")
	
	databases, err := h.db.GetDatabases(ctx)
	if err != nil {
		h.logger.Printf("Error fetching databases from ClickHouse: %v", err)
		respondError(w, http.StatusInternalServerError, "Could not fetch databases")
		return
	}
	
	h.logger.Printf("Successfully fetched %d databases", len(databases))
	
	response := DatabaseResponse{
		Databases: databases,
	}
	
	respondJSON(w, http.StatusOK, response)
}

// GetTables retrieves all tables for the specified database
func (h *ExploreHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	database := chi.URLParam(r, "database")
	
	if database == "" {
		respondError(w, http.StatusBadRequest, "Database parameter is required")
		return
	}
	
	h.logger.Printf("Fetching tables for database: %s", database)
	
	tables, err := h.db.GetTables(ctx, database)
	if err != nil {
		h.logger.Printf("Error fetching tables for database %s: %v", database, err)
		respondError(w, http.StatusInternalServerError, "Could not fetch tables")
		return
	}
	
	h.logger.Printf("Successfully fetched %d tables for database %s", len(tables), database)
	
	response := TablesResponse{
		Tables: tables,
	}
	
	respondJSON(w, http.StatusOK, response)
}

// GetTableFields retrieves all fields for the specified table (excluding id fields)
func (h *ExploreHandler) GetTableFields(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	database := chi.URLParam(r, "database")
	table := chi.URLParam(r, "table")
	
	if database == "" || table == "" {
		respondError(w, http.StatusBadRequest, "Database and table parameters are required")
		return
	}
	
	h.logger.Printf("Fetching fields for table: %s.%s", database, table)
	
	fields, err := h.db.GetTableFields(ctx, database, table)
	if err != nil {
		h.logger.Printf("Error fetching fields for table %s.%s: %v", database, table, err)
		respondError(w, http.StatusInternalServerError, "Could not fetch table fields")
		return
	}
	
	h.logger.Printf("Successfully fetched %d fields for table %s.%s", len(fields), database, table)
	
	response := TableFieldsResponse{
		Fields: fields,
	}
	
	respondJSON(w, http.StatusOK, response)
}

// ExecuteQuery executes a dynamic explore query
func (h *ExploreHandler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req database.ExploreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if req.Database == "" || req.Table == "" {
		respondError(w, http.StatusBadRequest, "Database and table are required")
		return
	}
	
	h.logger.Printf("Executing explore query for table: %s.%s", req.Database, req.Table)
	
	result, err := h.db.ExecuteExploreQuery(ctx, req)
	if err != nil {
		h.logger.Printf("Error executing explore query: %v", err)
		respondError(w, http.StatusInternalServerError, "Could not execute query")
		return
	}
	
	h.logger.Printf("Successfully executed explore query, returning %d rows", result.Total)
	
	respondJSON(w, http.StatusOK, result)
}

// Helper functions are imported from logs.go handler