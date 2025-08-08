package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/justinlyon12/ancli/internal/scheduler"
	"github.com/justinlyon12/ancli/internal/storage"
	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func main() {
	fmt.Println("AnCLI Storage + FSRS Integration Demo")
	fmt.Println("====================================")

	// Create temporary database
	tmpDir, err := os.MkdirTemp("", "ancli-demo")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "demo.db")
	db, err := storage.NewDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create a scheduler
	sched := scheduler.NewScheduler()

	fmt.Println("\n1. Creating a deck...")
	deck := &storage.Deck{
		Name:        "Linux Basics",
		Description: "Essential Linux command-line skills",
		Version:     "1.0.0",
		Author:      "AnCLI Team",
	}
	err = db.CreateDeck(deck)
	if err != nil {
		panic(err)
	}
	fmt.Printf("   Created deck: %s (ID: %d)\n", deck.Name, deck.ID)

	fmt.Println("\n2. Creating cards...")
	cards := []*storage.Card{
		{
			DeckID:      deck.ID,
			CardKey:     "ls-basic",
			Title:       "List Directory Contents",
			Description: "Learn to list files and directories",
			Command:     "ls -la /tmp",
			Tags:        `["filesystem", "basic"]`,
		},
		{
			DeckID:      deck.ID,
			CardKey:     "grep-search",
			Title:       "Search Text with Grep",
			Description: "Find text patterns in files",
			Command:     "echo 'Hello World' | grep 'World'",
			Tags:        `["text-processing", "intermediate"]`,
		},
	}

	for _, card := range cards {
		err = db.CreateCard(card)
		if err != nil {
			panic(err)
		}
		fmt.Printf("   Created card: %s (ID: %d)\n", card.Title, card.ID)
	}

	fmt.Println("\n3. Getting due cards...")
	dueCards, err := db.GetDueCards()
	if err != nil {
		panic(err)
	}
	fmt.Printf("   Found %d due cards\n", len(dueCards))

	fmt.Println("\n4. Simulating review session...")
	for i, card := range dueCards {
		fmt.Printf("\n   Card %d: %s\n", i+1, card.Title)
		fmt.Printf("   Command: %s\n", card.Command)

		// Convert to FSRS format
		fsrsCard := card.ToFSRSCard()
		fmt.Printf("   Current state: %v (reps: %d)\n", fsrsCard.State, fsrsCard.Reps)

		// Show scheduling options
		options := sched.GetSchedulingOptions(fsrsCard)
		fmt.Println("   Scheduling options:")
		for rating, info := range options {
			days := sched.DaysUntilDue(info.Card)
			fmt.Printf("     %s: Next review in %d days\n", rating.String(), days)
		}

		// Simulate user rating "Good"
		fmt.Println("   User rates: Good")
		result := sched.ReviewCard(fsrsCard, fsrs.Good)

		// Update card with new FSRS state
		beforeReps := card.FSRSReps
		card.UpdateFromFSRSCard(result.Card)

		// Save to database
		err = db.UpdateCardFSRS(card)
		if err != nil {
			panic(err)
		}

		// Create review record
		review := &storage.Review{
			CardID:               card.ID,
			Rating:               int(fsrs.Good),
			ExecutionSuccess:     true,
			ExitCode:             func() *int { code := 0; return &code }(),
			Stdout:               "Command executed successfully",
			Stderr:               "",
			Attempts:             1,
			HelpAccessed:         false,
			FSRSDueBefore:        fsrsCard.Due,
			FSRSDueAfter:         result.Card.Due,
			FSRSStabilityBefore:  fsrsCard.Stability,
			FSRSStabilityAfter:   result.Card.Stability,
			FSRSDifficultyBefore: fsrsCard.Difficulty,
			FSRSDifficultyAfter:  result.Card.Difficulty,
		}
		err = db.CreateReview(review)
		if err != nil {
			panic(err)
		}

		fmt.Printf("   Updated: reps %d -> %d, next due: %s\n",
			beforeReps, card.FSRSReps, card.FSRSDue.Format("2006-01-02 15:04"))
	}

	fmt.Println("\n5. Checking updated due cards...")
	dueCards, err = db.GetDueCards()
	if err != nil {
		panic(err)
	}
	fmt.Printf("   Now %d cards are due (should be 0 after reviews)\n", len(dueCards))

	fmt.Println("\n6. Deck statistics...")
	allDecks, err := db.ListDecks()
	if err != nil {
		panic(err)
	}
	fmt.Printf("   Total decks: %d\n", len(allDecks))

	fmt.Println("\nDemo completed successfully!")
	fmt.Printf("Database created at: %s\n", dbPath)
}
