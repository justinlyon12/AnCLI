package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Database
	Database DatabaseConfig `mapstructure:"database"`

	// Sandbox
	Sandbox SandboxConfig `mapstructure:"sandbox"`

	// Review
	Review ReviewConfig `mapstructure:"review"`

	// Logging
	LogLevel string `mapstructure:"log_level"`
	LogJSON  bool   `mapstructure:"log_json"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// SandboxConfig holds sandbox-related configuration
type SandboxConfig struct {
	Driver         string        `mapstructure:"driver"`
	DefaultImage   string        `mapstructure:"default_image"`
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`
	NetworkEnabled bool          `mapstructure:"network_enabled"`
}

// ReviewConfig holds review session configuration
type ReviewConfig struct {
	MaxCardsPerSession int           `mapstructure:"max_cards_per_session"`
	SessionTimeout     time.Duration `mapstructure:"session_timeout"`
	AutoAdvance        bool          `mapstructure:"auto_advance"`
}

// Load reads configuration from files, environment variables, and flags
func Load() (*Config, error) {
	// Set defaults
	if err := setDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	// Set up config file search paths
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	viper.SetConfigName("ancli")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(home, ".ancli"))
	viper.AddConfigPath(".")

	// Enable environment variable support with proper key mapping
	viper.SetEnvPrefix("ANCLI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind specific environment variables for nested keys
	_ = viper.BindEnv("database.path", "ANCLI_DATABASE_PATH")
	_ = viper.BindEnv("sandbox.driver", "ANCLI_SANDBOX_DRIVER")
	_ = viper.BindEnv("sandbox.default_image", "ANCLI_SANDBOX_DEFAULT_IMAGE")
	_ = viper.BindEnv("sandbox.default_timeout", "ANCLI_SANDBOX_DEFAULT_TIMEOUT")
	_ = viper.BindEnv("sandbox.network_enabled", "ANCLI_SANDBOX_NETWORK_ENABLED")

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		// Only return error if config file exists but can't be read
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal into struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand paths
	config.Database.Path = expandPath(config.Database.Path)

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() error {
	// Database defaults
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory for defaults: %w", err)
	}

	defaultDBPath := filepath.Join(home, ".ancli", "ancli.db")
	viper.SetDefault("database.path", defaultDBPath)

	// Sandbox defaults
	viper.SetDefault("sandbox.driver", "podman")
	viper.SetDefault("sandbox.default_image", "alpine:3.18")
	viper.SetDefault("sandbox.default_timeout", "30s")
	viper.SetDefault("sandbox.network_enabled", false)

	// Review defaults
	viper.SetDefault("review.max_cards_per_session", 20)
	viper.SetDefault("review.session_timeout", "30m")
	viper.SetDefault("review.auto_advance", false)

	// Logging defaults
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_json", false)

	return nil
}

// expandPath expands ~ and environment variables in paths
func expandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = home
		}
	}

	return path
}

// GetDatabasePath returns the database file path, creating directories if needed
func (c *Config) GetDatabasePath() (string, error) {
	dbPath := c.Database.Path
	dir := filepath.Dir(dbPath)

	// Create directory with secure permissions (0700 for user data)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create database directory %s: %w", dir, err)
	}

	return dbPath, nil
}
