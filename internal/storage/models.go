package storage

import (
	"time"
)

// Deck represents a collection of flashcards with shared configuration
type Deck struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Version     string    `json:"version" db:"version"`
	Author      string    `json:"author" db:"author"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Sandbox defaults
	DefaultImage          string `json:"default_image" db:"default_image"`
	DefaultTimeout        int    `json:"default_timeout" db:"default_timeout"`
	DefaultNetworkEnabled bool   `json:"default_network_enabled" db:"default_network_enabled"`
	DefaultCapabilities   string `json:"default_capabilities" db:"default_capabilities"` // JSON array

	// FSRS parameters for this deck
	FSRSParameters string `json:"fsrs_parameters" db:"fsrs_parameters"` // JSON blob
}

// Card represents an individual flashcard with command execution details
type Card struct {
	ID     int `json:"id" db:"id"`
	DeckID int `json:"deck_id" db:"deck_id"`

	// Card identification
	CardKey     string `json:"card_key" db:"card_key"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`

	// Command execution
	Command         string `json:"command" db:"command"`
	WorkingDir      string `json:"working_dir" db:"working_dir"`
	EnvironmentVars string `json:"environment_vars" db:"environment_vars"` // JSON object

	// Sandbox overrides (NULL = use deck defaults)
	Image          *string `json:"image" db:"image"`
	Timeout        *int    `json:"timeout" db:"timeout"`
	NetworkEnabled *bool   `json:"network_enabled" db:"network_enabled"`
	Capabilities   *string `json:"capabilities" db:"capabilities"` // JSON array

	// Learning metadata
	DifficultyLevel int    `json:"difficulty_level" db:"difficulty_level"`
	Tags            string `json:"tags" db:"tags"` // JSON array

	// Prerequisites (symbolic linking approach)
	Prerequisites    string `json:"prerequisites" db:"prerequisites"`         // JSON array of card_keys
	PrerequisiteMode string `json:"prerequisite_mode" db:"prerequisite_mode"` // 'enforce' or 'link'

	// FSRS state - embedded for performance
	FSRSDue           time.Time  `json:"fsrs_due" db:"fsrs_due"`
	FSRSStability     float64    `json:"fsrs_stability" db:"fsrs_stability"`
	FSRSDifficulty    float64    `json:"fsrs_difficulty" db:"fsrs_difficulty"`
	FSRSElapsedDays   int        `json:"fsrs_elapsed_days" db:"fsrs_elapsed_days"`
	FSRSScheduledDays int        `json:"fsrs_scheduled_days" db:"fsrs_scheduled_days"`
	FSRSReps          int        `json:"fsrs_reps" db:"fsrs_reps"`
	FSRSLapses        int        `json:"fsrs_lapses" db:"fsrs_lapses"`
	FSRSState         int        `json:"fsrs_state" db:"fsrs_state"` // 0=New, 1=Learning, 2=Review, 3=Relearning
	FSRSLastReview    *time.Time `json:"fsrs_last_review" db:"fsrs_last_review"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Review represents a single review session of a card
type Review struct {
	ID     int `json:"id" db:"id"`
	CardID int `json:"card_id" db:"card_id"`

	// Review session
	ReviewedAt time.Time `json:"reviewed_at" db:"reviewed_at"`
	Rating     int       `json:"rating" db:"rating"` // 1=Again, 2=Hard, 3=Good, 4=Easy

	// Execution results
	ExecutionSuccess bool   `json:"execution_success" db:"execution_success"`
	ExitCode         *int   `json:"exit_code" db:"exit_code"`
	Stdout           string `json:"stdout" db:"stdout"`
	Stderr           string `json:"stderr" db:"stderr"`

	// Enhanced timing metrics
	ThinkingTimeMs  *int `json:"thinking_time_ms" db:"thinking_time_ms"`   // time from card shown to command started
	ExecutionTimeMs *int `json:"execution_time_ms" db:"execution_time_ms"` // actual command execution time
	TotalTimeMs     *int `json:"total_time_ms" db:"total_time_ms"`         // total time for the card

	// Interaction metrics
	Attempts     int  `json:"attempts" db:"attempts"`
	HelpAccessed bool `json:"help_accessed" db:"help_accessed"`

	// FSRS state transitions
	FSRSDueBefore        time.Time `json:"fsrs_due_before" db:"fsrs_due_before"`
	FSRSDueAfter         time.Time `json:"fsrs_due_after" db:"fsrs_due_after"`
	FSRSStabilityBefore  float64   `json:"fsrs_stability_before" db:"fsrs_stability_before"`
	FSRSStabilityAfter   float64   `json:"fsrs_stability_after" db:"fsrs_stability_after"`
	FSRSDifficultyBefore float64   `json:"fsrs_difficulty_before" db:"fsrs_difficulty_before"`
	FSRSDifficultyAfter  float64   `json:"fsrs_difficulty_after" db:"fsrs_difficulty_after"`
}

// DeckAsset represents a supporting file that cards within a deck can reference
// Cards reference these assets by filename in their commands (e.g., "cp /assets/config.json /etc/")
type DeckAsset struct {
	ID          int       `json:"id" db:"id"`
	DeckID      int       `json:"deck_id" db:"deck_id"`
	Filename    string    `json:"filename" db:"filename"`
	Content     []byte    `json:"content" db:"content"`
	ContentType string    `json:"content_type" db:"content_type"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// DeckVersion tracks changes to decks for in-place updates
type DeckVersion struct {
	ID        int       `json:"id" db:"id"`
	DeckID    int       `json:"deck_id" db:"deck_id"`
	Version   string    `json:"version" db:"version"`
	Changes   string    `json:"changes" db:"changes"` // JSON diff of what changed
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
