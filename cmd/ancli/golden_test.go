package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/justinlyon12/ancli/internal/config"
)

// TestGoldenFiles verifies CLI output matches golden files
func TestGoldenFiles(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		golden string
	}{
		{
			name:   "main help",
			args:   []string{"--help"},
			golden: "help.golden",
		},
		{
			name:   "review help",
			args:   []string{"review", "--help"},
			golden: "review-help.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test configuration
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

			// Create test config loader for help output
			// We don't need real storage/sandbox connections for help text
			testLoader := &TestConfigLoader{
				Config: testConfig,
			}

			// Create command using the same factory function as production
			cmd := NewRootCmd(testLoader)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			// Execute command
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			// Read golden file
			goldenPath := filepath.Join("../../testdata", tt.golden)
			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
			}

			// Compare output
			actual := buf.String()
			if actual != string(expected) {
				t.Errorf("Output doesn't match golden file %s\nGot:\n%s\nExpected:\n%s", tt.golden, actual, string(expected))

				// Offer to update golden file (useful during development)
				t.Logf("To update golden file, run: echo %q > %s", actual, goldenPath)
			}
		})
	}
}

// TestMainRun tests the main run function with different scenarios
func TestMainRun(t *testing.T) {
	tests := []struct {
		name             string
		configLoader     ConfigLoader
		expectedExitCode int
		expectError      bool
	}{
		{
			name: "config load failure",
			configLoader: &TestConfigLoader{
				Config: nil, // Will cause error
			},
			expectedExitCode: 1,
			expectError:      true,
		},
		{
			name: "valid config with help",
			configLoader: &TestConfigLoader{
				Config: &config.Config{
					Database: config.DatabaseConfig{
						Path: "/tmp/test.db",
					},
					Sandbox: config.SandboxConfig{
						Driver: "podman",
					},
				},
			},
			expectedExitCode: 0,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr to verify error messages
			originalStderr := os.Stderr
			_, w, _ := os.Pipe()
			os.Stderr = w

			// Note: We can't easily test the full run() function here because
			// it would try to create real storage connections. In a full implementation,
			// you'd want to add dependency injection for storage/sandbox factories too.

			// For now, just test the config loading part
			_, err := tt.configLoader.Load()

			w.Close()
			os.Stderr = originalStderr

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
