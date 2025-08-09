package podman

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/justinlyon12/ancli/internal/sandbox"
)

// Driver implements the sandbox interface using Podman
// Supports both per-card and session-reuse container lifecycles
type Driver struct {
	podmanPath string
	lifecycle  sandbox.ContainerLifecycle

	// Container state (protected by mutex)
	mu            sync.Mutex
	containerID   string
	containerName string
}

// init registers the Podman driver with the sandbox registry
func init() {
	sandbox.Register("podman", func() (sandbox.Sandbox, error) {
		return New()
	})
}

// New creates a new Podman driver instance
func New() (*Driver, error) {
	// Check if podman is available
	podmanPath, err := exec.LookPath("podman")
	if err != nil {
		return nil, fmt.Errorf("podman not found in PATH: %w", err)
	}

	return &Driver{
		podmanPath: podmanPath,
		lifecycle:  sandbox.SessionReuse, // Default to session-reuse for performance
	}, nil
}

// Name returns the driver identifier
func (d *Driver) Name() string {
	return "podman"
}

// Run executes a command in a Podman container
// Uses session-reuse by default for performance
func (d *Driver) Run(ctx context.Context, config sandbox.ExecutionConfig) (*sandbox.ExecutionResult, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	startTime := time.Now()
	logger := slog.With(
		"correlation_id", config.CorrelationID,
		"driver", "podman",
		"image", config.Image,
		"lifecycle", d.lifecycle,
	)

	switch d.lifecycle {
	case sandbox.PerCard:
		return d.runPerCard(ctx, config, logger, startTime)
	case sandbox.SessionReuse:
		return d.runSessionReuse(ctx, config, logger, startTime)
	default:
		return nil, fmt.Errorf("unsupported lifecycle mode: %s", d.lifecycle)
	}
}

// runPerCard creates a new container for each command (MVP: not implemented)
func (d *Driver) runPerCard(ctx context.Context, config sandbox.ExecutionConfig, logger *slog.Logger, startTime time.Time) (*sandbox.ExecutionResult, error) {
	// For MVP, we'll focus on session-reuse
	// This would use `podman run --rm` for each command
	return nil, fmt.Errorf("per-card lifecycle not implemented in MVP")
}

// runSessionReuse reuses a container across multiple commands in a session
func (d *Driver) runSessionReuse(ctx context.Context, config sandbox.ExecutionConfig, logger *slog.Logger, startTime time.Time) (*sandbox.ExecutionResult, error) {
	// Ensure we have a running container for this session
	if err := d.ensureContainer(ctx, config, logger); err != nil {
		return nil, fmt.Errorf("failed to ensure container: %w", err)
	}

	// Execute command in the running container
	return d.execInContainer(ctx, config, logger, startTime)
}

// ensureContainer starts a container if one isn't already running
func (d *Driver) ensureContainer(ctx context.Context, config sandbox.ExecutionConfig, logger *slog.Logger) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.containerID != "" {
		// Check if container is running (not just exists)
		checkCmd := exec.CommandContext(ctx, d.podmanPath, "container", "inspect", d.containerID, "--format", "{{.State.Running}}")
		var output bytes.Buffer
		checkCmd.Stdout = &output

		if checkCmd.Run() == nil && strings.TrimSpace(output.String()) == "true" {
			logger.Debug("reusing existing running container", "container_id", d.containerID)
			return nil
		}

		// Container doesn't exist or isn't running, clear the state
		logger.Debug("container not running, starting new one", "old_container_id", d.containerID)
		d.containerID = ""
		d.containerName = ""
	}

	// Generate a unique container name
	d.containerName = fmt.Sprintf("ancli-session-%d", time.Now().UnixNano())

	// Build podman run command for long-running container
	args := []string{
		"run",
		"--detach",                // Run in background
		"--name", d.containerName, // Named container for reuse
		"--cap-drop=ALL",                             // Drop all capabilities
		"--security-opt=no-new-privileges",           // Prevent privilege escalation
		"--read-only",                                // Read-only root filesystem
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=100m", // Writable /tmp with security
	}

	// Add network configuration
	if !config.NetworkEnabled {
		args = append(args, "--network=none")
	}

	// Add initial working directory (can be overridden per-exec)
	if config.WorkingDir != "" {
		args = append(args, "--workdir", config.WorkingDir)
	}

	// Add environment variables
	for key, value := range config.Environment {
		args = append(args, "--env", fmt.Sprintf("%s=%s", key, value))
	}

	// Add resource limits
	if config.MemoryLimit != "" {
		args = append(args, "--memory", config.MemoryLimit)
	}
	if config.CPULimit != "" {
		args = append(args, "--cpus", config.CPULimit)
	}

	// Add image and keep-alive command
	args = append(args, config.Image, "sleep", "3600") // Keep container alive for 1 hour

	logger.Debug("starting session container", "args", args)

	// Start the container - capture stdout only for container ID
	cmd := exec.CommandContext(ctx, d.podmanPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %w, stderr: %s", err, stderr.String())
	}

	// Extract container ID from stdout only (ignore stderr warnings)
	d.containerID = strings.TrimSpace(stdout.String())
	if d.containerID == "" {
		return fmt.Errorf("failed to get container ID from stdout")
	}

	logger.Info("started session container", "container_id", d.containerID, "name", d.containerName)
	return nil
}

// execInContainer executes a command in the running session container
func (d *Driver) execInContainer(ctx context.Context, config sandbox.ExecutionConfig, logger *slog.Logger, startTime time.Time) (*sandbox.ExecutionResult, error) {
	d.mu.Lock()
	containerID := d.containerID
	d.mu.Unlock()

	// Build podman exec command
	args := []string{"exec"}

	// Add working directory for this specific command
	if config.WorkingDir != "" {
		args = append(args, "--workdir", config.WorkingDir)
	}

	// Add environment variables for this specific command
	for key, value := range config.Environment {
		args = append(args, "--env", fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args, containerID)
	args = append(args, config.Command...)

	// Create context with command timeout
	execCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	logger.Debug("executing command in container", "command", config.Command, "workdir", config.WorkingDir)

	// Execute the command
	cmd := exec.CommandContext(execCtx, d.podmanPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime)

	// Extract exit code
	exitCode := 0
	success := err == nil

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else {
			// Non-exit error (e.g., timeout)
			exitCode = -1
		}
	}

	result := &sandbox.ExecutionResult{
		ExitCode:      exitCode,
		Success:       success,
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		StartedAt:     startTime,
		Duration:      duration,
		ContainerID:   containerID,
		ImageUsed:     config.Image,
		CorrelationID: config.CorrelationID,
	}

	logger.Info("command execution completed",
		"exit_code", exitCode,
		"success", success,
		"duration_ms", duration.Milliseconds(),
		"stdout_bytes", len(result.Stdout),
		"stderr_bytes", len(result.Stderr),
	)

	// Wrap execution errors with context
	if err != nil && exitCode == -1 {
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	return result, nil
}

// Cleanup stops and removes the session container
func (d *Driver) Cleanup(ctx context.Context) error {
	d.mu.Lock()
	containerID := d.containerID
	containerName := d.containerName
	d.containerID = ""
	d.containerName = ""
	d.mu.Unlock()

	if containerID == "" {
		return nil // Nothing to clean up
	}

	logger := slog.With("container_id", containerID, "driver", "podman")
	logger.Debug("cleaning up session container")

	// Stop and remove the container
	stopCmd := exec.CommandContext(ctx, d.podmanPath, "container", "stop", containerID)
	if err := stopCmd.Run(); err != nil {
		logger.Warn("failed to stop container", "error", err)
	}

	rmCmd := exec.CommandContext(ctx, d.podmanPath, "container", "rm", containerID)
	if err := rmCmd.Run(); err != nil {
		logger.Warn("failed to remove container", "error", err)
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}

	logger.Info("session container cleaned up", "container_name", containerName)
	return nil
}

// IsAvailable checks if Podman is available and functional
func IsAvailable() error {
	_, err := exec.LookPath("podman")
	if err != nil {
		return fmt.Errorf("podman not found: %w", err)
	}

	// Test basic podman functionality
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "podman", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman not functional: %w", err)
	}

	return nil
}
