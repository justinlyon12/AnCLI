package podman

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/justinlyon12/ancli/internal/sandbox"
)

func TestNew(t *testing.T) {
	// Check if podman is available in the test environment
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("podman not available, skipping test")
	}

	driver, err := New()
	if err != nil {
		t.Fatalf("failed to create podman driver: %v", err)
	}

	if driver.Name() != "podman" {
		t.Errorf("expected driver name 'podman', got %s", driver.Name())
	}

	if driver.lifecycle != sandbox.SessionReuse {
		t.Errorf("expected default lifecycle SessionReuse, got %s", driver.lifecycle)
	}

	if driver.containerID != "" {
		t.Errorf("expected empty container ID on new driver, got %s", driver.containerID)
	}
}

func TestNewWithoutPodman(t *testing.T) {
	// This test verifies error handling when podman is not found
	// We can't easily mock exec.LookPath in Go, so this is a structural test

	// If podman is available, skip this test
	if _, err := exec.LookPath("podman"); err == nil {
		t.Skip("podman is available, cannot test 'not found' condition")
	}

	// If podman is not available, verify New() returns appropriate error
	_, err := New()
	if err == nil {
		t.Error("expected error when podman not available")
	}

	if !strings.Contains(err.Error(), "podman not found") {
		t.Errorf("expected 'podman not found' error, got: %v", err)
	}
}

func TestDriverInterface(t *testing.T) {
	// Verify Driver implements the Sandbox interface
	var _ sandbox.Sandbox = (*Driver)(nil)
}

func TestIsAvailable(t *testing.T) {
	err := IsAvailable()
	if err != nil {
		t.Skip("podman not available, skipping test")
	}
}

func TestRunWithInvalidConfig(t *testing.T) {
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("podman not available, skipping test")
	}

	driver, err := New()
	if err != nil {
		t.Skipf("failed to create driver: %v", err)
	}

	ctx := context.Background()
	invalidConfigs := []sandbox.ExecutionConfig{
		// Missing image
		sandbox.NewExecutionConfig().WithCommand("echo", "test"),
		// Missing command
		sandbox.NewExecutionConfig().WithImage("alpine:latest"),
		// Invalid timeout
		{
			Image:   "alpine:latest",
			Command: []string{"echo", "test"},
			Timeout: -1 * time.Second,
		},
	}

	for i, config := range invalidConfigs {
		t.Run(fmt.Sprintf("invalid_config_%d", i), func(t *testing.T) {
			_, err := driver.Run(ctx, config)
			if err == nil {
				t.Error("expected validation error for invalid config")
			}
			if !strings.Contains(err.Error(), "invalid config") {
				t.Errorf("expected 'invalid config' error, got: %v", err)
			}
		})
	}
}

func TestRunPerCardNotImplemented(t *testing.T) {
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("podman not available, skipping test")
	}

	driver, err := New()
	if err != nil {
		t.Skipf("failed to create driver: %v", err)
	}

	// Override lifecycle to test per-card mode
	driver.lifecycle = sandbox.PerCard

	ctx := context.Background()
	config := sandbox.NewExecutionConfig().
		WithImage("alpine:latest").
		WithCommand("echo", "test").
		WithCorrelationID("test-per-card")

	_, err = driver.Run(ctx, config)
	if err == nil {
		t.Error("expected error for unimplemented per-card mode")
	}

	if !strings.Contains(err.Error(), "per-card lifecycle not implemented") {
		t.Errorf("expected 'per-card lifecycle not implemented' error, got: %v", err)
	}
}

func TestCleanupWithoutContainer(t *testing.T) {
	driver := &Driver{podmanPath: "/usr/bin/podman"}

	ctx := context.Background()
	err := driver.Cleanup(ctx)

	// Cleanup should not fail when no container exists
	if err != nil {
		t.Errorf("cleanup should not fail with no container, got: %v", err)
	}
}

func TestDriverRegistration(t *testing.T) {
	// Test that the driver registers itself via init()
	if !sandbox.IsRegistered("podman") {
		t.Error("podman driver should be registered via init()")
	}

	// Test that we can create a driver via the registry
	driver, err := sandbox.Get("podman")
	if err != nil {
		// Skip if podman is not available
		if strings.Contains(err.Error(), "not found") {
			t.Skip("podman not available, skipping registry test")
		}
		t.Fatalf("failed to get podman driver from registry: %v", err)
	}

	if driver.Name() != "podman" {
		t.Errorf("expected driver name 'podman', got %s", driver.Name())
	}
}

func TestSecurityDefaults(t *testing.T) {
	tests := []struct {
		name   string
		config sandbox.ExecutionConfig
		desc   string
	}{
		{
			name:   "network disabled by default",
			config: sandbox.NewExecutionConfig(),
			desc:   "network should be disabled by default",
		},
		{
			name:   "capabilities dropped",
			config: sandbox.NewExecutionConfig(),
			desc:   "capabilities should be empty (dropped) by default",
		},
		{
			name:   "read-only filesystem",
			config: sandbox.NewExecutionConfig(),
			desc:   "root filesystem should be read-only by default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Structural test - verify our configuration has the right security defaults
			if tt.name == "network disabled by default" && tt.config.NetworkEnabled {
				t.Error("network should be disabled by default")
			}

			if tt.name == "read-only filesystem" && !tt.config.ReadOnlyRootFS {
				t.Error("root filesystem should be read-only by default")
			}

			if tt.name == "capabilities dropped" && len(tt.config.Capabilities) != 0 {
				t.Error("capabilities should be empty (dropped) by default")
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("podman not available, skipping concurrency test")
	}

	driver, err := New()
	if err != nil {
		t.Skipf("failed to create driver: %v", err)
	}

	// Test that concurrent access to driver state doesn't cause issues
	done := make(chan struct{})
	errors := make(chan error, 2)

	// Start cleanup in one goroutine
	go func() {
		defer close(done)
		err := driver.Cleanup(context.Background())
		errors <- err
	}()

	// Try to access containerID in another goroutine
	go func() {
		<-done
		driver.mu.Lock()
		_ = driver.containerID // Access protected field
		driver.mu.Unlock()
		errors <- nil
	}()

	// Wait for both operations
	for i := 0; i < 2; i++ {
		select {
		case err := <-errors:
			if err != nil {
				t.Errorf("concurrent operation failed: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent operations")
		}
	}
}
