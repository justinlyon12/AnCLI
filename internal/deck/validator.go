package deck

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationResult contains the outcome of deck validation
type ValidationResult struct {
	Valid    bool                `json:"valid"`
	Errors   []ValidationError   `json:"errors"`
	Warnings []ValidationWarning `json:"warnings"`
	Info     []ValidationInfo    `json:"info"`
}

// ValidationError represents a validation error that prevents deck use
type ValidationError struct {
	Level   string `json:"level"`   // Always "error"
	File    string `json:"file"`    // File where error occurred
	Line    int    `json:"line"`    // Line number (0 if not applicable)
	Column  int    `json:"column"`  // Column number (0 if not applicable)
	Code    string `json:"code"`    // Error code like "DECK001"
	Message string `json:"message"` // Human-readable message
	Details string `json:"details"` // Additional context
}

// ValidationWarning represents a potential issue that doesn't prevent deck use
type ValidationWarning struct {
	Level   string `json:"level"` // Always "warning"
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

// ValidationInfo represents informational messages
type ValidationInfo struct {
	Level   string `json:"level"` // Always "info"
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Validation error codes
const (
	// Structure Errors (STRUCT)
	STRUCT001 = "STRUCT001" // Missing required file
	STRUCT002 = "STRUCT002" // Invalid file format
	STRUCT003 = "STRUCT003" // File encoding error

	// Deck Errors (DECK)
	DECK001 = "DECK001" // Missing required field
	DECK002 = "DECK002" // Invalid version format
	DECK003 = "DECK003" // Invalid container image
	DECK004 = "DECK004" // Invalid timeout value
	DECK005 = "DECK005" // Invalid FSRS parameters

	// Card Errors (CARD)
	CARD001 = "CARD001" // Duplicate card key
	CARD002 = "CARD002" // Missing required field
	CARD003 = "CARD003" // Invalid prerequisite reference
	CARD004 = "CARD004" // Circular dependency detected
	CARD005 = "CARD005" // Command syntax error
	CARD006 = "CARD006" // Setup without cleanup

	// Security Warnings (SEC)
	SEC001 = "SEC001" // Network enabled globally
	SEC002 = "SEC002" // Dangerous capability requested
	SEC003 = "SEC003" // Privileged container detected
	SEC004 = "SEC004" // Write access to host filesystem

	// Performance Warnings (PERF)
	PERF001 = "PERF001" // Timeout too short (<5s)
	PERF002 = "PERF002" // Timeout too long (>300s)
	PERF003 = "PERF003" // Excessive memory limit
	PERF004 = "PERF004" // Too many prerequisites

	// Usability Warnings (UX)
	UX001 = "UX001" // Missing description
	UX002 = "UX002" // No hint provided
	UX003 = "UX003" // Difficulty progression issue
	UX004 = "UX004" // Learning path too long
)

// DeckSpec represents the parsed deck.yaml structure
type DeckSpec struct {
	Name            string   `yaml:"name"`
	Version         string   `yaml:"version"`
	Author          string   `yaml:"author"`
	Description     string   `yaml:"description"`
	Tags            []string `yaml:"tags"`
	License         string   `yaml:"license"`
	DifficultyRange []int    `yaml:"difficulty_range"`

	Container struct {
		Image       string            `yaml:"image"`
		Timeout     int               `yaml:"timeout"`
		Network     bool              `yaml:"network"`
		Environment map[string]string `yaml:"environment"`
		WorkingDir  string            `yaml:"working_dir"`
	} `yaml:"container"`

	Cleanup struct {
		Mode           string `yaml:"mode"`
		PreserveOnFail bool   `yaml:"preserve_on_fail"`
		Timeout        int    `yaml:"timeout"`
	} `yaml:"cleanup"`

	FSRS struct {
		RequestRetention  float64 `yaml:"request_retention"`
		MaximumInterval   int     `yaml:"maximum_interval"`
		InitialDifficulty float64 `yaml:"initial_difficulty"`
	} `yaml:"fsrs"`

	Settings struct {
		ShuffleCards     bool   `yaml:"shuffle_cards"`
		PrerequisiteMode string `yaml:"prerequisite_mode"`
		ShowSolutions    bool   `yaml:"show_solutions"`
		ShowExplanations bool   `yaml:"show_explanations"`
		AutoCleanup      bool   `yaml:"auto_cleanup"`
	} `yaml:"settings"`
}

// CardSpec represents a parsed card from CSV
type CardSpec struct {
	Key           string
	Title         string
	Command       string
	Description   string
	Setup         string
	Cleanup       string
	Prerequisites string
	Verify        string
	Hint          string
	Solution      string
	Explanation   string
	Difficulty    int
	Tags          string
}

// ValidateDeck performs comprehensive validation of a deck directory
func ValidateDeck(deckPath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Info:     []ValidationInfo{},
	}

	// Phase 1: Structure validation
	if err := validateStructure(deckPath, result); err != nil {
		return result, err
	}

	// Stop if structure validation failed
	if len(result.Errors) > 0 {
		result.Valid = false
		return result, nil
	}

	// Phase 2: Parse and validate deck.yaml
	deckSpec, err := parseDeckYAML(deckPath, result)
	if err != nil {
		return result, err
	}

	// Phase 3: Parse and validate cards.csv
	cards, err := parseCardsCSV(deckPath, result)
	if err != nil {
		return result, err
	}

	// Phase 4: Cross-validation between deck and cards
	validateDeckCardConsistency(deckSpec, cards, result)

	// Phase 5: Dependency graph validation
	validateDependencyGraph(cards, result)

	// Phase 6: Security validation
	validateSecurity(deckSpec, cards, result)

	// Phase 7: Usability validation
	validateUsability(deckSpec, cards, result)

	// Set final validation result
	result.Valid = len(result.Errors) == 0

	return result, nil
}

// validateStructure checks that required files exist and are readable
func validateStructure(deckPath string, result *ValidationResult) error {
	// Check if deck path exists
	if _, err := os.Stat(deckPath); os.IsNotExist(err) {
		result.Errors = append(result.Errors, ValidationError{
			Level:   "error",
			File:    deckPath,
			Code:    STRUCT001,
			Message: "Deck directory does not exist",
			Details: fmt.Sprintf("Path: %s", deckPath),
		})
		return nil
	}

	// Required files
	requiredFiles := []string{"deck.yaml", "cards.csv"}

	for _, file := range requiredFiles {
		filePath := filepath.Join(deckPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			result.Errors = append(result.Errors, ValidationError{
				Level:   "error",
				File:    file,
				Code:    STRUCT001,
				Message: fmt.Sprintf("Required file '%s' is missing", file),
				Details: "Every deck must have deck.yaml and cards.csv",
			})
		}
	}

	// Optional files
	optionalFiles := []string{"README.md", "assets"}
	for _, file := range optionalFiles {
		filePath := filepath.Join(deckPath, file)
		if _, err := os.Stat(filePath); err == nil {
			result.Info = append(result.Info, ValidationInfo{
				Level:   "info",
				Code:    "INFO001",
				Message: fmt.Sprintf("Found optional file: %s", file),
			})
		}
	}

	return nil
}

// parseDeckYAML reads and validates the deck.yaml file
func parseDeckYAML(deckPath string, result *ValidationResult) (*DeckSpec, error) {
	filePath := filepath.Join(deckPath, "deck.yaml")

	data, err := os.ReadFile(filePath)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Level:   "error",
			File:    "deck.yaml",
			Code:    STRUCT003,
			Message: "Failed to read deck.yaml",
			Details: err.Error(),
		})
		return nil, nil
	}

	var spec DeckSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Level:   "error",
			File:    "deck.yaml",
			Code:    STRUCT002,
			Message: "Invalid YAML format in deck.yaml",
			Details: err.Error(),
		})
		return nil, nil
	}

	// Validate required fields
	requiredFields := map[string]string{
		"name":        spec.Name,
		"version":     spec.Version,
		"author":      spec.Author,
		"description": spec.Description,
	}

	for field, value := range requiredFields {
		if strings.TrimSpace(value) == "" {
			result.Errors = append(result.Errors, ValidationError{
				Level:   "error",
				File:    "deck.yaml",
				Code:    DECK001,
				Message: fmt.Sprintf("Required field '%s' is missing or empty", field),
				Details: "All decks must have name, version, author, and description",
			})
		}
	}

	// Validate version format (semantic versioning)
	versionPattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if !versionPattern.MatchString(spec.Version) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Level:   "warning",
			File:    "deck.yaml",
			Code:    DECK002,
			Message: "Version should follow semantic versioning (x.y.z)",
			Details: fmt.Sprintf("Current version: %s", spec.Version),
		})
	}

	// Validate container settings
	validateContainerSpec(&spec, result)

	return &spec, nil
}

// validateContainerSpec validates container configuration
func validateContainerSpec(spec *DeckSpec, result *ValidationResult) {
	// Check timeout values
	if spec.Container.Timeout > 0 && spec.Container.Timeout < 5 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Level:   "warning",
			File:    "deck.yaml",
			Code:    PERF001,
			Message: "Container timeout is very short (<5s)",
			Details: "Short timeouts may cause legitimate commands to fail",
		})
	}

	if spec.Container.Timeout > 300 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Level:   "warning",
			File:    "deck.yaml",
			Code:    PERF002,
			Message: "Container timeout is very long (>300s)",
			Details: "Long timeouts may indicate inefficient commands",
		})
	}

	// Security warnings
	if spec.Container.Network {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Level:   "warning",
			File:    "deck.yaml",
			Code:    SEC001,
			Message: "Network access is enabled globally",
			Details: "Consider enabling network only for specific cards that need it",
		})
	}
}

// parseCardsCSV reads and validates the cards.csv file
func parseCardsCSV(deckPath string, result *ValidationResult) ([]CardSpec, error) {
	filePath := filepath.Join(deckPath, "cards.csv")

	file, err := os.Open(filePath)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Level:   "error",
			File:    "cards.csv",
			Code:    STRUCT003,
			Message: "Failed to read cards.csv",
			Details: err.Error(),
		})
		return nil, nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Level:   "error",
			File:    "cards.csv",
			Code:    STRUCT002,
			Message: "Invalid CSV format in cards.csv",
			Details: err.Error(),
		})
		return nil, nil
	}

	if len(records) < 2 {
		result.Errors = append(result.Errors, ValidationError{
			Level:   "error",
			File:    "cards.csv",
			Code:    CARD002,
			Message: "cards.csv must have header row and at least one card",
			Details: "Empty decks are not allowed",
		})
		return nil, nil
	}

	// Validate header
	expectedHeader := []string{"key", "title", "command", "description", "setup", "cleanup", "prerequisites", "verify", "hint", "solution", "explanation", "difficulty", "tags"}
	header := records[0]

	for i, expected := range expectedHeader {
		if i >= len(header) || header[i] != expected {
			result.Errors = append(result.Errors, ValidationError{
				Level:   "error",
				File:    "cards.csv",
				Line:    1,
				Column:  i + 1,
				Code:    STRUCT002,
				Message: fmt.Sprintf("CSV header mismatch at column %d: expected '%s', got '%s'", i+1, expected, getColumn(header, i)),
				Details: "CSV header must match expected format exactly",
			})
		}
	}

	if len(result.Errors) > 0 {
		return nil, nil // Header validation failed
	}

	// Parse card records
	var cards []CardSpec
	cardKeys := make(map[string]int) // Track duplicate keys

	for lineNum, record := range records[1:] {
		line := lineNum + 2 // +1 for 0-based, +1 for header

		if len(record) != len(expectedHeader) {
			result.Errors = append(result.Errors, ValidationError{
				Level:   "error",
				File:    "cards.csv",
				Line:    line,
				Code:    STRUCT002,
				Message: fmt.Sprintf("Card at line %d has %d fields, expected %d", line, len(record), len(expectedHeader)),
				Details: "All cards must have all CSV fields (can be empty)",
			})
			continue
		}

		difficulty, err := strconv.Atoi(record[11])
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Level:   "error",
				File:    "cards.csv",
				Line:    line,
				Column:  12,
				Code:    CARD002,
				Message: "Difficulty must be a number",
				Details: fmt.Sprintf("Got: '%s'", record[11]),
			})
			continue
		}

		card := CardSpec{
			Key:           strings.TrimSpace(record[0]),
			Title:         strings.TrimSpace(record[1]),
			Command:       strings.TrimSpace(record[2]),
			Description:   strings.TrimSpace(record[3]),
			Setup:         strings.TrimSpace(record[4]),
			Cleanup:       strings.TrimSpace(record[5]),
			Prerequisites: strings.TrimSpace(record[6]),
			Verify:        strings.TrimSpace(record[7]),
			Hint:          strings.TrimSpace(record[8]),
			Solution:      strings.TrimSpace(record[9]),
			Explanation:   strings.TrimSpace(record[10]),
			Difficulty:    difficulty,
			Tags:          strings.TrimSpace(record[12]),
		}

		// Validate required fields
		validateCardFields(card, line, result)

		// Check for duplicate keys
		if prevLine, exists := cardKeys[card.Key]; exists {
			result.Errors = append(result.Errors, ValidationError{
				Level:   "error",
				File:    "cards.csv",
				Line:    line,
				Code:    CARD001,
				Message: fmt.Sprintf("Duplicate card key '%s' (also used at line %d)", card.Key, prevLine),
				Details: "Card keys must be unique within a deck",
			})
		} else {
			cardKeys[card.Key] = line
		}

		cards = append(cards, card)
	}

	return cards, nil
}

// validateCardFields validates individual card field requirements
func validateCardFields(card CardSpec, line int, result *ValidationResult) {
	// Required fields
	requiredFields := map[string]string{
		"key":         card.Key,
		"title":       card.Title,
		"command":     card.Command,
		"description": card.Description,
	}

	for field, value := range requiredFields {
		if value == "" {
			result.Errors = append(result.Errors, ValidationError{
				Level:   "error",
				File:    "cards.csv",
				Line:    line,
				Code:    CARD002,
				Message: fmt.Sprintf("Required field '%s' is empty", field),
				Details: "Key, title, command, and description are required for all cards",
			})
		}
	}

	// Validate key format (alphanumeric, hyphens, underscores)
	keyPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !keyPattern.MatchString(card.Key) {
		result.Errors = append(result.Errors, ValidationError{
			Level:   "error",
			File:    "cards.csv",
			Line:    line,
			Code:    CARD005,
			Message: "Card key contains invalid characters",
			Details: "Keys must contain only letters, numbers, hyphens, and underscores",
		})
	}

	// Difficulty range validation
	if card.Difficulty < 1 || card.Difficulty > 6 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Level:   "warning",
			File:    "cards.csv",
			Line:    line,
			Code:    UX003,
			Message: fmt.Sprintf("Difficulty %d is outside recommended range (1-6)", card.Difficulty),
			Details: "Standard difficulty: 1=trivial, 2=easy, 3=medium, 4=hard, 5=expert, 6=insane",
		})
	}

	// Setup/cleanup pairing
	if card.Setup != "" && card.Cleanup == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Level:   "warning",
			File:    "cards.csv",
			Line:    line,
			Code:    CARD006,
			Message: "Card has setup command but no cleanup command",
			Details: "Setup commands should have corresponding cleanup for repeatability",
		})
	}

	// Usability warnings
	if card.Hint == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Level:   "warning",
			File:    "cards.csv",
			Line:    line,
			Code:    UX002,
			Message: "Card has no hint provided",
			Details: "Hints help users learn without giving away the answer",
		})
	}

	if card.Explanation == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Level:   "warning",
			File:    "cards.csv",
			Line:    line,
			Code:    UX001,
			Message: "Card has no explanation provided",
			Details: "Explanations help users understand command output and concepts",
		})
	}
}

// validateDependencyGraph checks prerequisites for cycles and missing references
func validateDependencyGraph(cards []CardSpec, result *ValidationResult) {
	cardMap := make(map[string]*CardSpec)
	for i := range cards {
		cardMap[cards[i].Key] = &cards[i]
	}

	// Build dependency graph and check for missing prerequisites
	graph := make(map[string][]string)
	for _, card := range cards {
		if card.Prerequisites == "" {
			continue
		}

		prereqs := strings.Split(card.Prerequisites, ",")
		for _, prereq := range prereqs {
			prereq = strings.TrimSpace(prereq)
			if prereq == "" {
				continue
			}

			// Check if prerequisite exists
			if _, exists := cardMap[prereq]; !exists {
				result.Errors = append(result.Errors, ValidationError{
					Level:   "error",
					File:    "cards.csv",
					Code:    CARD003,
					Message: fmt.Sprintf("Card '%s' references non-existent prerequisite '%s'", card.Key, prereq),
					Details: "All prerequisites must reference existing card keys",
				})
				continue
			}

			graph[card.Key] = append(graph[card.Key], prereq)
		}
	}

	// Check for circular dependencies using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for cardKey := range cardMap {
		if !visited[cardKey] {
			if hasCycle(cardKey, graph, visited, recStack) {
				result.Errors = append(result.Errors, ValidationError{
					Level:   "error",
					File:    "cards.csv",
					Code:    CARD004,
					Message: fmt.Sprintf("Circular dependency detected involving card '%s'", cardKey),
					Details: "Prerequisites must form a directed acyclic graph (DAG)",
				})
			}
		}
	}
}

// hasCycle detects cycles in dependency graph using DFS
func hasCycle(node string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range graph[node] {
		if !visited[neighbor] && hasCycle(neighbor, graph, visited, recStack) {
			return true
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[node] = false
	return false
}

// validateSecurity checks for security-related issues
func validateSecurity(spec *DeckSpec, cards []CardSpec, result *ValidationResult) {
	// Check for dangerous commands
	dangerousPatterns := []string{
		`rm\s+-rf\s+/`,   // Dangerous rm commands
		`sudo`,           // Privilege escalation
		`su\s`,           // User switching
		`chmod\s+777`,    // Overly permissive permissions
		`wget\s+http://`, // Unencrypted downloads
		`curl.*http://`,  // Unencrypted HTTP
	}

	for _, card := range cards {
		commands := []string{card.Command, card.Setup, card.Cleanup}
		for _, cmd := range commands {
			if cmd == "" {
				continue
			}

			for _, pattern := range dangerousPatterns {
				if matched, _ := regexp.MatchString(pattern, cmd); matched {
					result.Warnings = append(result.Warnings, ValidationWarning{
						Level:   "warning",
						File:    "cards.csv",
						Code:    SEC003,
						Message: fmt.Sprintf("Card '%s' contains potentially dangerous command", card.Key),
						Details: fmt.Sprintf("Pattern matched: %s", pattern),
					})
				}
			}
		}
	}
}

// validateUsability checks for usability and learning experience issues
func validateUsability(spec *DeckSpec, cards []CardSpec, result *ValidationResult) {
	// Check difficulty progression
	difficulties := make([]int, len(cards))
	for i, card := range cards {
		difficulties[i] = card.Difficulty
	}

	// Look for large difficulty jumps
	for i := 1; i < len(difficulties); i++ {
		jump := difficulties[i] - difficulties[i-1]
		if jump > 2 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Level: "warning",
				File:  "cards.csv",
				Code:  UX003,
				Message: fmt.Sprintf("Large difficulty jump from card %d to %d (difficulty %d to %d)",
					i, i+1, difficulties[i-1], difficulties[i]),
				Details: "Consider adding intermediate difficulty cards for smoother learning progression",
			})
		}
	}

	// Check for very long prerequisite chains
	for _, card := range cards {
		chainLength := getPrerequisiteChainLength(card.Key, cards, make(map[string]bool))
		if chainLength > 5 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Level:   "warning",
				File:    "cards.csv",
				Code:    UX004,
				Message: fmt.Sprintf("Card '%s' has very long prerequisite chain (%d levels)", card.Key, chainLength),
				Details: "Long chains may frustrate users - consider restructuring prerequisites",
			})
		}
	}
}

// validateDeckCardConsistency ensures deck and cards are consistent
func validateDeckCardConsistency(spec *DeckSpec, cards []CardSpec, result *ValidationResult) {
	if len(cards) == 0 {
		return
	}

	// Check if difficulty range in deck matches actual cards
	if len(spec.DifficultyRange) == 2 {
		minDifficulty, maxDifficulty := spec.DifficultyRange[0], spec.DifficultyRange[1]

		for _, card := range cards {
			if card.Difficulty < minDifficulty || card.Difficulty > maxDifficulty {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Level: "warning",
					File:  "deck.yaml",
					Code:  DECK005,
					Message: fmt.Sprintf("Card '%s' difficulty (%d) outside declared range [%d, %d]",
						card.Key, card.Difficulty, minDifficulty, maxDifficulty),
					Details: "Update difficulty_range in deck.yaml or adjust card difficulties",
				})
			}
		}
	}
}

// Helper functions

func getColumn(slice []string, index int) string {
	if index < len(slice) {
		return slice[index]
	}
	return "<missing>"
}

func getPrerequisiteChainLength(cardKey string, cards []CardSpec, visited map[string]bool) int {
	if visited[cardKey] {
		return 0 // Circular reference protection
	}

	visited[cardKey] = true
	defer func() { visited[cardKey] = false }()

	// Find the card
	var card *CardSpec
	for i := range cards {
		if cards[i].Key == cardKey {
			card = &cards[i]
			break
		}
	}

	if card == nil || card.Prerequisites == "" {
		return 0
	}

	// Find maximum chain length among prerequisites
	maxLength := 0
	prereqs := strings.Split(card.Prerequisites, ",")
	for _, prereq := range prereqs {
		prereq = strings.TrimSpace(prereq)
		if prereq == "" {
			continue
		}

		length := getPrerequisiteChainLength(prereq, cards, visited)
		if length > maxLength {
			maxLength = length
		}
	}

	return maxLength + 1
}

// PrintValidationResult outputs validation results in a human-readable format
func PrintValidationResult(result *ValidationResult, verbose bool) {
	if result.Valid {
		fmt.Println("✅ Deck validation passed!")
	} else {
		fmt.Println("❌ Deck validation failed!")
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n🚨 Errors (%d):\n", len(result.Errors))
		for _, err := range result.Errors {
			location := err.File
			if err.Line > 0 {
				location += fmt.Sprintf(":%d", err.Line)
				if err.Column > 0 {
					location += fmt.Sprintf(":%d", err.Column)
				}
			}
			fmt.Printf("  %s [%s] %s\n", location, err.Code, err.Message)
			if verbose && err.Details != "" {
				fmt.Printf("    %s\n", err.Details)
			}
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings (%d):\n", len(result.Warnings))
		for _, warn := range result.Warnings {
			location := warn.File
			if warn.Line > 0 {
				location += fmt.Sprintf(":%d", warn.Line)
			}
			fmt.Printf("  %s [%s] %s\n", location, warn.Code, warn.Message)
			if verbose && warn.Details != "" {
				fmt.Printf("    %s\n", warn.Details)
			}
		}
	}

	if verbose && len(result.Info) > 0 {
		fmt.Printf("\nℹ️  Info (%d):\n", len(result.Info))
		for _, info := range result.Info {
			fmt.Printf("  %s: %s\n", info.Code, info.Message)
		}
	}

	fmt.Printf("\nSummary: %d errors, %d warnings\n", len(result.Errors), len(result.Warnings))
}
