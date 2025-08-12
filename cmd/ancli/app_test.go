package main

import (
	"testing"
	"time"

	"github.com/justinlyon12/ancli/internal/config"
)

// TestNewApp tests the app creation and dependency injection
func TestNewApp(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Path: "/tmp/test_app.db",
				},
				Sandbox: config.SandboxConfig{
					Driver:         "podman",
					DefaultImage:   "alpine:3.18",
					DefaultTimeout: 30 * time.Second,
					NetworkEnabled: false,
				},
				LogLevel: "info",
				LogJSON:  false,
			},
			expectError: false,
		},
		{
			name: "invalid sandbox driver",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Path: "/tmp/test_app.db",
				},
				Sandbox: config.SandboxConfig{
					Driver: "invalid-driver",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := NewApp(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify app is properly initialized
			if app == nil {
				t.Fatal("App should not be nil")
			}
			if app.Config == nil {
				t.Error("App.Config should not be nil")
			}
			if app.Storage == nil {
				t.Error("App.Storage should not be nil")
			}
			if app.Scheduler == nil {
				t.Error("App.Scheduler should not be nil")
			}
			if app.Sandbox == nil {
				t.Error("App.Sandbox should not be nil")
			}
			if app.ReviewService == nil {
				t.Error("App.ReviewService should not be nil")
			}

			// Test cleanup
			if err := app.Close(); err != nil {
				t.Errorf("Failed to close app: %v", err)
			}
		})
	}
}

// TestConfigLoaders tests the ConfigLoader implementations
func TestConfigLoaders(t *testing.T) {
	t.Run("TestConfigLoader with valid config", func(t *testing.T) {
		testConfig := &config.Config{
			Database: config.DatabaseConfig{
				Path: "/tmp/test.db",
			},
			Sandbox: config.SandboxConfig{
				Driver: "podman",
			},
		}

		loader := &TestConfigLoader{Config: testConfig}
		cfg, err := loader.Load()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if cfg != testConfig {
			t.Error("Should return the same config instance")
		}
	})

	t.Run("TestConfigLoader with nil config", func(t *testing.T) {
		loader := &TestConfigLoader{Config: nil}
		_, err := loader.Load()

		if err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("DefaultConfigLoader", func(t *testing.T) {
		loader := &DefaultConfigLoader{}
		// We can't easily test this without a real config file,
		// so just verify it implements the interface
		var _ ConfigLoader = loader
	})
}

// TestCommandCreation tests the factory functions
func TestCommandCreation(t *testing.T) {
	testConfig := &config.Config{
		Database: config.DatabaseConfig{
			Path: "/tmp/test.db",
		},
		Sandbox: config.SandboxConfig{
			Driver:         "podman",
			DefaultImage:   "alpine:3.18",
			NetworkEnabled: false,
		},
		LogLevel: "info",
		LogJSON:  false,
	}

	loader := &TestConfigLoader{Config: testConfig}

	t.Run("NewRootCmd", func(t *testing.T) {
		cmd := NewRootCmd(loader)

		if cmd == nil {
			t.Fatal("Command should not be nil")
		}
		if cmd.Use != "ancli" {
			t.Errorf("Expected Use='ancli', got %q", cmd.Use)
		}
		if len(cmd.Commands()) == 0 {
			t.Error("Root command should have subcommands")
		}

		// Verify global flags are present
		if cmd.PersistentFlags().Lookup("config") == nil {
			t.Error("Missing --config flag")
		}
		if cmd.PersistentFlags().Lookup("log-level") == nil {
			t.Error("Missing --log-level flag")
		}
		if cmd.PersistentFlags().Lookup("database-path") == nil {
			t.Error("Missing --database-path flag")
		}
	})

	t.Run("NewReviewCmd", func(t *testing.T) {
		cmd := NewReviewCmd(loader)

		if cmd == nil {
			t.Fatal("Command should not be nil")
		}
		if cmd.Use != "review" {
			t.Errorf("Expected Use='review', got %q", cmd.Use)
		}

		// Verify review-specific flags are present
		if cmd.Flags().Lookup("deck-id") == nil {
			t.Error("Missing --deck-id flag")
		}
		if cmd.Flags().Lookup("max-cards") == nil {
			t.Error("Missing --max-cards flag")
		}
		if cmd.Flags().Lookup("no-network") == nil {
			t.Error("Missing --no-network flag")
		}
	})
}

// TestInitializeApp tests the app initialization function
func TestInitializeApp(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		testConfig := &config.Config{
			Database: config.DatabaseConfig{
				Path: "/tmp/test_init.db",
			},
			Sandbox: config.SandboxConfig{
				Driver:         "podman",
				DefaultImage:   "alpine:3.18",
				DefaultTimeout: 30 * time.Second,
				NetworkEnabled: false,
			},
			LogLevel: "info",
			LogJSON:  false,
		}

		loader := &TestConfigLoader{Config: testConfig}
		app, err := initializeApp(loader)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if app == nil {
			t.Error("App should not be nil")
		}

		// Cleanup
		if app != nil {
			if err := app.Close(); err != nil {
				t.Errorf("Failed to close app: %v", err)
			}
		}
	})

	t.Run("config load failure", func(t *testing.T) {
		loader := &TestConfigLoader{Config: nil}
		_, err := initializeApp(loader)

		if err == nil {
			t.Error("Expected error for failed config load")
		}
	})

	t.Run("app creation failure", func(t *testing.T) {
		badConfig := &config.Config{
			Sandbox: config.SandboxConfig{
				Driver: "invalid-driver",
			},
		}

		loader := &TestConfigLoader{Config: badConfig}
		_, err := initializeApp(loader)

		if err == nil {
			t.Error("Expected error for invalid sandbox driver")
		}
	})
}
