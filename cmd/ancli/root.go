package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRootCmd creates a new root command with lazy initialization
// This factory function eliminates global state and enables testing
func NewRootCmd(loader ConfigLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ancli",
		Short: "A flashcard platform for learning and practicing CLI with spaced repetition",
		Long: `AnCLI is a CLI-first flashcard platform that teaches real-world command-line skills
with spaced repetition. The use must execute an actual shell command in a rootless OCI container
for each card, let the user grade themselves (Again | Hard | Good | Easy), and reschedules with 
the FSRS 4-parameter algorithm.`,
		SilenceUsage: true,
	}

	// Add global flags
	addGlobalFlags(cmd)

	// Add subcommands
	cmd.AddCommand(NewReviewCmd(loader))
	cmd.AddCommand(NewDeckCmd())

	return cmd
}

// addGlobalFlags adds global flags to the command
// Flags are now defined without side effects (no init(), no viper binding)
func addGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("config", "", "config file (default is $HOME/.ancli/ancli.yaml)")
	cmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	cmd.PersistentFlags().Bool("log-json", false, "log in JSON format")
	cmd.PersistentFlags().String("database-path", "", "database file path")
	cmd.PersistentFlags().String("sandbox-driver", "podman", "sandbox driver (podman, docker)")
	cmd.PersistentFlags().Bool("sandbox-network", false, "enable network access for sandbox")
}

// initializeApp loads configuration and creates the application
// This is called only when actually needed by subcommands
func initializeApp(loader ConfigLoader) (*App, error) {
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	app, err := NewApp(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize application: %w", err)
	}

	return app, nil
}
