//go:build integration

package podman

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/justinlyon12/ancli/internal/sandbox"
)

func TestPodmanIntegration(t *testing.T) {
	// Skip if podman is not available
	if err := IsAvailable(); err != nil {
		t.Skipf("podman not available: %v", err)
	}

	driver, err := New()
	if err != nil {
		t.Fatalf("failed to create podman driver: %v", err)
	}

	// Cleanup any existing containers after test
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		driver.Cleanup(ctx)
	}()

	tests := []struct {
		name           string
		config         sandbox.ExecutionConfig
		expectedStdout string
		expectedCode   int
		shouldSucceed  bool
	}{
		{
			name: "simple echo command",
			config: sandbox.NewExecutionConfig().
				WithImage("alpine:latest").
				WithCommand("echo", "hello world").
				WithCorrelationID("test-echo"),
			expectedStdout: "hello world\n",
			expectedCode:   0,
			shouldSucceed:  true,
		},
		{
			name: "command with exit code",
			config: sandbox.NewExecutionConfig().
				WithImage("alpine:latest").
				WithCommand("sh", "-c", "exit 42").
				WithCorrelationID("test-exit-code"),
			expectedStdout: "",
			expectedCode:   42,
			shouldSucceed:  false,
		},
		{
			name: "working directory test",
			config: sandbox.NewExecutionConfig().
				WithImage("alpine:latest").
				WithCommand("pwd").
				WithCorrelationID("test-workdir"),
			expectedStdout: "/tmp\n",
			expectedCode:   0,
			shouldSucceed:  true,
		},
		{
			name: "environment variables",
			config: func() sandbox.ExecutionConfig {
				c := sandbox.NewExecutionConfig().
					WithImage("alpine:latest").
					WithCommand("sh", "-c", "echo $TEST_VAR").
					WithCorrelationID("test-env")
				c.Environment = map[string]string{"TEST_VAR": "test-value"}
				return c
			}(),
			expectedStdout: "test-value\n",
			expectedCode:   0,
			shouldSucceed:  true,
		},
		{
			name: "tmpfs mount verification",
			config: sandbox.NewExecutionConfig().
				WithImage("alpine:latest").
				WithCommand("sh", "-c", "echo test > /tmp/test.txt && cat /tmp/test.txt").
				WithCorrelationID("test-tmpfs"),
			expectedStdout: "test\n",
			expectedCode:   0,
			shouldSucceed:  true,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := driver.Run(ctx, tt.config)

			// Check execution error
			if tt.shouldSucceed && err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
			if !tt.shouldSucceed && tt.expectedCode > 0 && err != nil {
				// For non-zero exit codes, we might get an error, but result should still be returned
				if result == nil {
					t.Fatalf("expected result even with error, got nil")
				}
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			// Check exit code
			if result.ExitCode != tt.expectedCode {
				t.Errorf("expected exit code %d, got %d", tt.expectedCode, result.ExitCode)
			}

			// Check success flag
			if result.Success != tt.shouldSucceed {
				t.Errorf("expected success %v, got %v", tt.shouldSucceed, result.Success)
			}

			// Check stdout
			if result.Stdout != tt.expectedStdout {
				t.Errorf("expected stdout %q, got %q", tt.expectedStdout, result.Stdout)
			}

			// Verify result metadata
			if result.ImageUsed != tt.config.Image {
				t.Errorf("expected image %s, got %s", tt.config.Image, result.ImageUsed)
			}

			if result.CorrelationID != tt.config.CorrelationID {
				t.Errorf("expected correlation ID %s, got %s", tt.config.CorrelationID, result.CorrelationID)
			}

			if result.Duration <= 0 {
				t.Error("expected positive duration")
			}

			if result.StartedAt.IsZero() {
				t.Error("expected non-zero start time")
			}
		})
	}
}

func TestPodmanSessionReuse(t *testing.T) {
	// Skip if podman is not available
	if err := IsAvailable(); err != nil {
		t.Skipf("podman not available: %v", err)
	}

	driver, err := New()
	if err != nil {
		t.Fatalf("failed to create podman driver: %v", err)
	}

	// Cleanup after test
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		driver.Cleanup(ctx)
	}()

	ctx := context.Background()
	config := sandbox.NewExecutionConfig().
		WithImage("alpine:latest").
		WithCorrelationID("test-reuse")

	// First command - this will start a new container
	config1 := config.WithCommand("echo", "first")
	result1, err := driver.Run(ctx, config1)
	if err != nil {
		t.Fatalf("first command failed: %v", err)
	}

	if result1.Stdout != "first\n" {
		t.Errorf("expected 'first\\n', got %q", result1.Stdout)
	}

	containerID1 := result1.ContainerID

	// Second command - this should reuse the same container
	config2 := config.WithCommand("echo", "second")
	result2, err := driver.Run(ctx, config2)
	if err != nil {
		t.Fatalf("second command failed: %v", err)
	}

	if result2.Stdout != "second\n" {
		t.Errorf("expected 'second\\n', got %q", result2.Stdout)
	}

	containerID2 := result2.ContainerID

	// Verify same container was reused
	if containerID1 != containerID2 {
		t.Errorf("expected same container ID, got %s vs %s", containerID1, containerID2)
	}

	// Third command - verify container state persists
	config3 := config.WithCommand("sh", "-c", "echo test > /tmp/state && cat /tmp/state")
	result3, err := driver.Run(ctx, config3)
	if err != nil {
		t.Fatalf("third command failed: %v", err)
	}

	if result3.Stdout != "test\n" {
		t.Errorf("expected 'test\\n', got %q", result3.Stdout)
	}
}

func TestPodmanSecurityHardening(t *testing.T) {
	// Skip if podman is not available
	if err := IsAvailable(); err != nil {
		t.Skipf("podman not available: %v", err)
	}

	driver, err := New()
	if err != nil {
		t.Fatalf("failed to create podman driver: %v", err)
	}

	// Cleanup after test
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		driver.Cleanup(ctx)
	}()

	ctx := context.Background()

	tests := []struct {
		name    string
		command []string
		desc    string
	}{
		{
			name:    "read-only root filesystem",
			command: []string{"sh", "-c", "echo test > /test.txt 2>&1 || echo 'read-only confirmed'"},
			desc:    "should not be able to write to root filesystem",
		},
		{
			name:    "no network access",
			command: []string{"sh", "-c", "wget -q -O - httpbin.org/get 2>&1 || echo 'network blocked'"},
			desc:    "should not have network access",
		},
		{
			name:    "tmpfs /tmp writeable",
			command: []string{"sh", "-c", "echo test > /tmp/test.txt && cat /tmp/test.txt"},
			desc:    "should be able to write to /tmp (tmpfs)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := sandbox.NewExecutionConfig().
				WithImage("alpine:latest").
				WithCommand(tt.command...).
				WithCorrelationID("test-security-" + tt.name)

			result, err := driver.Run(ctx, config)
			if err != nil && result == nil {
				t.Fatalf("command failed: %v", err)
			}

			t.Logf("Command output: %q", result.Stdout)
			t.Logf("Command stderr: %q", result.Stderr)

			// These are behavioral tests - we verify the commands ran
			// and check their output for expected security behavior
			switch tt.name {
			case "read-only root filesystem":
				if !strings.Contains(result.Stdout, "read-only confirmed") && result.ExitCode == 0 {
					t.Error("expected read-only filesystem to prevent writes to root")
				}
			case "no network access":
				if !strings.Contains(result.Stdout, "network blocked") && result.ExitCode == 0 {
					t.Error("expected network access to be blocked")
				}
			case "tmpfs /tmp writeable":
				if result.Stdout != "test\n" {
					t.Errorf("expected tmpfs /tmp to be writeable, got %q", result.Stdout)
				}
			}
		})
	}
}
