package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

// AutocompleteRequest represents the request for SQL autocomplete
type AutocompleteRequest struct {
	Database string `json:"database"`
	Query    string `json:"query"`
	Position int    `json:"position"`
}

// AutocompleteSuggestion represents a single autocomplete suggestion
type AutocompleteSuggestion struct {
	Text        string `json:"text"`
	Type        string `json:"type"` // "table", "column", "keyword", "function"
	Description string `json:"description,omitempty"`
}

// AutocompleteResponse represents the response for SQL autocomplete
type AutocompleteResponse struct {
	Suggestions []AutocompleteSuggestion `json:"suggestions"`
}

// RawSQLRequest represents a raw SQL query request
type RawSQLRequest struct {
	Database string `json:"database"`
	Query    string `json:"query"`
}

// RawSQLResponse represents the response from a raw SQL query
type RawSQLResponse struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	Total   int                      `json:"total"`
	Query   string                   `json:"query"`
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
	r.Post("/autocomplete", h.GetAutocomplete)
	r.Post("/execute-sql", h.ExecuteRawSQL)
	
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

// GetAutocomplete provides SQL autocomplete suggestions
func (h *ExploreHandler) GetAutocomplete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req AutocompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if req.Database == "" {
		respondError(w, http.StatusBadRequest, "Database is required")
		return
	}
	
	h.logger.Printf("Getting autocomplete suggestions for database: %s, query: %s", req.Database, req.Query)
	
	suggestions, err := h.getAutocompleteSuggestions(ctx, req)
	if err != nil {
		h.logger.Printf("Error getting autocomplete suggestions: %v", err)
		respondError(w, http.StatusInternalServerError, "Could not get autocomplete suggestions")
		return
	}
	
	response := AutocompleteResponse{
		Suggestions: suggestions,
	}
	
	respondJSON(w, http.StatusOK, response)
}

// getAutocompleteSuggestions generates autocomplete suggestions based on the query context
func (h *ExploreHandler) getAutocompleteSuggestions(ctx context.Context, req AutocompleteRequest) ([]AutocompleteSuggestion, error) {
	var suggestions []AutocompleteSuggestion
	
	// Convert query to lowercase for pattern matching
	query := strings.ToLower(req.Query)
	
	// Get the word at cursor position
	wordAtCursor := h.getWordAtPosition(req.Query, req.Position)
	
	// Always include SQL keywords
	keywords := []string{
		"SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "HAVING", "LIMIT", "OFFSET",
		"JOIN", "LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "OUTER JOIN", "FULL JOIN",
		"ON", "AND", "OR", "NOT", "IN", "LIKE", "BETWEEN", "IS", "NULL", "TRUE", "FALSE",
		"COUNT", "SUM", "AVG", "MIN", "MAX", "DISTINCT", "AS", "ASC", "DESC",
	}
	
	for _, keyword := range keywords {
		if strings.HasPrefix(strings.ToLower(keyword), strings.ToLower(wordAtCursor)) {
			suggestions = append(suggestions, AutocompleteSuggestion{
				Text:        keyword,
				Type:        "keyword",
				Description: "SQL keyword",
			})
		}
	}
	
	// Get table suggestions if we're in a context where tables are expected
	if h.shouldSuggestTables(query, req.Position) {
		tables, err := h.db.GetTables(ctx, req.Database)
		if err == nil {
			for _, table := range tables {
				if strings.HasPrefix(strings.ToLower(table), strings.ToLower(wordAtCursor)) {
					suggestions = append(suggestions, AutocompleteSuggestion{
						Text:        table,
						Type:        "table",
						Description: fmt.Sprintf("Table in %s database", req.Database),
					})
				}
			}
		}
	}
	
	// Get column suggestions if we're in a context where columns are expected
	if tableName := h.getTableFromQuery(query); tableName != "" {
		fields, err := h.db.GetTableFields(ctx, req.Database, tableName)
		if err == nil {
			for _, field := range fields {
				if strings.HasPrefix(strings.ToLower(field.Name), strings.ToLower(wordAtCursor)) {
					suggestions = append(suggestions, AutocompleteSuggestion{
						Text:        field.Name,
						Type:        "column",
						Description: fmt.Sprintf("Column (%s) in %s.%s", field.Type, req.Database, tableName),
					})
				}
			}
		}
	}
	
	// Limit suggestions to avoid overwhelming the UI
	if len(suggestions) > 20 {
		suggestions = suggestions[:20]
	}
	
	return suggestions, nil
}

// getWordAtPosition extracts the word at the given cursor position
func (h *ExploreHandler) getWordAtPosition(query string, position int) string {
	if position < 0 || position > len(query) {
		return ""
	}
	
	// Find the start of the word
	start := position
	for start > 0 && (isAlphaNumeric(query[start-1]) || query[start-1] == '_') {
		start--
	}
	
	// Find the end of the word
	end := position
	for end < len(query) && (isAlphaNumeric(query[end]) || query[end] == '_') {
		end++
	}
	
	return query[start:end]
}

// shouldSuggestTables determines if we should suggest table names based on query context
func (h *ExploreHandler) shouldSuggestTables(query string, position int) bool {
	// Simple heuristics for when to suggest tables
	fromIndex := strings.LastIndex(query[:position], "from")
	joinIndex := strings.LastIndex(query[:position], "join")
	
	// If we find FROM or JOIN keywords recently, suggest tables
	return fromIndex >= 0 || joinIndex >= 0
}

// getTableFromQuery attempts to extract the main table name from the query
func (h *ExploreHandler) getTableFromQuery(query string) string {
	// Simple regex to find "FROM tablename" pattern
	fromIndex := strings.Index(query, "from")
	if fromIndex == -1 {
		return ""
	}
	
	// Look for the table name after FROM
	afterFrom := query[fromIndex+4:] // Skip "from"
	afterFrom = strings.TrimSpace(afterFrom)
	
	// Get the first word (table name)
	words := strings.Fields(afterFrom)
	if len(words) > 0 {
		return words[0]
	}
	
	return ""
}

// isAlphaNumeric checks if a character is alphanumeric
func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// ExecuteRawSQL executes a raw SQL query
func (h *ExploreHandler) ExecuteRawSQL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req RawSQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if req.Database == "" {
		respondError(w, http.StatusBadRequest, "Database is required")
		return
	}
	
	if req.Query == "" {
		respondError(w, http.StatusBadRequest, "Query is required")
		return
	}
	
	// Basic security: prevent dangerous operations
	queryLower := strings.ToLower(strings.TrimSpace(req.Query))
	if strings.HasPrefix(queryLower, "drop") || 
	   strings.HasPrefix(queryLower, "delete") || 
	   strings.HasPrefix(queryLower, "truncate") || 
	   strings.HasPrefix(queryLower, "alter") ||
	   strings.HasPrefix(queryLower, "create") ||
	   strings.HasPrefix(queryLower, "insert") ||
	   strings.HasPrefix(queryLower, "update") {
		respondError(w, http.StatusBadRequest, "Only SELECT queries are allowed")
		return
	}
	
	h.logger.Printf("Executing raw SQL query on database %s: %s", req.Database, req.Query)
	
	// Execute the query
	columns, results, err := h.db.QueryRaw(ctx, req.Query)
	if err != nil {
		h.logger.Printf("Error executing raw SQL query: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to execute query")
		return
	}
	
	response := RawSQLResponse{
		Columns: columns,
		Rows:    results,
		Total:   len(results),
		Query:   req.Query,
	}
	
	h.logger.Printf("Successfully executed raw SQL query, returning %d rows", len(results))
	respondJSON(w, http.StatusOK, response)
}

// Helper functions are imported from logs.go handler