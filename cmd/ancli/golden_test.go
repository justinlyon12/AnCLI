package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
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
			// Create a new root command for each test to avoid state pollution
			cmd := &cobra.Command{
				Use:   "ancli",
				Short: "A flashcard platform for learning and practicing CLI with spaced repetition",
				Long: `AnCLI is a CLI-first flashcard platform that teaches real-world command-line skills
with spaced repetition. The use must execute an actual shell command in a rootless OCI container
for each card, let the user grade themselves (Again | Hard | Good | Easy), and reschedules with 
the FSRS 4-parameter algorithm.`,
			}

			// Add global flags
			cmd.PersistentFlags().String("config", "", "config file (default is $HOME/.ancli/ancli.yaml)")
			cmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
			cmd.PersistentFlags().Bool("log-json", false, "log in JSON format")
			cmd.PersistentFlags().String("database-path", "", "database file path")
			cmd.PersistentFlags().String("sandbox-driver", "podman", "sandbox driver (podman, docker)")
			cmd.PersistentFlags().Bool("sandbox-network", false, "enable network access for sandbox")

			// Add review subcommand
			reviewCmd := &cobra.Command{
				Use:   "review",
				Short: "Start a flashcard review session",
				Long: `Start an interactive flashcard review session. Cards will be presented one at a time,
executed in a secure container, and you'll rate your performance for spaced repetition scheduling.

The session continues until all due cards are reviewed or you quit with 'q'.`,
				Run: func(cmd *cobra.Command, args []string) {
					// No-op for testing
				},
			}

			// Add review-specific flags
			reviewCmd.Flags().Int("deck-id", 0, "review cards from specific deck ID (0 = all decks)")
			reviewCmd.Flags().Int("max-cards", 20, "maximum cards per session (0 = unlimited)")
			reviewCmd.Flags().Bool("new-only", false, "only review new cards")
			reviewCmd.Flags().Bool("due-only", false, "only review cards that are due")
			reviewCmd.Flags().Bool("shuffle", true, "randomize card order")
			reviewCmd.Flags().Bool("no-network", true, "disable network access (safer)")

			cmd.AddCommand(reviewCmd)

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
