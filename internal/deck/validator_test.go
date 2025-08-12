package deck

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDeck_BasicFunctionality(t *testing.T) {
	// Create a basic test that validates core functionality works
	tmpDir, err := os.MkdirTemp("", "deck-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create minimal valid deck
	createFile(t, filepath.Join(tmpDir, "deck.yaml"), `
name: test-deck
version: 1.0.0
author: Test Author
description: A test deck
`)
	createFile(t, filepath.Join(tmpDir, "cards.csv"), `
key,title,command,description,setup,cleanup,prerequisites,verify,hint,solution,explanation,difficulty,tags
basic,"Basic","echo hello","Print hello",,,,,"Use echo","echo hello","Prints hello",1,"basic"
`)

	// Validate deck
	result, err := ValidateDeck(tmpDir)
	if err != nil {
		t.Fatalf("ValidateDeck returned error: %v", err)
	}

	// Should have a result
	if result == nil {
		t.Fatal("ValidateDeck returned nil result")
	}

	// Test with non-existent directory
	result2, err := ValidateDeck("/nonexistent/path")
	if err != nil {
		t.Fatalf("ValidateDeck with non-existent path returned error: %v", err)
	}

	if result2 == nil {
		t.Fatal("ValidateDeck with non-existent path returned nil result")
	}

	// Should be invalid
	if result2.Valid {
		t.Error("Expected validation to fail for non-existent path")
	}
}

func TestValidateStructure(t *testing.T) {
	tests := []struct {
		name         string
		setupFiles   func(string)
		expectedCode string
	}{
		{
			name: "non-existent directory fails",
			setupFiles: func(dir string) {
				// Don't create the directory
			},
			expectedCode: "STRUCT001",
		},
		{
			name: "directory with all required files passes",
			setupFiles: func(dir string) {
				createFile(t, filepath.Join(dir, "deck.yaml"), "name: test")
				createFile(t, filepath.Join(dir, "cards.csv"), "header")
			},
			expectedCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "struct-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			testDir := tmpDir
			if tt.name != "non-existent directory fails" {
				tt.setupFiles(testDir)
			} else {
				testDir = filepath.Join(tmpDir, "nonexistent")
			}

			result := &ValidationResult{}
			err = validateStructure(testDir, result)
			if err != nil {
				t.Fatalf("validateStructure returned error: %v", err)
			}

			if tt.expectedCode == "" {
				if len(result.Errors) > 0 {
					t.Errorf("expected no errors, got %d", len(result.Errors))
				}
			} else {
				found := false
				for _, e := range result.Errors {
					if e.Code == tt.expectedCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error code %s not found", tt.expectedCode)
				}
			}
		})
	}
}

func TestHasCycle(t *testing.T) {
	tests := []struct {
		name     string
		graph    map[string][]string
		start    string
		expected bool
	}{
		{
			name: "no cycle",
			graph: map[string][]string{
				"a": {"b"},
				"b": {"c"},
				"c": {},
			},
			start:    "a",
			expected: false,
		},
		{
			name: "simple cycle",
			graph: map[string][]string{
				"a": {"b"},
				"b": {"a"},
			},
			start:    "a",
			expected: true,
		},
		{
			name: "complex cycle",
			graph: map[string][]string{
				"a": {"b"},
				"b": {"c"},
				"c": {"a"},
			},
			start:    "a",
			expected: true,
		},
		{
			name: "self cycle",
			graph: map[string][]string{
				"a": {"a"},
			},
			start:    "a",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visited := make(map[string]bool)
			recStack := make(map[string]bool)

			result := hasCycle(tt.start, tt.graph, visited, recStack)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetPrerequisiteChainLength_BasicFunctionality(t *testing.T) {
	cards := []CardSpec{
		{Key: "a", Prerequisites: ""},
		{Key: "b", Prerequisites: "a"},
		{Key: "c", Prerequisites: "b"},
	}

	// Test basic chain length calculation
	visited := make(map[string]bool)
	length := getPrerequisiteChainLength("c", cards, visited)
	if length != 2 {
		t.Errorf("expected length 2, got %d", length)
	}

	// Test no prerequisites
	visited = make(map[string]bool)
	length = getPrerequisiteChainLength("a", cards, visited)
	if length != 0 {
		t.Errorf("expected length 0, got %d", length)
	}
}

func TestValidateCardFields_BasicFunctionality(t *testing.T) {
	// Test valid card
	validCard := CardSpec{
		Key:         "valid-key",
		Title:       "Valid Title",
		Command:     "echo hello",
		Description: "Valid description",
		Hint:        "Use echo",
		Explanation: "Prints hello",
		Difficulty:  3,
	}

	result := &ValidationResult{}
	validateCardFields(validCard, 1, result)

	// Should have no errors for valid card
	if len(result.Errors) > 0 {
		t.Errorf("expected no errors for valid card, got %d", len(result.Errors))
	}

	// Test missing required fields
	invalidCard := CardSpec{
		Key:         "",
		Title:       "",
		Command:     "",
		Description: "",
	}

	result = &ValidationResult{}
	validateCardFields(invalidCard, 1, result)

	// Should have errors for missing required fields
	if len(result.Errors) == 0 {
		t.Error("expected errors for missing required fields")
	}

	// Should find CARD002 errors
	found := false
	for _, err := range result.Errors {
		if err.Code == "CARD002" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CARD002 error for missing required fields")
	}
}

// Helper function to create test files
func createFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
