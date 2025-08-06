package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseClient struct {
	conn   clickhouse.Conn
	logger *log.Logger
}

type LogEntry struct {
	LineId      string `json:"lineId"`
	Timestamp   string `json:"timestamp"`
	Level       string `json:"level"`
	Component   string `json:"component"`
	PID         string `json:"pid"`
	Content     string `json:"content"`
	EventId     string `json:"eventId,omitempty"`
	RawMessage  string `json:"rawMessage"`
}

func NewClickHouseClient(host string, port int, username, password, database string, logger *log.Logger) (*ClickHouseClient, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", host, port)},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &ClickHouseClient{
		conn:   conn,
		logger: logger,
	}, nil
}

func (c *ClickHouseClient) Close() error {
	return c.conn.Close()
}

func (c *ClickHouseClient) GetLogs(ctx context.Context, limit, offset int, level, component, pattern string) ([]LogEntry, error) {
	query := `
		SELECT 
			toString(rowNumberInAllBlocks()) as line_id,
			toString(Timestamp) as timestamp,
			SeverityText as level,
			ServiceName as component,
			ResourceAttributes['process.pid'] as pid,
			Body as content,
			toString(cityHash64(Body)) as event_id,
			Body as raw_message
		FROM otel_logs 
		WHERE 1=1
	`
	
	args := []interface{}{}
	argIndex := 1

	if level != "" {
		query += fmt.Sprintf(" AND lower(SeverityText) = lower($%d)", argIndex)
		args = append(args, level)
		argIndex++
	}

	if component != "" {
		query += fmt.Sprintf(" AND lower(ServiceName) LIKE lower($%d)", argIndex)
		args = append(args, "%"+component+"%")
		argIndex++
	}

	if pattern != "" {
		query += fmt.Sprintf(" AND lower(Body) LIKE lower($%d)", argIndex)
		args = append(args, "%"+pattern+"%")
		argIndex++
	}

	query += " ORDER BY Timestamp DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		var pid sql.NullString
		
		err := rows.Scan(
			&log.LineId,
			&log.Timestamp,
			&log.Level,
			&log.Component,
			&pid,
			&log.Content,
			&log.EventId,
			&log.RawMessage,
		)
		if err != nil {
			c.logger.Printf("Error scanning row: %v", err)
			continue
		}

		if pid.Valid {
			log.PID = pid.String
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return logs, nil
}

func (c *ClickHouseClient) GetTop100Logs(ctx context.Context) ([]LogEntry, error) {
	return c.GetLogs(ctx, 100, 0, "", "", "")
}

// GetDatabases retrieves all databases from ClickHouse
func (c *ClickHouseClient) GetDatabases(ctx context.Context) ([]string, error) {
	query := `SELECT name FROM system.databases WHERE name NOT IN ('system', 'INFORMATION_SCHEMA', 'information_schema') ORDER BY name`
	
	rows, err := c.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			c.logger.Printf("Error scanning database row: %v", err)
			continue
		}
		databases = append(databases, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating database rows: %w", err)
	}

	return databases, nil
}

// GetTables retrieves all tables from the specified database
func (c *ClickHouseClient) GetTables(ctx context.Context, database string) ([]string, error) {
	if database == "" {
		return nil, fmt.Errorf("database name cannot be empty")
	}

	query := `SELECT name FROM system.tables WHERE database = ? ORDER BY name`
	
	rows, err := c.conn.Query(ctx, query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables for database %s: %w", database, err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			c.logger.Printf("Error scanning table row: %v", err)
			continue
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	return tables, nil
}

// TableField represents a column in a table
type TableField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetTableFields retrieves all fields from the specified table (excluding id fields)
func (c *ClickHouseClient) GetTableFields(ctx context.Context, database, table string) ([]TableField, error) {
	if database == "" || table == "" {
		return nil, fmt.Errorf("database and table names cannot be empty")
	}

	query := `
		SELECT name, type 
		FROM system.columns 
		WHERE database = ? AND table = ? AND lower(name) NOT LIKE '%id%'
		ORDER BY name
	`
	
	rows, err := c.conn.Query(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query table fields for %s.%s: %w", database, table, err)
	}
	defer rows.Close()

	var fields []TableField
	for rows.Next() {
		var field TableField
		if err := rows.Scan(&field.Name, &field.Type); err != nil {
			c.logger.Printf("Error scanning field row: %v", err)
			continue
		}
		fields = append(fields, field)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating field rows: %w", err)
	}

	return fields, nil
}

// ExploreRequest represents the request structure for explore queries
type ExploreRequest struct {
	Database  string   `json:"database"`
	Table     string   `json:"table"`
	Fields    []string `json:"fields"`
	Aggregate string   `json:"aggregate,omitempty"`
	GroupBy   []string `json:"groupBy,omitempty"`
	OrderBy   string   `json:"orderBy,omitempty"`
	OrderDir  string   `json:"orderDir,omitempty"`
	FilterBy  string   `json:"filterBy,omitempty"`
	FilterOp  string   `json:"filterOp,omitempty"`
	FilterVal string   `json:"filterVal,omitempty"`
	Limit     int      `json:"limit,omitempty"`
}

// ExploreResponse represents the response structure for explore queries
type ExploreResponse struct {
	Columns []string                 `json:"columns"`
	Data    []map[string]interface{} `json:"data"`
	Total   int                      `json:"total"`
}

// ExecuteExploreQuery executes a dynamic explore query based on the request
func (c *ClickHouseClient) ExecuteExploreQuery(ctx context.Context, req ExploreRequest) (*ExploreResponse, error) {
	if req.Database == "" || req.Table == "" {
		return nil, fmt.Errorf("database and table are required")
	}

	// Build SELECT clause
	var selectClause string
	if req.Aggregate != "" && len(req.Fields) > 0 {
		switch req.Aggregate {
		case "count":
			selectClause = "COUNT(*) as count"
		case "sum":
			selectClause = fmt.Sprintf("SUM(%s) as sum_%s", req.Fields[0], req.Fields[0])
		case "avg":
			selectClause = fmt.Sprintf("AVG(%s) as avg_%s", req.Fields[0], req.Fields[0])
		case "min":
			selectClause = fmt.Sprintf("MIN(%s) as min_%s", req.Fields[0], req.Fields[0])
		case "max":
			selectClause = fmt.Sprintf("MAX(%s) as max_%s", req.Fields[0], req.Fields[0])
		default:
			return nil, fmt.Errorf("unsupported aggregate function: %s", req.Aggregate)
		}
		
		// Add group by fields to select if specified
		if len(req.GroupBy) > 0 {
			for _, field := range req.GroupBy {
				selectClause += fmt.Sprintf(", %s", field)
			}
		}
	} else {
		// Regular field selection
		if len(req.Fields) == 0 {
			selectClause = "*"
		} else {
			selectClause = ""
			for i, field := range req.Fields {
				if i > 0 {
					selectClause += ", "
				}
				selectClause += field
			}
		}
	}

	// Build query
	query := fmt.Sprintf("SELECT %s FROM %s.%s", selectClause, req.Database, req.Table)
	args := []interface{}{}
	argIndex := 1

	// Add WHERE clause if filter is specified
	if req.FilterBy != "" && req.FilterOp != "" && req.FilterVal != "" {
		switch req.FilterOp {
		case "eq":
			query += fmt.Sprintf(" WHERE %s = $%d", req.FilterBy, argIndex)
		case "ne":
			query += fmt.Sprintf(" WHERE %s != $%d", req.FilterBy, argIndex)
		case "gt":
			query += fmt.Sprintf(" WHERE %s > $%d", req.FilterBy, argIndex)
		case "lt":
			query += fmt.Sprintf(" WHERE %s < $%d", req.FilterBy, argIndex)
		case "gte":
			query += fmt.Sprintf(" WHERE %s >= $%d", req.FilterBy, argIndex)
		case "lte":
			query += fmt.Sprintf(" WHERE %s <= $%d", req.FilterBy, argIndex)
		case "like":
			query += fmt.Sprintf(" WHERE %s LIKE $%d", req.FilterBy, argIndex)
			req.FilterVal = "%" + req.FilterVal + "%"
		default:
			return nil, fmt.Errorf("unsupported filter operation: %s", req.FilterOp)
		}
		args = append(args, req.FilterVal)
		argIndex++
	}

	// Add GROUP BY clause
	if len(req.GroupBy) > 0 {
		query += " GROUP BY "
		for i, field := range req.GroupBy {
			if i > 0 {
				query += ", "
			}
			query += field
		}
	}

	// Add ORDER BY clause
	if req.OrderBy != "" {
		orderDir := "ASC"
		if req.OrderDir == "desc" {
			orderDir = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", req.OrderBy, orderDir)
	}

	// Add LIMIT clause
	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, req.Limit)
	} else {
		// Default limit
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, 1000)
	}

	c.logger.Printf("Executing explore query: %s with args: %v", query, args)

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute explore query: %w", err)
	}
	defer rows.Close()

	// Get column types
	columnTypes := rows.ColumnTypes()
	columns := make([]string, len(columnTypes))
	for i, col := range columnTypes {
		columns[i] = col.Name()
	}

	// Scan results
	var data []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			c.logger.Printf("Error scanning row: %v", err)
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &ExploreResponse{
		Columns: columns,
		Data:    data,
		Total:   len(data),
	}, nil
}

// QueryRaw executes a raw SQL query and returns the results as a structured response
func (c *ClickHouseClient) QueryRaw(ctx context.Context, query string) ([]string, []map[string]interface{}, error) {
	c.logger.Printf("Executing raw query: %s", query)
	
	rows, err := c.conn.Query(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute raw query: %w", err)
	}
	defer rows.Close()
	
	// Get column types
	columnTypes := rows.ColumnTypes()
	columns := make([]string, len(columnTypes))
	for i, col := range columnTypes {
		columns[i] = col.Name()
	}
	
	// Scan results
	var data []map[string]interface{}
	for rows.Next() {
		// Create typed variables based on column types
		valuePtrs := make([]interface{}, len(columns))
		
		for i, colType := range columnTypes {
			typeName := colType.DatabaseTypeName()
			
			switch {
			case strings.HasPrefix(typeName, "DateTime64") || typeName == "DateTime":
				var t time.Time
				valuePtrs[i] = &t
			case typeName == "Date":
				var d time.Time
				valuePtrs[i] = &d
			case typeName == "String" || 
				 strings.HasPrefix(typeName, "FixedString") || 
				 typeName == "LowCardinality(String)":
				var s string
				valuePtrs[i] = &s
			case typeName == "Int8":
				var n int8
				valuePtrs[i] = &n
			case typeName == "Int16":
				var n int16
				valuePtrs[i] = &n
			case typeName == "Int32":
				var n int32
				valuePtrs[i] = &n
			case typeName == "Int64":
				var n int64
				valuePtrs[i] = &n
			case typeName == "UInt8":
				var n uint8
				valuePtrs[i] = &n
			case typeName == "UInt16":
				var n uint16
				valuePtrs[i] = &n
			case typeName == "UInt32":
				var n uint32
				valuePtrs[i] = &n
			case typeName == "UInt64":
				var n uint64
				valuePtrs[i] = &n
			case typeName == "Float32":
				var f float32
				valuePtrs[i] = &f
			case typeName == "Float64":
				var f float64
				valuePtrs[i] = &f
			case typeName == "Bool":
				var b bool
				valuePtrs[i] = &b
			case typeName == "UUID":
				var u string
				valuePtrs[i] = &u
			case strings.HasPrefix(typeName, "Array("):
				var arr string // We'll store arrays as JSON strings
				valuePtrs[i] = &arr
			case strings.HasPrefix(typeName, "Map("):
				var m map[string]string
				valuePtrs[i] = &m
			default:
				// For truly unknown types, try to scan as string first
				var s string
				valuePtrs[i] = &s
			}
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		
		row := make(map[string]interface{})
		for i, col := range columns {
			ptr := valuePtrs[i]
			var val interface{}
			
			// Extract the actual value from the pointer
			switch v := ptr.(type) {
			case *time.Time:
				if !(*v).IsZero() {
					val = (*v).Format(time.RFC3339)
				} else {
					val = nil
				}
			case *string:
				if *v != "" {
					val = *v
				} else {
					val = *v // Keep empty strings as empty strings, not nil
				}
			case *int8:
				val = int64(*v)
			case *int16:
				val = int64(*v)
			case *int32:
				val = int64(*v)
			case *int64:
				val = *v
			case *uint8:
				val = uint64(*v)
			case *uint16:
				val = uint64(*v)
			case *uint32:
				val = uint64(*v)
			case *uint64:
				val = *v
			case *float32:
				val = float64(*v)
			case *float64:
				val = *v
			case *bool:
				val = *v
			case *map[string]string:
				// Convert map to JSON for the response
				if *v != nil {
					val = *v
				} else {
					val = map[string]string{}
				}
			default:
				// This should not happen now, but keeping as safety net
				val = nil
			}
			
			row[col] = val
		}
		data = append(data, row)
	}
	
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating rows: %w", err)
	}
	
	return columns, data, nil
}