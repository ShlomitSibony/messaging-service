package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"messaging-service/internal/config"
	"messaging-service/internal/container"
	"messaging-service/internal/logger"
	"messaging-service/internal/router"
	"messaging-service/internal/telemetry"

	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"
)

// App represents the application instance
type App struct {
	config    *config.Config
	container *container.Container
	server    *http.Server
	logger    *zap.Logger
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) *App {
	return &App{
		config: cfg,
		logger: logger.Get(),
	}
}

// Initialize sets up all dependencies and connections
func (a *App) Initialize() error {
	a.logger.Info("Initializing application")

	// Initialize telemetry
	if err := telemetry.InitTelemetry(a.logger); err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Connect to database
	db, err := a.connectDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize dependency container
	a.container = container.NewContainer(a.config, db)

	// Setup router
	router := a.setupRouter()

	// Create server
	a.server = &http.Server{
		Addr:         ":" + a.config.Server.Port,
		Handler:      router,
		ReadTimeout:  a.config.Server.ReadTimeout,
		WriteTimeout: a.config.Server.WriteTimeout,
		IdleTimeout:  a.config.Server.IdleTimeout,
	}

	a.logger.Info("Application initialized successfully")
	return nil
}

// Start starts the application server
func (a *App) Start() error {
	a.logger.Info("Starting server", zap.String("port", a.config.Server.Port))
	return a.server.ListenAndServe()
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down application")

	// Shutdown telemetry
	if err := telemetry.Shutdown(ctx); err != nil {
		a.logger.Error("Failed to shutdown telemetry", zap.Error(err))
	}

	// Close container resources
	if a.container != nil {
		if err := a.container.Close(); err != nil {
			a.logger.Error("Failed to close container resources", zap.Error(err))
		}
	}

	// Shutdown server
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	a.logger.Info("Application shutdown complete")
	return nil
}

// connectDatabase establishes database connection
func (a *App) connectDatabase() (*sql.DB, error) {
	db, err := sql.Open("postgres", a.config.Database.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(a.config.Database.MaxOpenConns)
	db.SetMaxIdleConns(a.config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(a.config.Database.ConnMaxLifetime)

	a.logger.Info("Database connected successfully")
	return db, nil
}

// setupRouter configures the HTTP router with all middleware and routes
func (a *App) setupRouter() http.Handler {
	// Create router
	router := router.NewRouter()

	// Setup routes with handlers from container
	router.SetupRoutes(a.container.MessagingHandler, a.logger)

	return router.GetEngine()
}
