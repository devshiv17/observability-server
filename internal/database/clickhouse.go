package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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