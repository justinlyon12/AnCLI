package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	// Create temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestDeckOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Test CreateDeck
	deck := &Deck{
		Name:                  "Test Deck",
		Description:           "A test deck for unit testing",
		Version:               "1.0.0",
		Author:                "Test Author",
		DefaultImage:          "alpine:latest",
		DefaultTimeout:        10,
		DefaultNetworkEnabled: false,
		DefaultCapabilities:   `["NET_ADMIN"]`,
		FSRSParameters:        `{"w":[1,2,3,4]}`,
	}

	err := db.CreateDeck(deck)
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}

	if deck.ID == 0 {
		t.Error("Expected deck ID to be set after creation")
	}

	// Test GetDeck
	retrieved, err := db.GetDeck(deck.ID)
	if err != nil {
		t.Fatalf("Failed to get deck: %v", err)
	}

	if retrieved.Name != deck.Name {
		t.Errorf("Expected name %s, got %s", deck.Name, retrieved.Name)
	}

	// Test ListDecks
	decks, err := db.ListDecks()
	if err != nil {
		t.Fatalf("Failed to list decks: %v", err)
	}

	if len(decks) != 1 {
		t.Errorf("Expected 1 deck, got %d", len(decks))
	}
}

func TestCardOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a deck first
	deck := &Deck{
		Name:        "Test Deck",
		Description: "Test deck for card operations",
	}
	err := db.CreateDeck(deck)
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}

	// Test CreateCard
	card := &Card{
		DeckID:           deck.ID,
		CardKey:          "test-card-1",
		Title:            "Test Card",
		Description:      "A test card",
		Command:          "echo 'Hello World'",
		WorkingDir:       "/tmp",
		EnvironmentVars:  `{"TEST_VAR":"test_value"}`,
		DifficultyLevel:  2,
		Tags:             `["linux","basic"]`,
		Prerequisites:    `["prerequisite-card"]`,
		PrerequisiteMode: "link",
	}

	err = db.CreateCard(card)
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	if card.ID == 0 {
		t.Error("Expected card ID to be set after creation")
	}

	// Verify FSRS state was initialized
	if card.FSRSState != 0 { // New card state
		t.Errorf("Expected FSRS state 0 (New), got %d", card.FSRSState)
	}

	// Test GetCard
	retrieved, err := db.GetCard(card.ID)
	if err != nil {
		t.Fatalf("Failed to get card: %v", err)
	}

	if retrieved.Title != card.Title {
		t.Errorf("Expected title %s, got %s", card.Title, retrieved.Title)
	}

	// Test GetDueCards (new card should be due immediately)
	dueCards, err := db.GetDueCards()
	if err != nil {
		t.Fatalf("Failed to get due cards: %v", err)
	}

	if len(dueCards) != 1 {
		t.Errorf("Expected 1 due card, got %d", len(dueCards))
	}
}

func TestFSRSIntegration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create deck and card
	deck := &Deck{Name: "FSRS Test Deck"}
	err := db.CreateDeck(deck)
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}

	card := &Card{
		DeckID:  deck.ID,
		CardKey: "fsrs-test-card",
		Title:   "FSRS Test Card",
		Command: "echo 'FSRS test'",
	}
	err = db.CreateCard(card)
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	// Test ToFSRSCard conversion
	fsrsCard := card.ToFSRSCard()
	if fsrsCard.State != fsrs.New {
		t.Errorf("Expected FSRS state New, got %v", fsrsCard.State)
	}

	// Simulate a review with Good rating
	scheduler := fsrs.NewFSRS(fsrs.DefaultParam())
	schedulingInfo := scheduler.Next(fsrsCard, time.Now(), fsrs.Good)

	// Update card with new FSRS state
	card.UpdateFromFSRSCard(schedulingInfo.Card)

	// Verify the card was updated
	if card.FSRSReps != 1 {
		t.Errorf("Expected 1 rep after review, got %d", card.FSRSReps)
	}

	if card.FSRSState != int(fsrs.Learning) {
		t.Errorf("Expected Learning state after first review, got %d", card.FSRSState)
	}

	// Test UpdateCardFSRS
	err = db.UpdateCardFSRS(card)
	if err != nil {
		t.Fatalf("Failed to update card FSRS state: %v", err)
	}

	// Verify the update persisted
	updated, err := db.GetCard(card.ID)
	if err != nil {
		t.Fatalf("Failed to get updated card: %v", err)
	}

	if updated.FSRSReps != 1 {
		t.Errorf("Expected persisted reps to be 1, got %d", updated.FSRSReps)
	}
}

func TestReviewOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create deck and card
	deck := &Deck{Name: "Review Test Deck"}
	err := db.CreateDeck(deck)
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}

	card := &Card{
		DeckID:  deck.ID,
		CardKey: "review-test-card",
		Title:   "Review Test Card",
		Command: "echo 'Review test'",
	}
	err = db.CreateCard(card)
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	// Create a review
	now := time.Now()
	review := &Review{
		CardID:               card.ID,
		Rating:               int(fsrs.Good),
		ExecutionSuccess:     true,
		ExitCode:             func() *int { code := 0; return &code }(),
		Stdout:               "Review test\n",
		Stderr:               "",
		ThinkingTimeMs:       func() *int { ms := 5000; return &ms }(),
		ExecutionTimeMs:      func() *int { ms := 100; return &ms }(),
		TotalTimeMs:          func() *int { ms := 5100; return &ms }(),
		Attempts:             1,
		HelpAccessed:         false,
		FSRSDueBefore:        now,
		FSRSDueAfter:         now.Add(24 * time.Hour),
		FSRSStabilityBefore:  1.0,
		FSRSStabilityAfter:   2.5,
		FSRSDifficultyBefore: 5.0,
		FSRSDifficultyAfter:  4.8,
	}

	err = db.CreateReview(review)
	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	if review.ID == 0 {
		t.Error("Expected review ID to be set after creation")
	}
}

func TestAssetOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create deck
	deck := &Deck{Name: "Asset Test Deck"}
	err := db.CreateDeck(deck)
	if err != nil {
		t.Fatalf("Failed to create deck: %v", err)
	}

	// Test StoreAsset
	asset := &DeckAsset{
		DeckID:      deck.ID,
		Filename:    "test-config.json",
		Content:     []byte(`{"test": "value"}`),
		ContentType: "application/json",
	}

	err = db.StoreAsset(asset)
	if err != nil {
		t.Fatalf("Failed to store asset: %v", err)
	}

	// Test GetAsset
	retrieved, err := db.GetAsset(deck.ID, "test-config.json")
	if err != nil {
		t.Fatalf("Failed to get asset: %v", err)
	}

	if string(retrieved.Content) != string(asset.Content) {
		t.Errorf("Expected content %s, got %s", string(asset.Content), string(retrieved.Content))
	}

	// Test ListDeckAssets
	assets, err := db.ListDeckAssets(deck.ID)
	if err != nil {
		t.Fatalf("Failed to list deck assets: %v", err)
	}

	if len(assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(assets))
	}
}

func TestDatabaseMigration(t *testing.T) {
	// Test that migration runs successfully on a fresh database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "migration_test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database with migration: %v", err)
	}
	defer db.Close()

	// Verify tables were created by trying to insert data
	deck := &Deck{Name: "Migration Test Deck"}
	err = db.CreateDeck(deck)
	if err != nil {
		t.Fatalf("Failed to create deck after migration: %v", err)
	}
}
