package models

import (
	"time"
)

// Gallery represents a gallery image record
type Gallery struct {
	ID          string    `json:"id" example:"gallery001"`
	EventCode   string    `json:"eventCode" example:"RAMADAN2025"`
	ImageURL    string    `json:"imageUrl" example:"/api/gallery/gallery001/image"`
	Description string    `json:"description" example:"Event photo documentation"`
	Timestamp   time.Time `json:"timestamp" example:"2025-11-20T10:30:00Z"`
	TxID        string    `json:"txId" example:"gallery-tx123"`
}

// CreateGalleryRequest represents the request payload for creating a gallery
type CreateGalleryRequest struct {
	ID          string `json:"id" validate:"required" example:"gallery001"`
	EventCode   string `json:"eventCode" validate:"required" example:"RAMADAN2025"`
	Description string `json:"description" example:"Event photo documentation"`
	// Image file will be uploaded via multipart/form-data
}

// UpdateGalleryRequest represents the request payload for updating a gallery
type UpdateGalleryRequest struct {
	EventCode   string `json:"eventCode" validate:"required" example:"RAMADAN2025"`
	Description string `json:"description" example:"Updated photo documentation"`
	// Image file is optional for update via multipart/form-data
}

// GalleryResponse represents the response for gallery operations
type GalleryResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Gallery created successfully"`
	Data    interface{} `json:"data,omitempty"`
}
