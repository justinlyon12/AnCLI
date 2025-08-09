package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
	path string
}

// NewDB creates a new database connection and runs migrations
func NewDB(dbPath string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite connection
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure SQLite connection
	conn.SetMaxOpenConns(1) // SQLite works best with single connection
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

	// Enable foreign keys and WAL mode for better performance
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if _, err := conn.Exec("PRAGMA journal_mode = WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	db := &DB{
		conn: conn,
		path: dbPath,
	}

	// Auto-migrate database schema
	if err := MigrateDatabase(db); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// CreateDeck creates a new deck
func (db *DB) CreateDeck(deck *Deck) error {
	query := `
		INSERT INTO decks (name, description, version, author, default_image, default_timeout, 
			default_network_enabled, default_capabilities, fsrs_parameters)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		deck.Name, deck.Description, deck.Version, deck.Author,
		deck.DefaultImage, deck.DefaultTimeout, deck.DefaultNetworkEnabled,
		deck.DefaultCapabilities, deck.FSRSParameters,
	)
	if err != nil {
		return fmt.Errorf("failed to create deck: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get deck ID: %w", err)
	}

	deck.ID = int(id)
	deck.CreatedAt = time.Now()
	deck.UpdatedAt = time.Now()

	return nil
}

// GetDeck retrieves a deck by ID
func (db *DB) GetDeck(id int) (*Deck, error) {
	query := `
		SELECT id, name, description, version, author, created_at, updated_at,
			default_image, default_timeout, default_network_enabled, 
			default_capabilities, fsrs_parameters
		FROM decks WHERE id = ?
	`

	deck := &Deck{}
	err := db.conn.QueryRow(query, id).Scan(
		&deck.ID, &deck.Name, &deck.Description, &deck.Version, &deck.Author,
		&deck.CreatedAt, &deck.UpdatedAt, &deck.DefaultImage, &deck.DefaultTimeout,
		&deck.DefaultNetworkEnabled, &deck.DefaultCapabilities, &deck.FSRSParameters,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deck not found")
		}
		return nil, fmt.Errorf("failed to get deck: %w", err)
	}

	return deck, nil
}

// ListDecks retrieves all decks
func (db *DB) ListDecks() ([]*Deck, error) {
	query := `
		SELECT id, name, description, version, author, created_at, updated_at,
			default_image, default_timeout, default_network_enabled, 
			default_capabilities, fsrs_parameters
		FROM decks ORDER BY name
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list decks: %w", err)
	}
	defer rows.Close()

	var decks []*Deck
	for rows.Next() {
		deck := &Deck{}
		err := rows.Scan(
			&deck.ID, &deck.Name, &deck.Description, &deck.Version, &deck.Author,
			&deck.CreatedAt, &deck.UpdatedAt, &deck.DefaultImage, &deck.DefaultTimeout,
			&deck.DefaultNetworkEnabled, &deck.DefaultCapabilities, &deck.FSRSParameters,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck: %w", err)
		}
		decks = append(decks, deck)
	}

	return decks, nil
}

// CreateCard creates a new card with initial FSRS state
func (db *DB) CreateCard(card *Card) error {
	// Initialize FSRS state for new card
	fsrsCard := fsrs.NewCard()
	card.FSRSDue = fsrsCard.Due
	card.FSRSStability = fsrsCard.Stability
	card.FSRSDifficulty = fsrsCard.Difficulty
	card.FSRSElapsedDays = int(fsrsCard.ElapsedDays)
	card.FSRSScheduledDays = int(fsrsCard.ScheduledDays)
	card.FSRSReps = int(fsrsCard.Reps)
	card.FSRSLapses = int(fsrsCard.Lapses)
	card.FSRSState = int(fsrsCard.State)
	card.FSRSLastReview = nil

	query := `
		INSERT INTO cards (deck_id, card_key, title, description, command, working_dir,
			environment_vars, image, timeout, network_enabled, capabilities,
			difficulty_level, tags, prerequisites, prerequisite_mode,
			fsrs_due, fsrs_stability, fsrs_difficulty, fsrs_elapsed_days,
			fsrs_scheduled_days, fsrs_reps, fsrs_lapses, fsrs_state, fsrs_last_review)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		card.DeckID, card.CardKey, card.Title, card.Description, card.Command,
		card.WorkingDir, card.EnvironmentVars, card.Image, card.Timeout,
		card.NetworkEnabled, card.Capabilities, card.DifficultyLevel, card.Tags,
		card.Prerequisites, card.PrerequisiteMode, card.FSRSDue, card.FSRSStability,
		card.FSRSDifficulty, card.FSRSElapsedDays, card.FSRSScheduledDays,
		card.FSRSReps, card.FSRSLapses, card.FSRSState, card.FSRSLastReview,
	)
	if err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get card ID: %w", err)
	}

	card.ID = int(id)
	card.CreatedAt = time.Now()
	card.UpdatedAt = time.Now()

	return nil
}

// GetCard retrieves a card by ID
func (db *DB) GetCard(id int) (*Card, error) {
	query := `
		SELECT id, deck_id, card_key, title, description, command, working_dir,
			environment_vars, image, timeout, network_enabled, capabilities,
			difficulty_level, tags, prerequisites, prerequisite_mode,
			fsrs_due, fsrs_stability, fsrs_difficulty, fsrs_elapsed_days,
			fsrs_scheduled_days, fsrs_reps, fsrs_lapses, fsrs_state, fsrs_last_review,
			created_at, updated_at
		FROM cards WHERE id = ?
	`

	card := &Card{}
	err := db.conn.QueryRow(query, id).Scan(
		&card.ID, &card.DeckID, &card.CardKey, &card.Title, &card.Description,
		&card.Command, &card.WorkingDir, &card.EnvironmentVars, &card.Image,
		&card.Timeout, &card.NetworkEnabled, &card.Capabilities, &card.DifficultyLevel,
		&card.Tags, &card.Prerequisites, &card.PrerequisiteMode, &card.FSRSDue,
		&card.FSRSStability, &card.FSRSDifficulty, &card.FSRSElapsedDays,
		&card.FSRSScheduledDays, &card.FSRSReps, &card.FSRSLapses, &card.FSRSState,
		&card.FSRSLastReview, &card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found")
		}
		return nil, fmt.Errorf("failed to get card: %w", err)
	}

	return card, nil
}

// GetDueCards retrieves all cards that are due for review
func (db *DB) GetDueCards() ([]*Card, error) {
	query := `
		SELECT id, deck_id, card_key, title, description, command, working_dir,
			environment_vars, image, timeout, network_enabled, capabilities,
			difficulty_level, tags, prerequisites, prerequisite_mode,
			fsrs_due, fsrs_stability, fsrs_difficulty, fsrs_elapsed_days,
			fsrs_scheduled_days, fsrs_reps, fsrs_lapses, fsrs_state, fsrs_last_review,
			created_at, updated_at
		FROM cards 
		WHERE fsrs_due <= datetime('now')
		ORDER BY fsrs_due ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get due cards: %w", err)
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card := &Card{}
		err := rows.Scan(
			&card.ID, &card.DeckID, &card.CardKey, &card.Title, &card.Description,
			&card.Command, &card.WorkingDir, &card.EnvironmentVars, &card.Image,
			&card.Timeout, &card.NetworkEnabled, &card.Capabilities, &card.DifficultyLevel,
			&card.Tags, &card.Prerequisites, &card.PrerequisiteMode, &card.FSRSDue,
			&card.FSRSStability, &card.FSRSDifficulty, &card.FSRSElapsedDays,
			&card.FSRSScheduledDays, &card.FSRSReps, &card.FSRSLapses, &card.FSRSState,
			&card.FSRSLastReview, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// UpdateCard updates a card's full state
func (db *DB) UpdateCard(card *Card) error {
	query := `
		UPDATE cards SET 
			title = ?, description = ?, command = ?, working_dir = ?,
			environment_vars = ?, image = ?, timeout = ?, network_enabled = ?,
			capabilities = ?, difficulty_level = ?, tags = ?, prerequisites = ?,
			prerequisite_mode = ?, fsrs_due = ?, fsrs_stability = ?, fsrs_difficulty = ?,
			fsrs_elapsed_days = ?, fsrs_scheduled_days = ?, fsrs_reps = ?,
			fsrs_lapses = ?, fsrs_state = ?, fsrs_last_review = ?,
			updated_at = datetime('now')
		WHERE id = ?
	`

	_, err := db.conn.Exec(query,
		card.Title, card.Description, card.Command, card.WorkingDir,
		card.EnvironmentVars, card.Image, card.Timeout, card.NetworkEnabled,
		card.Capabilities, card.DifficultyLevel, card.Tags, card.Prerequisites,
		card.PrerequisiteMode, card.FSRSDue, card.FSRSStability, card.FSRSDifficulty,
		card.FSRSElapsedDays, card.FSRSScheduledDays, card.FSRSReps,
		card.FSRSLapses, card.FSRSState, card.FSRSLastReview, card.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	return nil
}

// UpdateCardFSRS updates a card's FSRS state after review
func (db *DB) UpdateCardFSRS(card *Card) error {
	query := `
		UPDATE cards SET 
			fsrs_due = ?, fsrs_stability = ?, fsrs_difficulty = ?,
			fsrs_elapsed_days = ?, fsrs_scheduled_days = ?, fsrs_reps = ?,
			fsrs_lapses = ?, fsrs_state = ?, fsrs_last_review = ?,
			updated_at = datetime('now')
		WHERE id = ?
	`

	_, err := db.conn.Exec(query,
		card.FSRSDue, card.FSRSStability, card.FSRSDifficulty,
		card.FSRSElapsedDays, card.FSRSScheduledDays, card.FSRSReps,
		card.FSRSLapses, card.FSRSState, card.FSRSLastReview, card.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update card FSRS state: %w", err)
	}

	return nil
}

// GetCardsByDeck retrieves all cards for a specific deck
func (db *DB) GetCardsByDeck(deckID int) ([]*Card, error) {
	query := `
		SELECT id, deck_id, card_key, title, description, command, working_dir,
			environment_vars, image, timeout, network_enabled, capabilities,
			difficulty_level, tags, prerequisites, prerequisite_mode,
			fsrs_due, fsrs_stability, fsrs_difficulty, fsrs_elapsed_days,
			fsrs_scheduled_days, fsrs_reps, fsrs_lapses, fsrs_state, fsrs_last_review,
			created_at, updated_at
		FROM cards WHERE deck_id = ?
		ORDER BY card_key
	`

	rows, err := db.conn.Query(query, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards by deck: %w", err)
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card := &Card{}
		err := rows.Scan(
			&card.ID, &card.DeckID, &card.CardKey, &card.Title, &card.Description,
			&card.Command, &card.WorkingDir, &card.EnvironmentVars, &card.Image,
			&card.Timeout, &card.NetworkEnabled, &card.Capabilities, &card.DifficultyLevel,
			&card.Tags, &card.Prerequisites, &card.PrerequisiteMode, &card.FSRSDue,
			&card.FSRSStability, &card.FSRSDifficulty, &card.FSRSElapsedDays,
			&card.FSRSScheduledDays, &card.FSRSReps, &card.FSRSLapses, &card.FSRSState,
			&card.FSRSLastReview, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// GetAllCards retrieves all cards
func (db *DB) GetAllCards() ([]*Card, error) {
	query := `
		SELECT id, deck_id, card_key, title, description, command, working_dir,
			environment_vars, image, timeout, network_enabled, capabilities,
			difficulty_level, tags, prerequisites, prerequisite_mode,
			fsrs_due, fsrs_stability, fsrs_difficulty, fsrs_elapsed_days,
			fsrs_scheduled_days, fsrs_reps, fsrs_lapses, fsrs_state, fsrs_last_review,
			created_at, updated_at
		FROM cards
		ORDER BY deck_id, card_key
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all cards: %w", err)
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card := &Card{}
		err := rows.Scan(
			&card.ID, &card.DeckID, &card.CardKey, &card.Title, &card.Description,
			&card.Command, &card.WorkingDir, &card.EnvironmentVars, &card.Image,
			&card.Timeout, &card.NetworkEnabled, &card.Capabilities, &card.DifficultyLevel,
			&card.Tags, &card.Prerequisites, &card.PrerequisiteMode, &card.FSRSDue,
			&card.FSRSStability, &card.FSRSDifficulty, &card.FSRSElapsedDays,
			&card.FSRSScheduledDays, &card.FSRSReps, &card.FSRSLapses, &card.FSRSState,
			&card.FSRSLastReview, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// CreateReview records a review session
func (db *DB) CreateReview(review *Review) error {
	query := `
		INSERT INTO reviews (card_id, rating, execution_success, exit_code, stdout, stderr,
			thinking_time_ms, execution_time_ms, total_time_ms, attempts, help_accessed,
			fsrs_due_before, fsrs_due_after, fsrs_stability_before, fsrs_stability_after,
			fsrs_difficulty_before, fsrs_difficulty_after)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		review.CardID, review.Rating, review.ExecutionSuccess, review.ExitCode,
		review.Stdout, review.Stderr, review.ThinkingTimeMs, review.ExecutionTimeMs,
		review.TotalTimeMs, review.Attempts, review.HelpAccessed,
		review.FSRSDueBefore, review.FSRSDueAfter, review.FSRSStabilityBefore,
		review.FSRSStabilityAfter, review.FSRSDifficultyBefore, review.FSRSDifficultyAfter,
	)
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get review ID: %w", err)
	}

	review.ID = int(id)
	review.ReviewedAt = time.Now()

	return nil
}

// StoreAsset stores a deck asset
func (db *DB) StoreAsset(asset *DeckAsset) error {
	query := `
		INSERT OR REPLACE INTO card_assets (deck_id, filename, content, content_type)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		asset.DeckID, asset.Filename, asset.Content, asset.ContentType,
	)
	if err != nil {
		return fmt.Errorf("failed to store asset: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get asset ID: %w", err)
	}

	asset.ID = int(id)
	asset.CreatedAt = time.Now()

	return nil
}

// GetAsset retrieves a deck asset by filename
func (db *DB) GetAsset(deckID int, filename string) (*DeckAsset, error) {
	query := `
		SELECT id, deck_id, filename, content, content_type, created_at
		FROM card_assets WHERE deck_id = ? AND filename = ?
	`

	asset := &DeckAsset{}
	err := db.conn.QueryRow(query, deckID, filename).Scan(
		&asset.ID, &asset.DeckID, &asset.Filename, &asset.Content,
		&asset.ContentType, &asset.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("asset not found")
		}
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return asset, nil
}

// ListDeckAssets retrieves all assets for a deck
func (db *DB) ListDeckAssets(deckID int) ([]*DeckAsset, error) {
	query := `
		SELECT id, deck_id, filename, content, content_type, created_at
		FROM card_assets WHERE deck_id = ? ORDER BY filename
	`

	rows, err := db.conn.Query(query, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to list deck assets: %w", err)
	}
	defer rows.Close()

	var assets []*DeckAsset
	for rows.Next() {
		asset := &DeckAsset{}
		err := rows.Scan(
			&asset.ID, &asset.DeckID, &asset.Filename, &asset.Content,
			&asset.ContentType, &asset.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, asset)
	}

	return assets, nil
}
