package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/observio/backend/internal/config"
)

// DataSourceHandler handles data source-related API endpoints
type DataSourceHandler struct {
	cfg    *config.Config
	logger *log.Logger
}

// DataSource represents a data source for metrics, logs, or traces
type DataSource struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // prometheus, elasticsearch, loki, jaeger, etc.
	URL         string                 `json:"url"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
	IsDefault   bool                   `json:"isDefault"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// NewDataSourceHandler creates a new data source handler
func NewDataSourceHandler(cfg *config.Config, logger *log.Logger) http.Handler {
	h := &DataSourceHandler{
		cfg:    cfg,
		logger: logger,
	}

	r := chi.NewRouter()
	r.Get("/", h.ListDataSources)
	r.Post("/", h.CreateDataSource)
	r.Get("/{id}", h.GetDataSource)
	r.Put("/{id}", h.UpdateDataSource)
	r.Delete("/{id}", h.DeleteDataSource)
	r.Post("/{id}/test", h.TestDataSource)
	
	return r
}

// ListDataSources returns a list of all data sources
func (h *DataSourceHandler) ListDataSources(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would fetch data sources from a database
	now := time.Now()
	
	dataSources := []DataSource{
		{
			ID:          "ds-1",
			Name:        "Prometheus",
			Type:        "prometheus",
			URL:         "http://prometheus:9090",
			Description: "Main Prometheus server",
			Settings: map[string]interface{}{
				"scrapeInterval": "15s",
				"timeout":        "10s",
			},
			IsDefault: true,
			CreatedAt: now.Add(-72 * time.Hour),
			UpdatedAt: now.Add(-24 * time.Hour),
		},
		{
			ID:          "ds-2",
			Name:        "Elasticsearch",
			Type:        "elasticsearch",
			URL:         "http://elasticsearch:9200",
			Description: "Elasticsearch for logs",
			Settings: map[string]interface{}{
				"index":      "logs-*",
				"timeField":  "@timestamp",
				"esVersion":  7,
				"maxConcurrentShardRequests": 5,
			},
			IsDefault: false,
			CreatedAt: now.Add(-48 * time.Hour),
			UpdatedAt: now.Add(-12 * time.Hour),
		},
		{
			ID:          "ds-3",
			Name:        "Jaeger",
			Type:        "jaeger",
			URL:         "http://jaeger:16686",
			Description: "Jaeger for distributed tracing",
			Settings: map[string]interface{}{
				"queryTimeout": "30s",
			},
			IsDefault: false,
			CreatedAt: now.Add(-36 * time.Hour),
			UpdatedAt: now.Add(-6 * time.Hour),
		},
	}

	respondJSON(w, http.StatusOK, dataSources)
}

// GetDataSource returns a specific data source by ID
func (h *DataSourceHandler) GetDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would fetch the data source from a database
	now := time.Now()
	
	dataSource := DataSource{
		ID:          id,
		Name:        "Prometheus",
		Type:        "prometheus",
		URL:         "http://prometheus:9090",
		Description: "Main Prometheus server",
		Settings: map[string]interface{}{
			"scrapeInterval": "15s",
			"timeout":        "10s",
		},
		IsDefault: true,
		CreatedAt: now.Add(-72 * time.Hour),
		UpdatedAt: now.Add(-24 * time.Hour),
	}

	respondJSON(w, http.StatusOK, dataSource)
}

// CreateDataSource creates a new data source
func (h *DataSourceHandler) CreateDataSource(w http.ResponseWriter, r *http.Request) {
	var dataSource DataSource
	if err := json.NewDecoder(r.Body).Decode(&dataSource); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// In a real implementation, this would save the data source to a database
	dataSource.ID = "new-ds-id"
	dataSource.CreatedAt = time.Now()
	dataSource.UpdatedAt = time.Now()

	// If this is set as default, we would update other data sources to not be default
	if dataSource.IsDefault {
		h.logger.Println("Setting new data source as default")
	}

	respondJSON(w, http.StatusCreated, dataSource)
}

// UpdateDataSource updates an existing data source
func (h *DataSourceHandler) UpdateDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var dataSource DataSource
	if err := json.NewDecoder(r.Body).Decode(&dataSource); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// In a real implementation, this would update the data source in a database
	dataSource.ID = id
	dataSource.UpdatedAt = time.Now()

	// If this is set as default, we would update other data sources to not be default
	if dataSource.IsDefault {
		h.logger.Println("Setting updated data source as default")
	}

	respondJSON(w, http.StatusOK, dataSource)
}

// DeleteDataSource deletes a data source
func (h *DataSourceHandler) DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would delete the data source from a database
	h.logger.Printf("Deleting data source with ID: %s", id)

	// Check if this is a default data source
	// In a real implementation, we would prevent deletion of the default data source
	// or automatically set another one as default

	respondJSON(w, http.StatusOK, map[string]string{"message": "Data source deleted successfully"})
}

// TestDataSource tests the connection to a data source
func (h *DataSourceHandler) TestDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// In a real implementation, this would test the connection to the data source
	h.logger.Printf("Testing connection to data source with ID: %s", id)

	// Simulate a successful connection test
	result := map[string]interface{}{
		"status":  "success",
		"message": "Connection successful",
		"details": map[string]interface{}{
			"version":      "2.30.0",
			"responseTime": "42ms",
		},
	}

	respondJSON(w, http.StatusOK, result)
}
