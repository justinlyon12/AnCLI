package storage

import (
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// ToFSRSCard converts a storage Card to an FSRS Card for scheduling
// This is a lightweight conversion used during review sessions
func (c *Card) ToFSRSCard() fsrs.Card {
	var lastReview time.Time
	if c.FSRSLastReview != nil {
		lastReview = *c.FSRSLastReview
	}

	return fsrs.Card{
		Due:           c.FSRSDue,
		Stability:     c.FSRSStability,
		Difficulty:    c.FSRSDifficulty,
		ElapsedDays:   uint64(c.FSRSElapsedDays),
		ScheduledDays: uint64(c.FSRSScheduledDays),
		Reps:          uint64(c.FSRSReps),
		Lapses:        uint64(c.FSRSLapses),
		State:         fsrs.State(c.FSRSState),
		LastReview:    lastReview,
	}
}

// UpdateFromFSRSCard updates the storage Card with FSRS scheduling results
// This is called after the scheduler processes a review
func (c *Card) UpdateFromFSRSCard(fsrsCard fsrs.Card) {
	c.FSRSDue = fsrsCard.Due
	c.FSRSStability = fsrsCard.Stability
	c.FSRSDifficulty = fsrsCard.Difficulty
	c.FSRSElapsedDays = int(fsrsCard.ElapsedDays)
	c.FSRSScheduledDays = int(fsrsCard.ScheduledDays)
	c.FSRSReps = int(fsrsCard.Reps)
	c.FSRSLapses = int(fsrsCard.Lapses)
	c.FSRSState = int(fsrsCard.State)
	c.FSRSLastReview = &fsrsCard.LastReview
	c.UpdatedAt = time.Now()
}
