package main

import (
	"context"
	"fmt"

	"github.com/justinlyon12/ancli/internal/config"
	"github.com/justinlyon12/ancli/internal/review"
	"github.com/justinlyon12/ancli/internal/sandbox"
	"github.com/justinlyon12/ancli/internal/sandbox/podman"
	"github.com/justinlyon12/ancli/internal/scheduler"
	"github.com/justinlyon12/ancli/internal/storage"
)

// App holds all application dependencies and configuration
type App struct {
	Config        *config.Config
	Storage       storage.Storage
	Scheduler     *scheduler.Scheduler
	Sandbox       sandbox.Sandbox
	ReviewService *review.Service
}

// NewApp creates a new application with all dependencies wired up
func NewApp(cfg *config.Config) (*App, error) {
	app := &App{
		Config: cfg,
	}

	// Initialize storage
	dbPath, err := cfg.GetDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	app.Storage, err = storage.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize scheduler
	app.Scheduler = scheduler.NewScheduler()

	// Initialize sandbox based on config
	switch cfg.Sandbox.Driver {
	case "podman":
		app.Sandbox, err = podman.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create podman driver: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported sandbox driver: %s", cfg.Sandbox.Driver)
	}

	// Initialize review service
	app.ReviewService = review.NewService(app.Storage, app.Scheduler, app.Sandbox)

	return app, nil
}

// Close cleans up application resources
func (a *App) Close() error {
	var errs []error

	// Close storage
	if a.Storage != nil {
		if closer, ok := a.Storage.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close storage: %w", err))
			}
		}
	}

	// Cleanup sandbox
	if a.Sandbox != nil {
		if err := a.Sandbox.Cleanup(context.Background()); err != nil {
			errs = append(errs, fmt.Errorf("failed to cleanup sandbox: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during app cleanup: %v", errs)
	}

	return nil
}

// ConfigLoader defines how configuration is loaded
type ConfigLoader interface {
	Load() (*config.Config, error)
}

// DefaultConfigLoader loads config using the default Viper-based method
type DefaultConfigLoader struct{}

// Load implements ConfigLoader using the existing config.Load() function
func (l *DefaultConfigLoader) Load() (*config.Config, error) {
	return config.Load()
}

// TestConfigLoader allows injecting pre-built configuration for tests
type TestConfigLoader struct {
	Config *config.Config
}

// Load implements ConfigLoader by returning the pre-built config
func (l *TestConfigLoader) Load() (*config.Config, error) {
	if l.Config == nil {
		return nil, fmt.Errorf("no config provided to TestConfigLoader")
	}
	return l.Config, nil
}
