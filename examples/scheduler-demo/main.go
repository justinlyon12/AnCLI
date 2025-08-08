package main

import (
	"fmt"

	"github.com/justinlyon12/ancli/internal/scheduler"
	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func main() {
	fmt.Println("AnCLI FSRS Scheduler Demo")
	fmt.Println("========================")

	// Create a new scheduler
	sched := scheduler.NewScheduler()

	// Create a new card
	card := sched.NewCard()
	fmt.Printf("New card created with state: %v\n", card.State)
	fmt.Printf("Card is due: %v\n", sched.IsDue(card))

	// Show scheduling options for the new card
	fmt.Println("\nScheduling options for new card:")
	options := sched.GetSchedulingOptions(card)
	for rating, info := range options {
		days := sched.DaysUntilDue(info.Card)
		fmt.Printf("  %s: Next review in %d days\n", rating.String(), days)
	}

	// Review the card with "Good"
	fmt.Println("\nReviewing card with 'Good' rating...")
	result := sched.ReviewCard(card, fsrs.Good)
	updatedCard := result.Card

	fmt.Printf("Card state after review: %v\n", updatedCard.State)
	fmt.Printf("Card reps: %d\n", updatedCard.Reps)
	fmt.Printf("Card difficulty: %.2f\n", updatedCard.Difficulty)
	fmt.Printf("Card stability: %.2f\n", updatedCard.Stability)
	fmt.Printf("Next review due: %s\n", updatedCard.Due.Format("2006-01-02 15:04:05"))
	fmt.Printf("Days until due: %d\n", sched.DaysUntilDue(updatedCard))
	fmt.Printf("Current retrievability: %.2f\n", sched.GetRetrievability(updatedCard))

	// Show what would happen with different ratings
	fmt.Println("\nWhat would happen with different ratings:")
	newOptions := sched.GetSchedulingOptions(updatedCard)
	for rating, info := range newOptions {
		days := sched.DaysUntilDue(info.Card)
		fmt.Printf("  %s: Next review in %d days (difficulty: %.2f)\n",
			rating.String(), days, info.Card.Difficulty)
	}

	// Simulate reviewing again with "Hard"
	fmt.Println("\nSimulating another review with 'Hard' rating...")
	hardResult := sched.ReviewCard(updatedCard, fsrs.Hard)
	hardCard := hardResult.Card

	fmt.Printf("After 'Hard' review:\n")
	fmt.Printf("  Reps: %d\n", hardCard.Reps)
	fmt.Printf("  Lapses: %d\n", hardCard.Lapses)
	fmt.Printf("  Difficulty: %.2f\n", hardCard.Difficulty)
	fmt.Printf("  Days until due: %d\n", sched.DaysUntilDue(hardCard))
	fmt.Printf("  Retrievability: %.2f\n", sched.GetRetrievability(hardCard))
}
