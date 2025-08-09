package domain

import (
	"strings"
	"time"
)

// Rating represents the user's assessment of their recall (shared across review/scheduler)
type Rating int

const (
	Again Rating = iota + 1 // 1 - Complete failure, need to see again soon
	Hard                    // 2 - Difficult, longer interval than Again but shorter than Good
	Good                    // 3 - Correct response, standard interval
	Easy                    // 4 - Too easy, much longer interval
)

// String returns the string representation of a rating
func (r Rating) String() string {
	switch r {
	case Again:
		return "Again"
	case Hard:
		return "Hard"
	case Good:
		return "Good"
	case Easy:
		return "Easy"
	default:
		return "Unknown"
	}
}

// ParseRating parses a string or number into a Rating
func ParseRating(input string) (Rating, error) {
	// Trim whitespace
	input = strings.TrimSpace(input)

	switch input {
	case "1", "again", "Again", "AGAIN", "a", "A":
		return Again, nil
	case "2", "hard", "Hard", "HARD", "h", "H":
		return Hard, nil
	case "3", "good", "Good", "GOOD", "g", "G":
		return Good, nil
	case "4", "easy", "Easy", "EASY", "e", "E":
		return Easy, nil
	default:
		return 0, &InvalidRatingError{Input: input}
	}
}

// CardState represents the FSRS learning state (shared across storage/scheduler)
type CardState int

const (
	StateNew        CardState = iota // 0 - New card, never studied
	StateLearning                    // 1 - Currently learning
	StateReview                      // 2 - In review phase
	StateRelearning                  // 3 - Forgotten, relearning
)

// String returns the string representation of a card state
func (s CardState) String() string {
	switch s {
	case StateNew:
		return "New"
	case StateLearning:
		return "Learning"
	case StateReview:
		return "Review"
	case StateRelearning:
		return "Relearning"
	default:
		return "Unknown"
	}
}

// ExecutionResult captures command execution results (shared across sandbox/review)
type ExecutionResult struct {
	Success        bool          `json:"success"`
	ExitCode       int           `json:"exit_code"`
	Stdout         string        `json:"stdout"`
	Stderr         string        `json:"stderr"`
	Duration       time.Duration `json:"duration"`
	ThinkingTime   time.Duration `json:"thinking_time"` // Time from card shown to command executed
	ContainerID    string        `json:"container_id"`
	ImageUsed      string        `json:"image_used"`
	NetworkEnabled bool          `json:"network_enabled"`
}

// InvalidRatingError indicates an invalid rating input
type InvalidRatingError struct {
	Input string
}

func (e *InvalidRatingError) Error() string {
	return "invalid rating: " + e.Input + " (valid: 1-4, Again/Hard/Good/Easy, a/h/g/e)"
}
