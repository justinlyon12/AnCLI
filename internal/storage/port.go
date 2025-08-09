package storage

// Storage defines the interface for persistent data operations
type Storage interface {
	// Deck operations
	GetDeck(id int) (*Deck, error)

	// Card operations
	GetCard(id int) (*Card, error)
	UpdateCard(card *Card) error
	GetCardsByDeck(deckID int) ([]*Card, error)
	GetAllCards() ([]*Card, error)

	// Review operations
	CreateReview(review *Review) error

	// Lifecycle
	Close() error
}
