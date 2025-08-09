package sandbox

import (
	"context"
	"time"
)

// Sandbox defines the interface for container execution backends
type Sandbox interface {
	// Run executes a command in a sandboxed environment
	Run(ctx context.Context, config ExecutionConfig) (*ExecutionResult, error)

	// Cleanup performs any necessary cleanup operations
	Cleanup(ctx context.Context) error

	// Name returns the driver name for logging and identification
	Name() string
}

// ExecutionResult contains the output and metadata from command execution
type ExecutionResult struct {
	// Exit status
	ExitCode int
	Success  bool

	// Command output
	Stdout string
	Stderr string

	// Timing information
	StartedAt time.Time
	Duration  time.Duration

	// Execution metadata
	ContainerID   string
	ImageUsed     string
	CorrelationID string
}
