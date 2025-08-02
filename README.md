# ObservIO Backend

ObservIO is a modern observability platform backend built with Go. It provides APIs for monitoring metrics, creating dashboards, managing alerts, and connecting to various data sources.

## Features

- **Metrics Collection & Visualization**: Collect and visualize metrics from various sources
- **Dashboards**: Create and manage custom dashboards with multiple visualization panels
- **Alerting**: Define alert rules and manage alert notifications
- **Data Sources**: Connect to various data sources like Prometheus, Elasticsearch, and Jaeger
- **Log Management**: Real-time log exploration and filtering powered by ClickHouse

## Project Structure

```
├── cmd/            # Application entry points
│   └── server/     # Main server application
├── api/            # API specifications and documentation (empty)
├── internal/       # Private application code
│   ├── api/        # API handlers and routing implementation
│   │   └── handlers/ # API endpoint handlers
│   ├── config/     # Configuration management
│   └── database/   # Database clients (ClickHouse)
├── pkg/            # Public libraries that can be used by external applications
├── config/         # Configuration files
└── docs/           # Documentation and data files
```

## Getting Started

### Prerequisites

- Go 1.23 or higher
- ClickHouse server (for log data storage)
- OpenTelemetry Collector (configured to export logs to ClickHouse)

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/observio.git
   cd observio/backend
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Run the server:
   ```
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080` by default.

## API Endpoints

### Metrics
- `GET /api/v1/metrics` - List available metrics
- `POST /api/v1/metrics/query` - Query metrics data
- `GET /api/v1/metrics/{name}` - Get specific metric

### Dashboards
- `GET /api/v1/dashboards` - List dashboards
- `POST /api/v1/dashboards` - Create dashboard
- `GET /api/v1/dashboards/{id}` - Get dashboard
- `PUT /api/v1/dashboards/{id}` - Update dashboard
- `DELETE /api/v1/dashboards/{id}` - Delete dashboard

### Alerts
- `GET /api/v1/alerts` - List alerts
- `GET /api/v1/alerts/{id}` - Get alert
- `PUT /api/v1/alerts/{id}/resolve` - Resolve alert
- `GET /api/v1/alerts/rules` - List alert rules
- `POST /api/v1/alerts/rules` - Create alert rule
- `GET /api/v1/alerts/rules/{id}` - Get alert rule
- `PUT /api/v1/alerts/rules/{id}` - Update alert rule
- `DELETE /api/v1/alerts/rules/{id}` - Delete alert rule

### Data Sources
- `GET /api/v1/datasources` - List data sources
- `POST /api/v1/datasources` - Create data source
- `GET /api/v1/datasources/{id}` - Get data source
- `PUT /api/v1/datasources/{id}` - Update data source
- `DELETE /api/v1/datasources/{id}` - Delete data source
- `POST /api/v1/datasources/{id}/test` - Test data source connection

### Logs
- `GET /api/v1/logs` - Query logs with filtering (supports ?level, ?component, ?pattern, ?limit, ?offset)
- `GET /api/v1/logs/top100` - Get the 100 most recent log entries

## Data Sources Configuration

### ClickHouse Setup

The backend connects to ClickHouse for log data storage. The connection configuration is:

- **Host**: 192.168.1.2:9000 (configured in yml/gateway.yaml)
- **Database**: default
- **Credentials**: default/shiva1712
- **Table**: otel_logs (created by OpenTelemetry Collector)

To set up ClickHouse:

1. Install ClickHouse server:
   ```bash
   curl https://clickhouse.com/ | sh
   sudo ./clickhouse install
   ```

2. Start the server:
   ```bash
   sudo clickhouse start
   ```

3. Connect with client:
   ```bash
   clickhouse-client --password --host=192.168.1.2 --port=9000
   ```

### OpenTelemetry Collector

The system uses OpenTelemetry Collector to ingest logs into ClickHouse. Configuration is located in `yml/gateway.yaml` with:

- OTLP receiver on port 4318
- ClickHouse exporter configured for otel_logs table
- Batch processing for performance

## Implementation Details

### Log Data Structure

The backend queries ClickHouse's `otel_logs` table with the following mapped fields:

```sql
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
```

### Database Client

Located in `internal/database/clickhouse.go`, the ClickHouse client provides:
- Connection management with connection pooling
- Query methods for log retrieval with filtering
- Error handling and logging
- Support for pagination and search parameters

## Configuration

Configuration is loaded from `config/config.yaml` by default. You can specify a different configuration file using the `-config` flag.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
