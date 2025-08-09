package review

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/open-spaced-repetition/go-fsrs/v3"

	"github.com/justinlyon12/ancli/internal/domain"
	"github.com/justinlyon12/ancli/internal/sandbox"
	"github.com/justinlyon12/ancli/internal/scheduler"
	"github.com/justinlyon12/ancli/internal/storage"
)

// Service implements ReviewService using storage, scheduler, and sandbox adapters
type Service struct {
	storage   storage.Storage
	scheduler *scheduler.Scheduler
	sandbox   sandbox.Sandbox
	sessions  map[string]*sessionState // In-memory session tracking
}

// sessionState tracks the internal state of a review session
type sessionState struct {
	*Session
	cardQueue []int // Card IDs in order
}

// NewService creates a new review service
func NewService(storage storage.Storage, scheduler *scheduler.Scheduler, sandbox sandbox.Sandbox) *Service {
	return &Service{
		storage:   storage,
		scheduler: scheduler,
		sandbox:   sandbox,
		sessions:  make(map[string]*sessionState),
	}
}

// StartSession begins a new review session
func (s *Service) StartSession(ctx context.Context, opts SessionOptions) (*Session, error) {
	sessionID := uuid.New().String()

	// Query cards based on options
	cards, err := s.queryCardsForSession(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards for session: %w", err)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("no cards available for review with the given options")
	}

	// Create card queue
	cardQueue := make([]int, len(cards))
	for i, card := range cards {
		cardQueue[i] = card.ID
	}

	// Shuffle if requested
	if opts.ShuffleCards {
		rand.Shuffle(len(cardQueue), func(i, j int) {
			cardQueue[i], cardQueue[j] = cardQueue[j], cardQueue[i]
		})
	}

	// Limit cards if maxCards is set
	if opts.MaxCards > 0 && len(cardQueue) > opts.MaxCards {
		cardQueue = cardQueue[:opts.MaxCards]
	}

	// Create session
	session := &Session{
		ID:             sessionID,
		StartedAt:      time.Now(),
		DeckID:         opts.DeckID,
		Options:        opts,
		CardsReviewed:  0,
		CardsRemaining: len(cardQueue),
		CurrentCardID:  nil,
	}

	// Store session state
	s.sessions[sessionID] = &sessionState{
		Session:   session,
		cardQueue: cardQueue,
	}

	return session, nil
}

// GetNextCard retrieves the next card due for review in the session
func (s *Service) GetNextCard(ctx context.Context, sessionID string) (*ReviewCard, error) {
	state, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if len(state.cardQueue) == 0 {
		return nil, fmt.Errorf("no more cards remaining in session")
	}

	// Get next card ID from queue
	cardID := state.cardQueue[0]

	// Get card from storage
	storageCard, err := s.storage.GetCard(cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card %d: %w", cardID, err)
	}

	// Convert to ReviewCard (reuse existing logic, add conversion method to storage)
	reviewCard, err := s.convertToReviewCard(ctx, storageCard)
	if err != nil {
		return nil, fmt.Errorf("failed to convert card: %w", err)
	}

	// Update session state
	state.CurrentCardID = &cardID

	return reviewCard, nil
}

// SubmitReview processes a card review and updates FSRS state
func (s *Service) SubmitReview(ctx context.Context, sessionID string, cardID int, rating domain.Rating, executionResult *domain.ExecutionResult) error {
	state, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Get the card from storage
	card, err := s.storage.GetCard(cardID)
	if err != nil {
		return fmt.Errorf("failed to get card: %w", err)
	}

	// Convert to FSRS card using existing method
	fsrsCard := card.ToFSRSCard()

	// Convert domain.Rating to fsrs.Rating
	// TODO: Refactor when domain.Rating replaces local Rating types
	var fsrsRating fsrs.Rating
	switch rating {
	case domain.Again:
		fsrsRating = fsrs.Again
	case domain.Hard:
		fsrsRating = fsrs.Hard
	case domain.Good:
		fsrsRating = fsrs.Good
	case domain.Easy:
		fsrsRating = fsrs.Easy
	default:
		return fmt.Errorf("invalid rating: %d", rating)
	}

	// Schedule the next review
	scheduleInfo := s.scheduler.ReviewCard(fsrsCard, fsrsRating)

	// Update card using existing method
	card.UpdateFromFSRSCard(scheduleInfo.Card)

	// Update card in storage
	if err := s.storage.UpdateCard(card); err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	// Record the review
	if err := s.createReviewRecord(ctx, cardID, rating, executionResult, fsrsCard, scheduleInfo.Card); err != nil {
		return fmt.Errorf("failed to create review record: %w", err)
	}

	// Update session state
	state.CardsReviewed++
	state.CardsRemaining--
	state.CurrentCardID = nil

	// Remove card from queue
	if len(state.cardQueue) > 0 && state.cardQueue[0] == cardID {
		state.cardQueue = state.cardQueue[1:]
	}

	return nil
}

// EndSession finalizes the review session and returns statistics
func (s *Service) EndSession(ctx context.Context, sessionID string) (*SessionStats, error) {
	state, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	duration := time.Since(state.StartedAt)

	// TODO: Calculate detailed stats from review records
	stats := &SessionStats{
		SessionID:     sessionID,
		Duration:      duration,
		CardsReviewed: state.CardsReviewed,
		NewCards:      0, // TODO: Calculate from review records
		ReviewCards:   0, // TODO: Calculate from review records
	}

	// Clean up session
	delete(s.sessions, sessionID)

	return stats, nil
}

// queryCardsForSession queries cards based on session options
func (s *Service) queryCardsForSession(ctx context.Context, opts SessionOptions) ([]*storage.Card, error) {
	var cards []*storage.Card
	var err error

	if opts.DeckID != nil {
		cards, err = s.storage.GetCardsByDeck(*opts.DeckID)
	} else {
		cards, err = s.storage.GetAllCards()
	}

	if err != nil {
		return nil, err
	}

	// Filter based on options
	var filtered []*storage.Card
	now := time.Now()

	for _, card := range cards {
		// Filter by new/review status
		if opts.NewCardsOnly && card.FSRSReps > 0 {
			continue
		}
		if opts.ReviewCardsOnly && card.FSRSReps == 0 {
			continue
		}

		// Check if card is due
		if card.FSRSReps > 0 && card.FSRSDue.After(now) {
			continue
		}

		filtered = append(filtered, card)
	}

	return filtered, nil
}

// convertToReviewCard converts storage.Card to ReviewCard with resolved configuration
func (s *Service) convertToReviewCard(ctx context.Context, storageCard *storage.Card) (*ReviewCard, error) {
	// Get deck for defaults
	deck, err := s.storage.GetDeck(storageCard.DeckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck: %w", err)
	}

	// Parse JSON fields - TODO: Add parsing methods to storage.Card
	envVars := make(map[string]string)
	if storageCard.EnvironmentVars != "" {
		_ = json.Unmarshal([]byte(storageCard.EnvironmentVars), &envVars)
	}

	var tags []string
	if storageCard.Tags != "" {
		_ = json.Unmarshal([]byte(storageCard.Tags), &tags)
	}

	var capabilities []string
	capStr := deck.DefaultCapabilities
	if storageCard.Capabilities != nil {
		capStr = *storageCard.Capabilities
	}
	if capStr != "" {
		_ = json.Unmarshal([]byte(capStr), &capabilities)
	}

	// Resolve configuration
	image := deck.DefaultImage
	if storageCard.Image != nil {
		image = *storageCard.Image
	}

	timeout := time.Duration(deck.DefaultTimeout) * time.Second
	if storageCard.Timeout != nil {
		timeout = time.Duration(*storageCard.Timeout) * time.Second
	}

	networkEnabled := deck.DefaultNetworkEnabled
	if storageCard.NetworkEnabled != nil {
		networkEnabled = *storageCard.NetworkEnabled
	}

	return &ReviewCard{
		ID:              storageCard.ID,
		DeckID:          storageCard.DeckID,
		CardKey:         storageCard.CardKey,
		Title:           storageCard.Title,
		Description:     storageCard.Description,
		Command:         storageCard.Command,
		WorkingDir:      storageCard.WorkingDir,
		EnvironmentVars: envVars,
		Image:           image,
		Timeout:         timeout,
		NetworkEnabled:  networkEnabled,
		Capabilities:    capabilities,
		DifficultyLevel: storageCard.DifficultyLevel,
		Tags:            tags,
		DueAt:           storageCard.FSRSDue,
		Stability:       storageCard.FSRSStability,
		Difficulty:      storageCard.FSRSDifficulty,
		ElapsedDays:     storageCard.FSRSElapsedDays,
		ScheduledDays:   storageCard.FSRSScheduledDays,
		Reps:            storageCard.FSRSReps,
		Lapses:          storageCard.FSRSLapses,
		State:           domain.CardState(storageCard.FSRSState),
		LastReview:      storageCard.FSRSLastReview,
	}, nil
}

// createReviewRecord creates a review record
func (s *Service) createReviewRecord(ctx context.Context, cardID int, rating domain.Rating,
	executionResult *domain.ExecutionResult, fsrsCardBefore, fsrsCardAfter fsrs.Card) error {

	review := &storage.Review{
		CardID:               cardID,
		ReviewedAt:           time.Now(),
		Rating:               int(rating),
		FSRSDueBefore:        fsrsCardBefore.Due,
		FSRSDueAfter:         fsrsCardAfter.Due,
		FSRSStabilityBefore:  fsrsCardBefore.Stability,
		FSRSStabilityAfter:   fsrsCardAfter.Stability,
		FSRSDifficultyBefore: fsrsCardBefore.Difficulty,
		FSRSDifficultyAfter:  fsrsCardAfter.Difficulty,
	}

	if executionResult != nil {
		review.ExecutionSuccess = executionResult.Success
		review.ExitCode = &executionResult.ExitCode
		review.Stdout = executionResult.Stdout
		review.Stderr = executionResult.Stderr

		if executionResult.Duration > 0 {
			ms := int(executionResult.Duration.Nanoseconds() / 1000000)
			review.ExecutionTimeMs = &ms
		}
		if executionResult.ThinkingTime > 0 {
			ms := int(executionResult.ThinkingTime.Nanoseconds() / 1000000)
			review.ThinkingTimeMs = &ms
		}
	}

	return s.storage.CreateReview(review)
}
