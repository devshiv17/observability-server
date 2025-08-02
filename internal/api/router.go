package api

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/observio/backend/internal/api/handlers"
	"github.com/observio/backend/internal/config"
	"github.com/observio/backend/internal/database"
)


// NewRouter creates and configures a new HTTP router
func NewRouter(cfg *config.Config, logger *log.Logger) http.Handler {
	r := chi.NewRouter()

	// Initialize ClickHouse client
	clickhouseClient, err := database.NewClickHouseClient(
		"192.168.1.2", // host from gateway.yaml
		9000,          // port from gateway.yaml
		"default",     // username from gateway.yaml
		"shiva1712",   // password from gateway.yaml
		"default",     // database
		logger,
	)
	if err != nil {
		logger.Printf("Warning: Failed to connect to ClickHouse: %v. Logs endpoint may not work properly.", err)
		clickhouseClient = nil
	}

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second))
	// Allow both /logs and /logs/ (and similar) to work
	r.Use(middleware.StripSlashes)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Adjust for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Metrics endpoints
		r.Mount("/metrics", handlers.NewMetricsHandler(cfg, logger))

		// Dashboard endpoints
		r.Mount("/dashboards", handlers.NewDashboardHandler(cfg, logger))

		// Alerts endpoints
		r.Mount("/alerts", handlers.NewAlertsHandler(cfg, logger))

		// Data sources endpoints
		r.Mount("/datasources", handlers.NewDataSourceHandler(cfg, logger))

		// Logs exploration endpoint (ClickHouse-based)
		if clickhouseClient != nil {
			r.Mount("/logs", handlers.NewLogsHandler(cfg, logger, clickhouseClient))
			r.Mount("/explore", handlers.NewExploreHandler(cfg, logger, clickhouseClient))
		} else {
			logger.Printf("Warning: ClickHouse client not available, logs and explore endpoints disabled")
		}
	})

	// Cleanup function for ClickHouse client could be added here if needed
	// For now, we'll let the connection be cleaned up when the program exits
	
	return r
}
