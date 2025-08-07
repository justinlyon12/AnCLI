package scheduler

import (
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// Scheduler wraps the FSRS algorithm for our CLI application
type Scheduler struct {
	fsrs *fsrs.FSRS
}

// NewScheduler creates a new scheduler with default FSRS parameters
func NewScheduler() *Scheduler {
	return &Scheduler{
		fsrs: fsrs.NewFSRS(fsrs.DefaultParam()),
	}
}

// NewSchedulerWithParams creates a new scheduler with custom FSRS parameters
func NewSchedulerWithParams(params fsrs.Parameters) *Scheduler {
	return &Scheduler{
		fsrs: fsrs.NewFSRS(params),
	}
}

// NewCard creates a new card with initial FSRS state
func (s *Scheduler) NewCard() fsrs.Card {
	return fsrs.NewCard()
}

// ReviewCard processes a card review and returns the updated card
// rating should be one of: fsrs.Again, fsrs.Hard, fsrs.Good, fsrs.Easy
func (s *Scheduler) ReviewCard(card fsrs.Card, rating fsrs.Rating) fsrs.SchedulingInfo {
	now := time.Now()
	return s.fsrs.Next(card, now, rating)
}

// GetSchedulingOptions returns all possible scheduling outcomes for a card
// This allows the UI to show the user what will happen for each rating choice
func (s *Scheduler) GetSchedulingOptions(card fsrs.Card) fsrs.RecordLog {
	now := time.Now()
	return s.fsrs.Repeat(card, now)
}

// IsDue checks if a card is due for review
func (s *Scheduler) IsDue(card fsrs.Card) bool {
	return time.Now().After(card.Due) || time.Now().Equal(card.Due)
}

// GetRetrievability returns the current retrievability of a card (0.0 to 1.0)
func (s *Scheduler) GetRetrievability(card fsrs.Card) float64 {
	now := time.Now()
	return s.fsrs.GetRetrievability(card, now)
}

// DaysUntilDue returns the number of days until the card is due
// Returns 0 if the card is already due
func (s *Scheduler) DaysUntilDue(card fsrs.Card) int {
	now := time.Now()
	if s.IsDue(card) {
		return 0
	}
	duration := card.Due.Sub(now)
	return int(duration.Hours() / 24)
}
