package scheduler

import (
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestNewScheduler(t *testing.T) {
	scheduler := NewScheduler()
	if scheduler == nil {
		t.Fatal("NewScheduler() returned nil")
	}
	if scheduler.fsrs == nil {
		t.Fatal("NewScheduler() did not initialize FSRS instance")
	}
}

func TestNewSchedulerWithParams(t *testing.T) {
	params := fsrs.DefaultParam()
	params.RequestRetention = 0.95 // Custom retention rate

	scheduler := NewSchedulerWithParams(params)
	if scheduler == nil {
		t.Fatal("NewSchedulerWithParams() returned nil")
	}
	if scheduler.fsrs == nil {
		t.Fatal("NewSchedulerWithParams() did not initialize FSRS instance")
	}
}

func TestNewCard(t *testing.T) {
	scheduler := NewScheduler()
	card := scheduler.NewCard()

	// Verify initial card state
	if card.State != fsrs.New {
		t.Errorf("Expected new card state to be New, got %v", card.State)
	}
	if card.Reps != 0 {
		t.Errorf("Expected new card reps to be 0, got %d", card.Reps)
	}
	if card.Lapses != 0 {
		t.Errorf("Expected new card lapses to be 0, got %d", card.Lapses)
	}
	if card.Difficulty != 0 {
		t.Errorf("Expected new card difficulty to be 0, got %f", card.Difficulty)
	}
	if card.Stability != 0 {
		t.Errorf("Expected new card stability to be 0, got %f", card.Stability)
	}
}

func TestReviewCard(t *testing.T) {
	scheduler := NewScheduler()
	card := scheduler.NewCard()

	// Test all rating types
	ratings := []fsrs.Rating{fsrs.Again, fsrs.Hard, fsrs.Good, fsrs.Easy}

	for _, rating := range ratings {
		t.Run(rating.String(), func(t *testing.T) {
			result := scheduler.ReviewCard(card, rating)

			// Verify the result contains updated card and review log
			if result.Card.Reps != 1 {
				t.Errorf("Expected card reps to be 1 after review, got %d", result.Card.Reps)
			}
			if result.Card.State == fsrs.New {
				t.Error("Expected card state to change from New after review")
			}
			if result.ReviewLog.Rating != rating {
				t.Errorf("Expected review log rating to be %v, got %v", rating, result.ReviewLog.Rating)
			}
			if result.ReviewLog.Review.IsZero() {
				t.Error("Expected review log to have a review time")
			}
		})
	}
}

func TestGetSchedulingOptions(t *testing.T) {
	scheduler := NewScheduler()
	card := scheduler.NewCard()

	options := scheduler.GetSchedulingOptions(card)

	// Verify all rating options are present
	expectedRatings := []fsrs.Rating{fsrs.Again, fsrs.Hard, fsrs.Good, fsrs.Easy}
	for _, rating := range expectedRatings {
		if _, exists := options[rating]; !exists {
			t.Errorf("Expected scheduling options to contain rating %v", rating)
		}
	}

	// Verify each option has valid scheduling info
	for rating, info := range options {
		if info.Card.Reps != 1 {
			t.Errorf("Expected card reps to be 1 for rating %v, got %d", rating, info.Card.Reps)
		}
		if info.ReviewLog.Rating != rating {
			t.Errorf("Expected review log rating to match %v, got %v", rating, info.ReviewLog.Rating)
		}
	}
}

func TestIsDue(t *testing.T) {
	scheduler := NewScheduler()

	// Test with a card that's due now
	cardDueNow := scheduler.NewCard()
	cardDueNow.Due = time.Now().Add(-1 * time.Hour) // 1 hour ago

	if !scheduler.IsDue(cardDueNow) {
		t.Error("Expected card due 1 hour ago to be due")
	}

	// Test with a card that's due exactly now
	cardDueExactly := scheduler.NewCard()
	cardDueExactly.Due = time.Now()

	if !scheduler.IsDue(cardDueExactly) {
		t.Error("Expected card due exactly now to be due")
	}

	// Test with a card that's not due yet
	cardNotDue := scheduler.NewCard()
	cardNotDue.Due = time.Now().Add(1 * time.Hour) // 1 hour from now

	if scheduler.IsDue(cardNotDue) {
		t.Error("Expected card due in 1 hour to not be due")
	}
}

func TestGetRetrievability(t *testing.T) {
	scheduler := NewScheduler()
	card := scheduler.NewCard()

	// For a new card, retrievability should be a valid float between 0 and 1
	retrievability := scheduler.GetRetrievability(card)

	if retrievability < 0 || retrievability > 1 {
		t.Errorf("Expected retrievability to be between 0 and 1, got %f", retrievability)
	}

	// Test with a reviewed card
	reviewResult := scheduler.ReviewCard(card, fsrs.Good)
	reviewedCard := reviewResult.Card

	retrievability = scheduler.GetRetrievability(reviewedCard)
	if retrievability < 0 || retrievability > 1 {
		t.Errorf("Expected retrievability to be between 0 and 1 for reviewed card, got %f", retrievability)
	}
}

func TestDaysUntilDue(t *testing.T) {
	scheduler := NewScheduler()

	// Test with a card that's already due
	cardDue := scheduler.NewCard()
	cardDue.Due = time.Now().Add(-1 * time.Hour)

	days := scheduler.DaysUntilDue(cardDue)
	if days != 0 {
		t.Errorf("Expected 0 days until due for overdue card, got %d", days)
	}

	// Test with a card due in the future
	cardFuture := scheduler.NewCard()
	cardFuture.Due = time.Now().Add(49 * time.Hour) // Just over 2 days from now

	days = scheduler.DaysUntilDue(cardFuture)
	if days != 2 {
		t.Errorf("Expected 2 days until due, got %d", days)
	}

	// Test with a card due in less than a day
	cardSoon := scheduler.NewCard()
	cardSoon.Due = time.Now().Add(12 * time.Hour) // 12 hours from now

	days = scheduler.DaysUntilDue(cardSoon)
	if days != 0 {
		t.Errorf("Expected 0 days until due for card due in 12 hours, got %d", days)
	}
}

func TestSchedulerIntegration(t *testing.T) {
	scheduler := NewScheduler()

	// Create a new card
	card := scheduler.NewCard()

	// Verify it's due immediately
	if !scheduler.IsDue(card) {
		t.Error("New card should be due immediately")
	}

	// Review it with "Good"
	result := scheduler.ReviewCard(card, fsrs.Good)
	updatedCard := result.Card

	// Verify the card is no longer due immediately
	if scheduler.IsDue(updatedCard) {
		t.Error("Card should not be due immediately after first review")
	}

	// Verify days until due is non-negative (could be 0 for same-day scheduling)
	days := scheduler.DaysUntilDue(updatedCard)
	if days < 0 {
		t.Errorf("Expected non-negative days until due after review, got %d", days)
	}

	// Verify retrievability is high for a recently reviewed card
	retrievability := scheduler.GetRetrievability(updatedCard)
	if retrievability < 0.8 {
		t.Errorf("Expected high retrievability for recently reviewed card, got %f", retrievability)
	}
}

func TestSchedulerWithDifferentRatings(t *testing.T) {
	scheduler := NewScheduler()

	// Test that different ratings produce different scheduling outcomes
	baseCard := scheduler.NewCard()

	againResult := scheduler.ReviewCard(baseCard, fsrs.Again)
	hardResult := scheduler.ReviewCard(baseCard, fsrs.Hard)
	goodResult := scheduler.ReviewCard(baseCard, fsrs.Good)
	easyResult := scheduler.ReviewCard(baseCard, fsrs.Easy)

	// Easy should schedule further out than Good, which should be further than Again
	againDays := scheduler.DaysUntilDue(againResult.Card)
	hardDays := scheduler.DaysUntilDue(hardResult.Card)
	goodDays := scheduler.DaysUntilDue(goodResult.Card)
	easyDays := scheduler.DaysUntilDue(easyResult.Card)

	if !(againDays <= hardDays && hardDays <= goodDays && goodDays <= easyDays) {
		t.Errorf("Expected Again(%d) <= Hard(%d) <= Good(%d) <= Easy(%d) days until due", againDays, hardDays, goodDays, easyDays)
	}
}
