package domain

import (
	"testing"
	"time"
)

func TestParseRating(t *testing.T) {
	tests := []struct {
		input    string
		expected Rating
		wantErr  bool
	}{
		// Numeric inputs
		{"1", Again, false},
		{"2", Hard, false},
		{"3", Good, false},
		{"4", Easy, false},

		// String inputs (lowercase)
		{"again", Again, false},
		{"hard", Hard, false},
		{"good", Good, false},
		{"easy", Easy, false},

		// String inputs (uppercase)
		{"AGAIN", Again, false},
		{"HARD", Hard, false},
		{"GOOD", Good, false},
		{"EASY", Easy, false},

		// Single character inputs
		{"a", Again, false},
		{"A", Again, false},
		{"h", Hard, false},
		{"H", Hard, false},
		{"g", Good, false},
		{"G", Good, false},
		{"e", Easy, false},
		{"E", Easy, false},

		// Invalid inputs
		{"0", 0, true},
		{"5", 0, true},
		{"invalid", 0, true},
		{"", 0, true},
		{" ", 0, true},
		{"maybe", 0, true},
		{"ok", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			rating, err := ParseRating(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRating(%q) expected error but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRating(%q) unexpected error: %v", tt.input, err)
				return
			}

			if rating != tt.expected {
				t.Errorf("ParseRating(%q) = %v, expected %v", tt.input, rating, tt.expected)
			}
		})
	}
}

func TestRatingString(t *testing.T) {
	tests := []struct {
		rating   Rating
		expected string
	}{
		{Again, "Again"},
		{Hard, "Hard"},
		{Good, "Good"},
		{Easy, "Easy"},
		{Rating(99), "Unknown"}, // Unknown rating
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.rating.String()
			if result != tt.expected {
				t.Errorf("Rating(%v).String() = %q, expected %q", tt.rating, result, tt.expected)
			}
		})
	}
}

func TestRatingValues(t *testing.T) {
	// Test that rating constants have expected values for FSRS compatibility
	if int(Again) != 1 {
		t.Errorf("Again should be 1, got %d", int(Again))
	}
	if int(Hard) != 2 {
		t.Errorf("Hard should be 2, got %d", int(Hard))
	}
	if int(Good) != 3 {
		t.Errorf("Good should be 3, got %d", int(Good))
	}
	if int(Easy) != 4 {
		t.Errorf("Easy should be 4, got %d", int(Easy))
	}
}

func TestCardStateValues(t *testing.T) {
	// Test that CardState constants align with FSRS expectations
	if int(StateNew) != 0 {
		t.Errorf("StateNew should be 0, got %d", int(StateNew))
	}
	if int(StateLearning) != 1 {
		t.Errorf("StateLearning should be 1, got %d", int(StateLearning))
	}
	if int(StateReview) != 2 {
		t.Errorf("StateReview should be 2, got %d", int(StateReview))
	}
	if int(StateRelearning) != 3 {
		t.Errorf("StateRelearning should be 3, got %d", int(StateRelearning))
	}
}

func TestExecutionResult(t *testing.T) {
	// Test that ExecutionResult can be created and used
	result := &ExecutionResult{
		Success:        true,
		ExitCode:       0,
		Stdout:         "hello world",
		Stderr:         "",
		Duration:       100 * time.Millisecond,
		ThinkingTime:   2 * time.Second,
		ContainerID:    "test-container",
		ImageUsed:      "alpine:3.18",
		NetworkEnabled: false,
	}

	if !result.Success {
		t.Error("expected Success to be true")
	}

	if result.ExitCode != 0 {
		t.Errorf("expected ExitCode 0, got %d", result.ExitCode)
	}

	if result.Stdout != "hello world" {
		t.Errorf("expected stdout 'hello world', got %q", result.Stdout)
	}

	if result.Duration != 100*time.Millisecond {
		t.Errorf("expected duration 100ms, got %v", result.Duration)
	}

	if result.ThinkingTime != 2*time.Second {
		t.Errorf("expected thinking time 2s, got %v", result.ThinkingTime)
	}

	if result.ContainerID != "test-container" {
		t.Errorf("expected container ID 'test-container', got %q", result.ContainerID)
	}

	if result.NetworkEnabled {
		t.Error("expected NetworkEnabled to be false")
	}
}

func TestRatingEdgeCases(t *testing.T) {
	// Test whitespace handling
	tests := []struct {
		input    string
		expected Rating
	}{
		{" 1 ", Again},
		{"\t2\t", Hard},
		{"\n3\n", Good},
		{" good ", Good},
		{" EASY ", Easy},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			rating, err := ParseRating(tt.input)
			if err != nil {
				t.Errorf("ParseRating(%q) unexpected error: %v", tt.input, err)
				return
			}
			if rating != tt.expected {
				t.Errorf("ParseRating(%q) = %v, expected %v", tt.input, rating, tt.expected)
			}
		})
	}
}
