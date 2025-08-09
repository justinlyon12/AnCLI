package review

import (
	"context"
	"testing"
	"time"

	"github.com/justinlyon12/ancli/internal/domain"
	"github.com/justinlyon12/ancli/internal/sandbox"
	"github.com/justinlyon12/ancli/internal/scheduler"
	"github.com/justinlyon12/ancli/internal/storage"
)

// mockDB implements a simple mock for testing
type mockDB struct {
	decks   map[int]*storage.Deck
	cards   map[int]*storage.Card
	reviews []storage.Review
}

func newMockDB() *mockDB {
	return &mockDB{
		decks:   make(map[int]*storage.Deck),
		cards:   make(map[int]*storage.Card),
		reviews: make([]storage.Review, 0),
	}
}

func (m *mockDB) Close() error { return nil }

func (m *mockDB) GetDeck(id int) (*storage.Deck, error) {
	if deck, exists := m.decks[id]; exists {
		return deck, nil
	}
	return nil, &NotFoundError{Resource: "deck", ID: id}
}

func (m *mockDB) GetCard(id int) (*storage.Card, error) {
	if card, exists := m.cards[id]; exists {
		return card, nil
	}
	return nil, &NotFoundError{Resource: "card", ID: id}
}

func (m *mockDB) UpdateCard(card *storage.Card) error {
	m.cards[card.ID] = card
	return nil
}

func (m *mockDB) CreateReview(review *storage.Review) error {
	review.ID = len(m.reviews) + 1
	m.reviews = append(m.reviews, *review)
	return nil
}

func (m *mockDB) GetCardsByDeck(deckID int) ([]*storage.Card, error) {
	var cards []*storage.Card
	for _, card := range m.cards {
		if card.DeckID == deckID {
			cards = append(cards, card)
		}
	}
	return cards, nil
}

func (m *mockDB) GetAllCards() ([]*storage.Card, error) {
	var cards []*storage.Card
	for _, card := range m.cards {
		cards = append(cards, card)
	}
	return cards, nil
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return "not found"
}

// mockSandbox implements the sandbox interface for testing
type mockSandbox struct {
	results map[string]*sandbox.ExecutionResult
}

func newMockSandbox() *mockSandbox {
	return &mockSandbox{
		results: make(map[string]*sandbox.ExecutionResult),
	}
}

func (m *mockSandbox) Run(ctx context.Context, config sandbox.ExecutionConfig) (*sandbox.ExecutionResult, error) {
	// Return a mock result based on the command
	cmdStr := config.Command[0]
	if result, exists := m.results[cmdStr]; exists {
		return result, nil
	}

	// Default successful result
	return &sandbox.ExecutionResult{
		ExitCode:      0,
		Success:       true,
		Stdout:        "mock output",
		Stderr:        "",
		StartedAt:     time.Now(),
		Duration:      100 * time.Millisecond,
		ContainerID:   "mock-container",
		ImageUsed:     config.Image,
		CorrelationID: config.CorrelationID,
	}, nil
}

func (m *mockSandbox) Cleanup(ctx context.Context) error {
	return nil
}

func (m *mockSandbox) Name() string {
	return "mock"
}

func TestNewService(t *testing.T) {
	db := newMockDB()
	sched := scheduler.NewScheduler()
	sb := newMockSandbox()

	service := NewService(db, sched, sb)
	if service == nil {
		t.Error("service should not be nil")
		return
	}

	if len(service.sessions) != 0 {
		t.Error("new service should have empty sessions")
	}
}

func TestStartSession_Success(t *testing.T) {
	db := newMockDB()
	sched := scheduler.NewScheduler()
	sb := newMockSandbox()

	// Add test data
	deck := &storage.Deck{
		ID:                    1,
		Name:                  "Test Deck",
		DefaultImage:          "alpine:3.18",
		DefaultTimeout:        30,
		DefaultNetworkEnabled: false,
		DefaultCapabilities:   "[]",
	}
	db.decks[1] = deck

	card := &storage.Card{
		ID:          1,
		DeckID:      1,
		CardKey:     "test-card",
		Title:       "Test Card",
		Description: "A test card",
		Command:     "echo hello",
		WorkingDir:  "/tmp",
		FSRSDue:     time.Now().Add(-1 * time.Hour), // Due for review
		FSRSReps:    0,                              // New card
		FSRSState:   0,
	}
	db.cards[1] = card

	service := NewService(db, sched, sb)
	ctx := context.Background()

	opts := SessionOptions{
		MaxCards:     10,
		ShuffleCards: false,
	}

	session, err := service.StartSession(ctx, opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if session == nil {
		t.Error("session should not be nil")
		return
	}

	if session.ID == "" {
		t.Error("session ID should not be empty")
	}

	if session.CardsRemaining != 1 {
		t.Errorf("expected 1 card remaining, got %d", session.CardsRemaining)
	}

	if session.CardsReviewed != 0 {
		t.Errorf("expected 0 cards reviewed, got %d", session.CardsReviewed)
	}
}

func TestStartSession_NoCards(t *testing.T) {
	db := newMockDB()
	sched := scheduler.NewScheduler()
	sb := newMockSandbox()
	service := NewService(db, sched, sb)
	ctx := context.Background()

	opts := SessionOptions{
		MaxCards: 10,
	}

	_, err := service.StartSession(ctx, opts)
	if err == nil {
		t.Error("expected error when no cards available")
	}
}

func TestSubmitReview_Success(t *testing.T) {
	db := newMockDB()
	sched := scheduler.NewScheduler()
	sb := newMockSandbox()

	// Add test data
	deck := &storage.Deck{
		ID:                    1,
		Name:                  "Test Deck",
		DefaultImage:          "alpine:3.18",
		DefaultTimeout:        30,
		DefaultNetworkEnabled: false,
		DefaultCapabilities:   "[]",
	}
	db.decks[1] = deck

	card := &storage.Card{
		ID:        1,
		DeckID:    1,
		CardKey:   "test-card",
		Title:     "Test Card",
		Command:   "echo hello",
		FSRSDue:   time.Now().Add(-1 * time.Hour),
		FSRSReps:  0, // New card
		FSRSState: 0,
	}
	db.cards[1] = card

	service := NewService(db, sched, sb)
	ctx := context.Background()

	// Start a session
	session, err := service.StartSession(ctx, SessionOptions{MaxCards: 10})
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// Submit a review
	executionResult := &domain.ExecutionResult{
		Success:      true,
		ExitCode:     0,
		Stdout:       "hello",
		Duration:     100 * time.Millisecond,
		ThinkingTime: 2 * time.Second,
	}

	err = service.SubmitReview(ctx, session.ID, 1, domain.Good, executionResult)
	if err != nil {
		t.Errorf("unexpected error submitting review: %v", err)
		return
	}

	// Check that card was updated
	updatedCard := db.cards[1]
	if updatedCard.FSRSReps != 1 {
		t.Errorf("expected card reps to be 1, got %d", updatedCard.FSRSReps)
	}

	// Check that review was recorded
	if len(db.reviews) != 1 {
		t.Errorf("expected 1 review recorded, got %d", len(db.reviews))
		return
	}

	review := db.reviews[0]
	if review.Rating != int(domain.Good) {
		t.Errorf("expected rating %d, got %d", int(domain.Good), review.Rating)
	}

	if !review.ExecutionSuccess {
		t.Error("expected execution success to be true")
	}
}

func TestEndSession_Success(t *testing.T) {
	db := newMockDB()
	sched := scheduler.NewScheduler()
	sb := newMockSandbox()
	service := NewService(db, sched, sb)
	ctx := context.Background()

	// Add minimal test data
	deck := &storage.Deck{
		ID: 1, Name: "Test", DefaultImage: "alpine:3.18", DefaultTimeout: 30,
		DefaultCapabilities: "[]", DefaultNetworkEnabled: false,
	}
	db.decks[1] = deck

	card := &storage.Card{
		ID: 1, DeckID: 1, Title: "Test", Command: "echo test",
		FSRSDue: time.Now().Add(-1 * time.Hour), FSRSReps: 0, FSRSState: 0,
	}
	db.cards[1] = card

	// Start a session
	session, err := service.StartSession(ctx, SessionOptions{MaxCards: 10})
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// End the session
	stats, err := service.EndSession(ctx, session.ID)
	if err != nil {
		t.Errorf("unexpected error ending session: %v", err)
		return
	}

	if stats == nil {
		t.Error("session stats should not be nil")
		return
	}

	if stats.SessionID != session.ID {
		t.Errorf("expected session ID %s, got %s", session.ID, stats.SessionID)
	}

	// Verify session was cleaned up
	if len(service.sessions) != 0 {
		t.Error("session should be cleaned up after ending")
	}
}

func TestDomainRatingParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.Rating
		wantErr  bool
	}{
		{"1", domain.Again, false},
		{"again", domain.Again, false},
		{"A", domain.Again, false},
		{"2", domain.Hard, false},
		{"hard", domain.Hard, false},
		{"3", domain.Good, false},
		{"good", domain.Good, false},
		{"4", domain.Easy, false},
		{"easy", domain.Easy, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			rating, err := domain.ParseRating(tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && rating != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, rating)
			}
		})
	}
}
