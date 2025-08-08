package storage

const createTablesSQL = `
-- Deck metadata and configuration
CREATE TABLE IF NOT EXISTS decks (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    version TEXT,
    author TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    -- Deck-level sandbox defaults
    default_image TEXT DEFAULT 'alpine:latest',
    default_timeout INTEGER DEFAULT 5, -- seconds
    default_network_enabled BOOLEAN DEFAULT FALSE,
    default_capabilities TEXT, -- JSON array of capabilities
    -- FSRS parameters for this deck (can be tuned per deck)
    fsrs_parameters TEXT -- JSON blob of FSRS parameters
);

-- Individual cards within decks
CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY,
    deck_id INTEGER NOT NULL,
    -- Card identification
    card_key TEXT NOT NULL, -- unique within deck, from cards.csv
    title TEXT NOT NULL,
    description TEXT,
    -- Command execution
    command TEXT NOT NULL,
    working_dir TEXT DEFAULT '/tmp',
    environment_vars TEXT, -- JSON object of env vars
    -- Sandbox overrides (NULL = use deck defaults)
    image TEXT,
    timeout INTEGER,
    network_enabled BOOLEAN,
    capabilities TEXT, -- JSON array
    -- Learning metadata
    difficulty_level INTEGER DEFAULT 1, -- 1-5 scale
    tags TEXT, -- JSON array of tags
    -- Prerequisites (symbolic linking approach)
    prerequisites TEXT, -- JSON array of card_keys
    prerequisite_mode TEXT DEFAULT 'link', -- 'enforce' or 'link'
    -- FSRS state
    fsrs_due DATETIME NOT NULL,
    fsrs_stability REAL NOT NULL,
    fsrs_difficulty REAL NOT NULL,
    fsrs_elapsed_days INTEGER NOT NULL DEFAULT 0,
    fsrs_scheduled_days INTEGER NOT NULL DEFAULT 0,
    fsrs_reps INTEGER NOT NULL DEFAULT 0,
    fsrs_lapses INTEGER NOT NULL DEFAULT 0,
    fsrs_state INTEGER NOT NULL DEFAULT 0, -- 0=New, 1=Learning, 2=Review, 3=Relearning
    fsrs_last_review DATETIME,
    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
    UNIQUE(deck_id, card_key)
);

-- Review history for analytics and FSRS optimization
CREATE TABLE IF NOT EXISTS reviews (
    id INTEGER PRIMARY KEY,
    card_id INTEGER NOT NULL,
    -- Review session
    reviewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    rating INTEGER NOT NULL, -- 1=Again, 2=Hard, 3=Good, 4=Easy
    -- Execution results
    execution_success BOOLEAN NOT NULL,
    exit_code INTEGER,
    stdout TEXT,
    stderr TEXT,
    -- Enhanced timing metrics
    thinking_time_ms INTEGER, -- time from card shown to command started
    execution_time_ms INTEGER, -- actual command execution time
    total_time_ms INTEGER, -- total time for the card
    -- Interaction metrics
    attempts INTEGER DEFAULT 1, -- if user retries command
    help_accessed BOOLEAN DEFAULT FALSE, -- if user viewed hints/docs
    -- FSRS state transitions
    fsrs_due_before DATETIME NOT NULL,
    fsrs_due_after DATETIME NOT NULL,
    fsrs_stability_before REAL NOT NULL,
    fsrs_stability_after REAL NOT NULL,
    fsrs_difficulty_before REAL NOT NULL,
    fsrs_difficulty_after REAL NOT NULL,
    FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE
);

-- Supporting files that cards within a deck can reference
CREATE TABLE IF NOT EXISTS card_assets (
    id INTEGER PRIMARY KEY,
    deck_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    content BLOB NOT NULL,
    content_type TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
    UNIQUE(deck_id, filename)
);

-- Deck version tracking for in-place updates
CREATE TABLE IF NOT EXISTS deck_versions (
    id INTEGER PRIMARY KEY,
    deck_id INTEGER NOT NULL,
    version TEXT NOT NULL,
    changes TEXT, -- JSON diff of what changed
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_cards_due ON cards(fsrs_due);
CREATE INDEX IF NOT EXISTS idx_cards_deck ON cards(deck_id);
CREATE INDEX IF NOT EXISTS idx_cards_prerequisites ON cards(prerequisites) WHERE prerequisites IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reviews_card ON reviews(card_id);
CREATE INDEX IF NOT EXISTS idx_reviews_date ON reviews(reviewed_at);
CREATE INDEX IF NOT EXISTS idx_assets_deck ON card_assets(deck_id);
`

// MigrateDatabase creates all tables and indexes
// This is called automatically on every database connection
// Safe to run multiple times due to IF NOT EXISTS clauses
func MigrateDatabase(db *DB) error {
	_, err := db.conn.Exec(createTablesSQL)
	if err != nil {
		return err
	}
	return nil
}
