package review

import (
	"context"
	"time"

	"github.com/justinlyon12/ancli/internal/domain"
)

// ReviewService defines the interface for managing review sessions
type ReviewService interface {
	// StartSession begins a new review session
	StartSession(ctx context.Context, opts SessionOptions) (*Session, error)

	// GetNextCard retrieves the next card due for review in the session
	GetNextCard(ctx context.Context, sessionID string) (*ReviewCard, error)

	// SubmitReview processes a card review and updates FSRS state
	SubmitReview(ctx context.Context, sessionID string, cardID int, rating domain.Rating, executionResult *domain.ExecutionResult) error

	// EndSession finalizes the review session and returns statistics
	EndSession(ctx context.Context, sessionID string) (*SessionStats, error)
}

// SessionOptions configures a review session
type SessionOptions struct {
	DeckID          *int // If nil, review from all decks
	MaxCards        int  // Maximum cards per session (0 = unlimited)
	NewCardsOnly    bool // Only show new cards
	ReviewCardsOnly bool // Only show cards due for review
	ShuffleCards    bool // Randomize card order
	NetworkEnabled  bool // Allow network access for this session
}

// Session represents an active review session
type Session struct {
	ID             string         `json:"id"`
	StartedAt      time.Time      `json:"started_at"`
	DeckID         *int           `json:"deck_id"`
	Options        SessionOptions `json:"options"`
	CardsReviewed  int            `json:"cards_reviewed"`
	CardsRemaining int            `json:"cards_remaining"`
	CurrentCardID  *int           `json:"current_card_id"`
}

// ReviewCard represents a card ready for review with resolved configuration
// Lean domain type that adapters map to/from storage types
type ReviewCard struct {
	// Card identification
	ID          int    `json:"id"`
	DeckID      int    `json:"deck_id"`
	CardKey     string `json:"card_key"`
	Title       string `json:"title"`
	Description string `json:"description"`

	// Command execution
	Command         string            `json:"command"`
	WorkingDir      string            `json:"working_dir"`
	EnvironmentVars map[string]string `json:"environment_vars"`

	// Resolved sandbox configuration (deck defaults + card overrides)
	Image          string        `json:"image"`
	Timeout        time.Duration `json:"timeout"`
	NetworkEnabled bool          `json:"network_enabled"`
	Capabilities   []string      `json:"capabilities"`

	// Learning metadata
	DifficultyLevel int      `json:"difficulty_level"`
	Tags            []string `json:"tags"`

	// FSRS state
	DueAt         time.Time        `json:"due_at"`
	Stability     float64          `json:"stability"`
	Difficulty    float64          `json:"difficulty"`
	ElapsedDays   int              `json:"elapsed_days"`
	ScheduledDays int              `json:"scheduled_days"`
	Reps          int              `json:"reps"`
	Lapses        int              `json:"lapses"`
	State         domain.CardState `json:"state"`
	LastReview    *time.Time       `json:"last_review"`
}

// SessionStats provides summary information about a completed session
type SessionStats struct {
	SessionID     string        `json:"session_id"`
	Duration      time.Duration `json:"duration"`
	CardsReviewed int           `json:"cards_reviewed"`
	NewCards      int           `json:"new_cards"`
	ReviewCards   int           `json:"review_cards"`
	AgainCount    int           `json:"again_count"`
	HardCount     int           `json:"hard_count"`
	GoodCount     int           `json:"good_count"`
	EasyCount     int           `json:"easy_count"`
	AverageRating float64       `json:"average_rating"`
}
