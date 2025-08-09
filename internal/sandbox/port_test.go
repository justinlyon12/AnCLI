package sandbox

import (
	"context"
	"testing"
	"time"
)

// TestSandboxInterface verifies the contract of the Sandbox interface
func TestSandboxInterface(t *testing.T) {
	// This test ensures our interface is well-defined
	// Real implementations are tested in their respective packages
	var _ Sandbox = (*mockSandbox)(nil)
}

// TestNewExecutionConfig verifies secure defaults
func TestNewExecutionConfig(t *testing.T) {
	config := NewExecutionConfig()

	// Security defaults
	if config.NetworkEnabled {
		t.Error("default config should disable networking")
	}

	if !config.ReadOnlyRootFS {
		t.Error("default config should use read-only root filesystem")
	}

	if len(config.Capabilities) != 0 {
		t.Error("default config should drop all capabilities")
	}

	// Resource defaults
	if config.Timeout != 30*time.Second {
		t.Errorf("default timeout should be 30s, got %v", config.Timeout)
	}

	if config.Image != "" {
		t.Errorf("default image should be empty (must be set from deck), got %s", config.Image)
	}

	// Tmpfs mounts
	tmpfsOpts, exists := config.TmpfsMounts["/tmp"]
	if !exists {
		t.Error("default config should include /tmp tmpfs mount")
	}

	expectedTmpfs := "rw,noexec,nosuid,size=100m"
	if tmpfsOpts != expectedTmpfs {
		t.Errorf("default /tmp tmpfs options should be %s, got %s", expectedTmpfs, tmpfsOpts)
	}
}

// TestExecutionConfigBuilder tests the fluent builder methods
func TestExecutionConfigBuilder(t *testing.T) {
	config := NewExecutionConfig().
		WithImage("ubuntu:20.04").
		WithCommand("echo", "test").
		WithNetworking(true).
		WithTimeout(10 * time.Second).
		WithCorrelationID("test-123")

	if config.Image != "ubuntu:20.04" {
		t.Errorf("expected image ubuntu:20.04, got %s", config.Image)
	}

	expectedCmd := []string{"echo", "test"}
	if len(config.Command) != 2 || config.Command[0] != "echo" || config.Command[1] != "test" {
		t.Errorf("expected command %v, got %v", expectedCmd, config.Command)
	}

	if !config.NetworkEnabled {
		t.Error("expected networking to be enabled")
	}

	if config.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", config.Timeout)
	}

	if config.CorrelationID != "test-123" {
		t.Errorf("expected correlation ID test-123, got %s", config.CorrelationID)
	}
}

// TestExecutionConfigValidate tests validation logic
func TestExecutionConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    ExecutionConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: NewExecutionConfig().
				WithImage("alpine:latest").
				WithCommand("echo", "hello"),
			wantError: false,
		},
		{
			name: "missing image",
			config: NewExecutionConfig().
				WithCommand("echo", "hello"),
			wantError: true,
			errorMsg:  "image is required",
		},
		{
			name: "missing command",
			config: NewExecutionConfig().
				WithImage("alpine:latest"),
			wantError: true,
			errorMsg:  "command is required",
		},
		{
			name: "zero timeout",
			config: ExecutionConfig{
				Image:   "alpine:latest",
				Command: []string{"echo", "hello"},
				Timeout: 0,
			},
			wantError: true,
			errorMsg:  "timeout must be positive",
		},
		{
			name: "negative timeout",
			config: ExecutionConfig{
				Image:   "alpine:latest",
				Command: []string{"echo", "hello"},
				Timeout: -1 * time.Second,
			},
			wantError: true,
			errorMsg:  "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				if err == nil {
					t.Error("expected validation error, got nil")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

// TestContainerLifecycleConstants verifies lifecycle constants
func TestContainerLifecycleConstants(t *testing.T) {
	if PerCard != "per-card" {
		t.Errorf("expected PerCard to be 'per-card', got %s", PerCard)
	}
	if SessionReuse != "session-reuse" {
		t.Errorf("expected SessionReuse to be 'session-reuse', got %s", SessionReuse)
	}
	if DeckPersistent != "deck-persistent" {
		t.Errorf("expected DeckPersistent to be 'deck-persistent', got %s", DeckPersistent)
	}
}

// mockSandbox is a test implementation of the Sandbox interface
type mockSandbox struct{}

func (m *mockSandbox) Run(ctx context.Context, config ExecutionConfig) (*ExecutionResult, error) {
	return &ExecutionResult{
		ExitCode:      0,
		Success:       true,
		Stdout:        "mock output",
		Stderr:        "",
		StartedAt:     time.Now(),
		Duration:      100 * time.Millisecond,
		ImageUsed:     config.Image,
		CorrelationID: config.CorrelationID,
	}, nil
}

func (m *mockSandbox) Cleanup(ctx context.Context) error {
	return nil
}

func (m *mockSandbox) Name() string {
	return "mock"
}
