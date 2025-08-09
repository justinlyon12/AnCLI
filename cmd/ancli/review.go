package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/justinlyon12/ancli/internal/domain"
	"github.com/justinlyon12/ancli/internal/review"
	"github.com/justinlyon12/ancli/internal/sandbox"
	"github.com/justinlyon12/ancli/internal/sandbox/podman"
	"github.com/justinlyon12/ancli/internal/scheduler"
	"github.com/justinlyon12/ancli/internal/storage"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Start a flashcard review session",
	Long: `Start an interactive flashcard review session. Cards will be presented one at a time,
executed in a secure container, and you'll rate your performance for spaced repetition scheduling.

The session continues until all due cards are reviewed or you quit with 'q'.`,
	RunE: runReview,
}

var (
	reviewDeckID    int
	reviewMaxCards  int
	reviewNewOnly   bool
	reviewDueOnly   bool
	reviewShuffle   bool
	reviewNoNetwork bool
)

func init() {
	rootCmd.AddCommand(reviewCmd)

	// Review-specific flags
	reviewCmd.Flags().IntVar(&reviewDeckID, "deck-id", 0, "review cards from specific deck ID (0 = all decks)")
	reviewCmd.Flags().IntVar(&reviewMaxCards, "max-cards", 20, "maximum cards per session (0 = unlimited)")
	reviewCmd.Flags().BoolVar(&reviewNewOnly, "new-only", false, "only review new cards")
	reviewCmd.Flags().BoolVar(&reviewDueOnly, "due-only", false, "only review cards that are due")
	reviewCmd.Flags().BoolVar(&reviewShuffle, "shuffle", true, "randomize card order")
	reviewCmd.Flags().BoolVar(&reviewNoNetwork, "no-network", true, "disable network access (safer)")
}

func runReview(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize storage
	dbPath, err := cfg.GetDatabasePath()
	if err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	db, err := storage.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize scheduler
	sched := scheduler.NewScheduler()

	// Initialize sandbox
	var sb sandbox.Sandbox
	switch cfg.Sandbox.Driver {
	case "podman":
		var err error
		sb, err = podman.New()
		if err != nil {
			return fmt.Errorf("failed to create podman driver: %w", err)
		}
	default:
		return fmt.Errorf("unsupported sandbox driver: %s", cfg.Sandbox.Driver)
	}

	// Initialize review service
	reviewService := review.NewService(db, sched, sb)

	// Set up session options
	var deckID *int
	if reviewDeckID > 0 {
		deckID = &reviewDeckID
	}

	opts := review.SessionOptions{
		DeckID:          deckID,
		MaxCards:        reviewMaxCards,
		NewCardsOnly:    reviewNewOnly,
		ReviewCardsOnly: reviewDueOnly,
		ShuffleCards:    reviewShuffle,
		NetworkEnabled:  !reviewNoNetwork, // Invert the flag
	}

	// Start session
	fmt.Println("🚀 Starting review session...")
	session, err := reviewService.StartSession(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to start review session: %w", err)
	}

	fmt.Printf("📚 Session started with %d cards\n", session.CardsRemaining)
	if !opts.NetworkEnabled {
		fmt.Println("🔒 Network access disabled for security")
	} else {
		fmt.Println("🌐 Network access enabled")
	}

	// Review loop
	scanner := bufio.NewScanner(os.Stdin)
	cardsReviewed := 0

	for session.CardsRemaining > 0 {
		// Get next card
		card, err := reviewService.GetNextCard(ctx, session.ID)
		if err != nil {
			if strings.Contains(err.Error(), "no more cards") {
				fmt.Println("✅ No more cards to review!")
				break
			}
			return fmt.Errorf("failed to get next card: %w", err)
		}

		// Show card info
		fmt.Print("\n" + strings.Repeat("=", 60) + "\n")
		fmt.Printf("📋 Card: %s\n", card.Title)
		if card.Description != "" {
			fmt.Printf("📖 Description: %s\n", card.Description)
		}
		fmt.Printf("🐳 Image: %s | ⏱️  Timeout: %v\n", card.Image, card.Timeout)
		if card.NetworkEnabled {
			fmt.Printf("🌐 Network: ENABLED\n")
		}
		fmt.Printf("📁 Working Dir: %s\n", card.WorkingDir)
		fmt.Printf("🔧 Command: %s\n", card.Command)
		fmt.Print(strings.Repeat("=", 60) + "\n")

		// Wait for user to be ready
		fmt.Print("Press Enter when ready to execute the command (or 'q' to quit): ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "q" || input == "quit" {
			fmt.Println("👋 Quitting review session...")
			break
		}

		// Record thinking start time
		thinkingStart := time.Now()

		// Execute command
		fmt.Println("\n🏃 Executing command...")
		sandboxConfig := sandbox.ExecutionConfig{
			Command:        strings.Fields(card.Command), // Convert string to []string
			WorkingDir:     card.WorkingDir,
			Image:          card.Image,
			Timeout:        card.Timeout,
			NetworkEnabled: card.NetworkEnabled,
			Capabilities:   card.Capabilities,
			Environment:    card.EnvironmentVars,
		}

		result, err := sb.Run(ctx, sandboxConfig)
		if err != nil {
			fmt.Printf("❌ Execution failed: %v\n", err)
			// Still allow rating for learning purposes
		} else {
			fmt.Printf("✅ Command completed (exit code: %d)\n", result.ExitCode)
		}

		// Show output
		if result != nil {
			if result.Stdout != "" {
				fmt.Println("\n📤 STDOUT:")
				fmt.Println(result.Stdout)
			}
			if result.Stderr != "" {
				fmt.Println("\n📤 STDERR:")
				fmt.Println(result.Stderr)
			}
		}

		// Calculate thinking time
		thinkingTime := time.Since(thinkingStart)

		// Get user rating
		var rating domain.Rating
		for {
			fmt.Print("\n⭐ Rate your performance (1=Again, 2=Hard, 3=Good, 4=Easy): ")
			if !scanner.Scan() {
				return fmt.Errorf("failed to read rating")
			}

			ratingInput := strings.TrimSpace(scanner.Text())
			if ratingInput == "q" || ratingInput == "quit" {
				fmt.Println("👋 Quitting review session...")
				goto cleanup
			}

			parsedRating, err := domain.ParseRating(ratingInput)
			if err != nil {
				fmt.Printf("❌ Invalid rating: %v\n", err)
				continue
			}
			rating = parsedRating
			break
		}

		// Create execution result
		var executionResult *domain.ExecutionResult
		if result != nil {
			executionResult = &domain.ExecutionResult{
				Success:        result.Success,
				ExitCode:       result.ExitCode,
				Stdout:         result.Stdout,
				Stderr:         result.Stderr,
				Duration:       result.Duration,
				ThinkingTime:   thinkingTime,
				ContainerID:    result.ContainerID,
				ImageUsed:      result.ImageUsed,
				NetworkEnabled: card.NetworkEnabled,
			}
		}

		// Submit review
		err = reviewService.SubmitReview(ctx, session.ID, card.ID, rating, executionResult)
		if err != nil {
			return fmt.Errorf("failed to submit review: %w", err)
		}

		fmt.Printf("✅ Review submitted! Rating: %s\n", rating.String())
		cardsReviewed++
		session.CardsRemaining--
	}

cleanup:
	// End session and show stats
	fmt.Println("\n📊 Finalizing session...")
	stats, err := reviewService.EndSession(ctx, session.ID)
	if err != nil {
		fmt.Printf("Warning: Failed to get session stats: %v\n", err)
	} else {
		fmt.Printf("🎯 Session completed in %v\n", stats.Duration.Round(time.Second))
		fmt.Printf("📈 Cards reviewed: %d\n", stats.CardsReviewed)
	}

	// Cleanup sandbox
	if err := sb.Cleanup(ctx); err != nil {
		fmt.Printf("Warning: Failed to cleanup sandbox: %v\n", err)
	}

	fmt.Println("👋 Thanks for studying!")
	return nil
}
