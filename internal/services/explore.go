package services

import (
	"context"
	"fmt"
	"log"

	"github.com/observio/backend/internal/database"
)

// ExploreService provides business logic for explore functionality
type ExploreService struct {
	db     *database.ClickHouseClient
	logger *log.Logger
}

// NewExploreService creates a new explore service
func NewExploreService(db *database.ClickHouseClient, logger *log.Logger) *ExploreService {
	return &ExploreService{
		db:     db,
		logger: logger,
	}
}

// GetDatabases retrieves all available databases
func (s *ExploreService) GetDatabases(ctx context.Context) ([]string, error) {
	return s.db.GetDatabases(ctx)
}

// GetTables retrieves all tables for the specified database
func (s *ExploreService) GetTables(ctx context.Context, database string) ([]string, error) {
	if database == "" {
		return nil, fmt.Errorf("database name is required")
	}
	return s.db.GetTables(ctx, database)
}

// GetTableFields retrieves all fields for the specified table (excluding id fields)
func (s *ExploreService) GetTableFields(ctx context.Context, database, table string) ([]database.TableField, error) {
	if database == "" || table == "" {
		return nil, fmt.Errorf("database and table names are required")
	}
	return s.db.GetTableFields(ctx, database, table)
}

// ValidateExploreRequest validates the explore request parameters
func (s *ExploreService) ValidateExploreRequest(req database.ExploreRequest) error {
	if req.Database == "" {
		return fmt.Errorf("database is required")
	}
	if req.Table == "" {
		return fmt.Errorf("table is required")
	}
	
	// Validate aggregate function
	if req.Aggregate != "" {
		validAggregates := map[string]bool{
			"count": true,
			"sum":   true,
			"avg":   true,
			"min":   true,
			"max":   true,
		}
		if !validAggregates[req.Aggregate] {
			return fmt.Errorf("invalid aggregate function: %s", req.Aggregate)
		}
		
		// For sum, avg, min, max - ensure we have at least one field
		if req.Aggregate != "count" && len(req.Fields) == 0 {
			return fmt.Errorf("fields are required for aggregate function: %s", req.Aggregate)
		}
	}
	
	// Validate filter operation
	if req.FilterOp != "" {
		validOps := map[string]bool{
			"eq":   true,
			"ne":   true,
			"gt":   true,
			"lt":   true,
			"gte":  true,
			"lte":  true,
			"like": true,
		}
		if !validOps[req.FilterOp] {
			return fmt.Errorf("invalid filter operation: %s", req.FilterOp)
		}
		
		// Ensure filter field and value are provided
		if req.FilterBy == "" || req.FilterVal == "" {
			return fmt.Errorf("filter field and value are required when filter operation is specified")
		}
	}
	
	// Validate order direction
	if req.OrderDir != "" && req.OrderDir != "asc" && req.OrderDir != "desc" {
		return fmt.Errorf("invalid order direction: %s (must be 'asc' or 'desc')", req.OrderDir)
	}
	
	// Validate limit
	if req.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if req.Limit > 10000 {
		return fmt.Errorf("limit cannot exceed 10000 rows")
	}
	
	return nil
}

// ExecuteExploreQuery executes a validated explore query
func (s *ExploreService) ExecuteExploreQuery(ctx context.Context, req database.ExploreRequest) (*database.ExploreResponse, error) {
	// Validate the request
	if err := s.ValidateExploreRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	
	s.logger.Printf("Executing explore query for %s.%s with aggregate: %s", req.Database, req.Table, req.Aggregate)
	
	// Execute the query
	result, err := s.db.ExecuteExploreQuery(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	
	s.logger.Printf("Query executed successfully, returned %d rows", result.Total)
	
	return result, nil
}

// GetAvailableAggregates returns the list of available aggregate functions
func (s *ExploreService) GetAvailableAggregates() []string {
	return []string{"count", "sum", "avg", "min", "max"}
}

// GetAvailableFilterOperations returns the list of available filter operations
func (s *ExploreService) GetAvailableFilterOperations() []string {
	return []string{"eq", "ne", "gt", "lt", "gte", "lte", "like"}
}

// BuildExploreRequest helps build an explore request with sensible defaults
func (s *ExploreService) BuildExploreRequest(database, table string) database.ExploreRequest {
	return database.ExploreRequest{
		Database: database,
		Table:    table,
		Fields:   []string{}, // Will select all fields
		Limit:    100,        // Default limit
		OrderDir: "asc",      // Default order direction
	}
}