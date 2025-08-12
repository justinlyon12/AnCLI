package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewDeckCmd(t *testing.T) {
	cmd := NewDeckCmd()

	if cmd.Use != "deck" {
		t.Errorf("expected Use='deck', got %s", cmd.Use)
	}

	if cmd.Short != "Manage AnCLI decks" {
		t.Errorf("expected correct short description, got %s", cmd.Short)
	}

	// Check that subcommands are added
	subCmds := cmd.Commands()
	if len(subCmds) != 1 {
		t.Errorf("expected 1 subcommand, got %d", len(subCmds))
	}

	if subCmds[0].Use != "lint [deck-path]" {
		t.Errorf("expected lint subcommand, got %s", subCmds[0].Use)
	}
}

func TestNewLintCmd(t *testing.T) {
	cmd := NewLintCmd()

	if cmd.Use != "lint [deck-path]" {
		t.Errorf("expected Use='lint [deck-path]', got %s", cmd.Use)
	}

	if cmd.Short != "Validate deck structure and content" {
		t.Errorf("expected correct short description, got %s", cmd.Short)
	}

	// Check flags are present
	verboseFlag := cmd.Flags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("expected verbose flag to exist")
	} else if verboseFlag.Shorthand != "v" {
		t.Errorf("expected verbose flag shorthand 'v', got %s", verboseFlag.Shorthand)
	}

	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("expected json flag to exist")
	}
}

func TestLintCommand_ExecutesSuccessfully(t *testing.T) {
	// Test that the lint command executes without crashing
	cmd := NewLintCmd()

	// Test with non-existent path - should not crash
	cmd.SetArgs([]string{"/nonexistent/path"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	// If we get here, the command executed successfully
}

func TestLintCommand_MaxArgs(t *testing.T) {
	// Test that command accepts maximum of 1 argument
	cmd := NewLintCmd()

	// Check Args validation
	if cmd.Args == nil {
		t.Error("expected Args validation to be set")
	}

	// Test with too many arguments
	var err error
	err = cmd.Args(cmd, []string{"path1", "path2"})
	if err == nil {
		t.Error("expected error with too many arguments")
	}

	// Test with valid number of arguments
	err = cmd.Args(cmd, []string{"path1"})
	if err != nil {
		t.Errorf("expected no error with 1 argument, got: %v", err)
	}

	err = cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("expected no error with 0 arguments, got: %v", err)
	}
}

func TestDeckCommandIntegration(t *testing.T) {
	// Test that deck command integrates properly with root command
	rootCmd := &cobra.Command{Use: "ancli"}
	deckCmd := NewDeckCmd()
	rootCmd.AddCommand(deckCmd)

	// Check that deck command is added
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "deck" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected deck command to be added to root command")
	}

	// Test that lint subcommand is accessible
	lintCmd, _, err := rootCmd.Find([]string{"deck", "lint"})
	if err != nil {
		t.Fatalf("failed to find deck lint command: %v", err)
	}

	if lintCmd.Use != "lint [deck-path]" {
		t.Errorf("expected to find lint command, got %s", lintCmd.Use)
	}
}

func TestCommandHelp(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *cobra.Command
		contain []string
	}{
		{
			name: "deck command help",
			cmd:  NewDeckCmd(),
			contain: []string{
				"Manage AnCLI decks",
				"deck.yaml: Metadata and configuration",
				"cards.csv: Individual flashcard definitions",
			},
		},
		{
			name: "lint command help",
			cmd:  NewLintCmd(),
			contain: []string{
				"Validate a deck directory for structural integrity",
				"ancli deck lint .",
				"ancli deck lint . --verbose",
				"--verbose",
				"--json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.cmd.Long
			if help == "" {
				help = tt.cmd.Short
			}

			for _, expected := range tt.contain {
				if !strings.Contains(help, expected) {
					t.Errorf("expected help to contain %q, but it doesn't. Help: %s", expected, help)
				}
			}
		})
	}
}
