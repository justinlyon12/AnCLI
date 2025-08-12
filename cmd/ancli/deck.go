package main

import (
	"fmt"

	"github.com/justinlyon12/ancli/internal/deck"
	"github.com/spf13/cobra"
)

// NewDeckCmd creates a new deck management command
func NewDeckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deck",
		Short: "Manage AnCLI decks",
		Long: `Manage AnCLI decks including validation, testing, packaging, and installation.

Decks are collections of flashcards that teach command-line skills through
hands-on practice with spaced repetition. Each deck contains:
- deck.yaml: Metadata and configuration  
- cards.csv: Individual flashcard definitions
- assets/: Optional supporting files`,
	}

	// Add subcommands
	cmd.AddCommand(NewLintCmd())

	return cmd
}

// NewLintCmd creates a new lint command for validating decks
func NewLintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint [deck-path]",
		Short: "Validate deck structure and content",
		Long: `Validate a deck directory for structural integrity, content quality,
security issues, and usability concerns.

Examples:
  ancli deck lint .                    # Validate current directory
  ancli deck lint examples/my-deck     # Validate specific deck
  ancli deck lint . --verbose         # Show detailed validation info
  ancli deck lint . --json           # JSON output for automation`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			deckPath := "."
			if len(args) > 0 {
				deckPath = args[0]
			}

			verbose, _ := cmd.Flags().GetBool("verbose")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			result, err := deck.ValidateDeck(deckPath)
			if err != nil {
				fmt.Printf("Error validating deck: %v\n", err)
				return
			}

			if jsonOutput {
				// TODO: Output JSON format
				fmt.Printf("JSON output not yet implemented\n")
				return
			}

			deck.PrintValidationResult(result, verbose)

			if !result.Valid {
				// Exit with error code for CI/automation
				fmt.Printf("\nDeck validation failed. Please fix errors before using this deck.\n")
			}
		},
	}

	// Add flags
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed validation information")
	cmd.Flags().Bool("json", false, "Output validation results in JSON format")

	return cmd
}
