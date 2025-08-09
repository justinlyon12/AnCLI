package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/justinlyon12/ancli/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ancli",
	Short: "A flashcard platform for learning and practicing CLI with spaced repetition",
	Long: `AnCLI is a CLI-first flashcard platform that teaches real-world command-line skills
with spaced repetition. The use must execute an actual shell command in a rootless OCI container
for each card, let the user grade themselves (Again | Hard | Good | Easy), and reschedules with 
the FSRS 4-parameter algorithm.`,
	PersistentPreRunE: initConfig,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ancli/ancli.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().Bool("log-json", false, "log in JSON format")
	rootCmd.PersistentFlags().String("database-path", "", "database file path")
	rootCmd.PersistentFlags().String("sandbox-driver", "podman", "sandbox driver (podman, docker)")
	rootCmd.PersistentFlags().Bool("sandbox-network", false, "enable network access for sandbox")

	// Bind flags to viper
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("log_json", rootCmd.PersistentFlags().Lookup("log-json"))
	_ = viper.BindPFlag("database.path", rootCmd.PersistentFlags().Lookup("database-path"))
	_ = viper.BindPFlag("sandbox.driver", rootCmd.PersistentFlags().Lookup("sandbox-driver"))
	_ = viper.BindPFlag("sandbox.network_enabled", rootCmd.PersistentFlags().Lookup("sandbox-network"))
}

// initConfig reads in config file and ENV variables.
func initConfig(cmd *cobra.Command, args []string) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	var err error
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	return nil
}
