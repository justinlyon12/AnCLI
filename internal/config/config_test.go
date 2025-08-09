package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any existing viper state
	viper.Reset()

	config, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading default config, got: %v", err)
	}

	if config == nil {
		t.Fatal("config should not be nil")
	}

	// Test database defaults
	if config.Database.Path == "" {
		t.Error("database path should have default value")
	}

	// Test sandbox defaults
	if config.Sandbox.Driver != "podman" {
		t.Errorf("expected default driver 'podman', got: %s", config.Sandbox.Driver)
	}

	if config.Sandbox.DefaultImage != "alpine:3.18" {
		t.Errorf("expected default image 'alpine:3.18', got: %s", config.Sandbox.DefaultImage)
	}

	if config.Sandbox.DefaultTimeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got: %v", config.Sandbox.DefaultTimeout)
	}

	if config.Sandbox.NetworkEnabled {
		t.Error("expected network disabled by default")
	}

	// Test logging defaults
	if config.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got: %s", config.LogLevel)
	}

	if config.LogJSON {
		t.Error("expected JSON logging disabled by default")
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Clear any existing viper state
	viper.Reset()

	// Set environment variables
	os.Setenv("ANCLI_DATABASE_PATH", "/tmp/test.db")
	os.Setenv("ANCLI_SANDBOX_DRIVER", "docker")
	os.Setenv("ANCLI_SANDBOX_DEFAULT_IMAGE", "ubuntu:20.04")
	defer func() {
		os.Unsetenv("ANCLI_DATABASE_PATH")
		os.Unsetenv("ANCLI_SANDBOX_DRIVER")
		os.Unsetenv("ANCLI_SANDBOX_DEFAULT_IMAGE")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config with env vars, got: %v", err)
	}

	if config.Database.Path != "/tmp/test.db" {
		t.Errorf("expected database path from env var, got: %s", config.Database.Path)
	}

	if config.Sandbox.Driver != "docker" {
		t.Errorf("expected driver from env var, got: %s", config.Sandbox.Driver)
	}

	if config.Sandbox.DefaultImage != "ubuntu:20.04" {
		t.Errorf("expected image from env var, got: %s", config.Sandbox.DefaultImage)
	}
}

func TestGetDatabasePath(t *testing.T) {
	// Clear any existing viper state
	viper.Reset()

	config, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	dbPath, err := config.GetDatabasePath()
	if err != nil {
		t.Errorf("expected no error getting database path, got: %v", err)
	}

	if dbPath == "" {
		t.Error("database path should not be empty")
	}

	// Check that directory was created
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("database directory should be created: %s", dir)
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(string) bool
	}{
		{
			name:  "empty path",
			input: "",
			check: func(result string) bool { return result == "" },
		},
		{
			name:  "absolute path",
			input: "/tmp/test",
			check: func(result string) bool { return result == "/tmp/test" },
		},
		{
			name:  "tilde expansion",
			input: "~/test",
			check: func(result string) bool { return result != "~/test" && filepath.IsAbs(result) },
		},
		{
			name:  "environment variable",
			input: "$HOME/test",
			check: func(result string) bool { return result != "$HOME/test" && filepath.IsAbs(result) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if !tt.check(result) {
				t.Errorf("expandPath(%s) = %s, check failed", tt.input, result)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	// Clear any existing viper state
	viper.Reset()

	err := setDefaults()
	if err != nil {
		t.Errorf("expected no error setting defaults, got: %v", err)
	}

	// Check that viper has the expected default values
	if viper.GetString("sandbox.driver") != "podman" {
		t.Error("default sandbox driver not set correctly")
	}

	if viper.GetString("sandbox.default_image") != "alpine:3.18" {
		t.Error("default sandbox image not set correctly")
	}

	if viper.GetString("log_level") != "info" {
		t.Error("default log level not set correctly")
	}

	if viper.GetBool("sandbox.network_enabled") {
		t.Error("network should be disabled by default")
	}
}
