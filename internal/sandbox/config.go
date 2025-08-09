package sandbox

import (
	"fmt"
	"time"
)

// ExecutionConfig defines the parameters for sandboxed command execution
type ExecutionConfig struct {
	// Container configuration
	Image      string // Set from deck default or card override
	Command    []string
	WorkingDir string

	// Environment and variables
	Environment map[string]string

	// Security settings
	NetworkEnabled bool
	Capabilities   []string
	ReadOnlyRootFS bool
	TmpfsMounts    map[string]string

	// Resource limits
	Timeout     time.Duration // Per-command timeout, not container lifetime
	MemoryLimit string        // e.g., "128m"
	CPULimit    string        // e.g., "0.5"

	// Tracing and logging
	CorrelationID string
}

// ContainerLifecycle defines how containers are managed across card executions
type ContainerLifecycle string

const (
	// PerCard creates a new container for each card (high isolation, slow)
	PerCard ContainerLifecycle = "per-card"

	// SessionReuse reuses container across cards in a review session (balanced)
	SessionReuse ContainerLifecycle = "session-reuse"

	// DeckPersistent keeps container running for entire deck (fast, lower isolation)
	DeckPersistent ContainerLifecycle = "deck-persistent"
)

// NewExecutionConfig returns a secure-by-default configuration
// Image must be provided from deck/card configuration
func NewExecutionConfig() ExecutionConfig {
	return ExecutionConfig{
		Image:          "", // No default - set from deck/card
		WorkingDir:     "/tmp",
		Environment:    make(map[string]string),
		NetworkEnabled: false,
		Capabilities:   []string{}, // Empty = drop all capabilities
		ReadOnlyRootFS: true,
		TmpfsMounts: map[string]string{
			"/tmp": "rw,noexec,nosuid,size=100m",
		},
		Timeout: 30 * time.Second, // Per-command timeout, not container lifetime
	}
}

// WithImage sets the container image (from deck default or card override)
func (c ExecutionConfig) WithImage(image string) ExecutionConfig {
	c.Image = image
	return c
}

// WithCommand sets the command to execute
func (c ExecutionConfig) WithCommand(cmd ...string) ExecutionConfig {
	c.Command = cmd
	return c
}

// WithNetworking enables or disables network access
func (c ExecutionConfig) WithNetworking(enabled bool) ExecutionConfig {
	c.NetworkEnabled = enabled
	return c
}

// WithTimeout sets the per-command execution timeout
func (c ExecutionConfig) WithTimeout(timeout time.Duration) ExecutionConfig {
	c.Timeout = timeout
	return c
}

// WithCorrelationID sets the correlation ID for tracing
func (c ExecutionConfig) WithCorrelationID(id string) ExecutionConfig {
	c.CorrelationID = id
	return c
}

// Validate checks that required fields are set
func (c ExecutionConfig) Validate() error {
	if c.Image == "" {
		return fmt.Errorf("image is required")
	}
	if len(c.Command) == 0 {
		return fmt.Errorf("command is required")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	return nil
}
