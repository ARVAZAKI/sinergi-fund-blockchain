package entities

import (
	"time"
)

// Gallery represents a gallery image for an event
type Gallery struct {
	ID          string    `json:"id"`
	EventCode   string    `json:"event_code"`
	ImageURL    string    `json:"image_url"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	TxID        string    `json:"tx_id"`
}

// NewGallery creates a new Gallery instance
func NewGallery(id string, eventCode string, imageURL string, description string, timestamp time.Time, txID string) *Gallery {
	return &Gallery{
		ID:          id,
		EventCode:   eventCode,
		ImageURL:    imageURL,
		Description: description,
		Timestamp:   timestamp,
		TxID:        txID,
	}
}

// IsValid validates the gallery data
func (g *Gallery) IsValid() bool {
	if g.ID == "" {
		return false
	}
	if g.EventCode == "" {
		return false
	}
	if g.ImageURL == "" {
		return false
	}
	return true
}
