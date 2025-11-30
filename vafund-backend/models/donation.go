package models

import (
	"time"
)

// Donation represents a donation record
type Donation struct {
	ID         string    `json:"id" example:"donation123"`
	SenderName string    `json:"senderName" example:"John Doe"`
	Amount     float64   `json:"amount" example:"100000"`
	Message    string    `json:"message,omitempty" example:"Semoga bermanfaat"`
	EventCode  string    `json:"eventCode,omitempty" example:"RAMADAN2025"`
	Timestamp  time.Time `json:"timestamp" example:"2025-11-10T10:30:00Z"`
	TxID       string    `json:"txId" example:"tx123abc"`
}

// CreateDonationRequest represents the request payload for creating a donation
type CreateDonationRequest struct {
	DonationID string `json:"donationId" validate:"required" example:"donation123"`
	SenderName string `json:"senderName" validate:"required" example:"John Doe"`
	Amount     string `json:"amount" validate:"required" example:"100000"`
	Message    string `json:"message" example:"Semoga bermanfaat"`
	EventCode  string `json:"eventCode" example:"RAMADAN2025"`
}

// DonationResponse represents the response for donation operations
type DonationResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Donation created successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// TotalDonationResponse represents the response for total donation amount
type TotalDonationResponse struct {
	Success       bool    `json:"success" example:"true"`
	Message       string  `json:"message" example:"Total donations retrieved successfully"`
	TotalAmount   float64 `json:"totalAmount" example:"350000"`
	CurrentAmount float64 `json:"currentAmount,omitempty" example:"300000"`
	TotalCount    int     `json:"totalCount" example:"3"`
}

// Event represents an event record
type Event struct {
	Code        string    `json:"code" example:"RAMADAN2025"`
	Name        string    `json:"name" example:"Ramadan Charity Drive"`
	Description string    `json:"description" example:"Annual charity drive during Ramadan month"`
	ImgUrl      string    `json:"imgUrl" example:"/api/events/RAMADAN2025/image"`
	StartDate   time.Time `json:"startDate" example:"2025-03-01T00:00:00Z"`
	EndDate     time.Time `json:"endDate" example:"2025-04-30T23:59:59Z"`
	IsActive    bool      `json:"isActive" example:"true"`
	Timestamp   time.Time `json:"timestamp" example:"2025-11-17T10:30:00Z"`
	TxID        string    `json:"txId" example:"event-tx123"`
}

// CreateEventRequest represents the request payload for creating an event
type CreateEventRequest struct {
	Code        string `json:"code" validate:"required" example:"RAMADAN2025"`
	Name        string `json:"name" validate:"required" example:"Ramadan Charity Drive"`
	Description string `json:"description" validate:"required" example:"Annual charity drive during Ramadan month"`
	StartDate   string `json:"startDate" validate:"required" example:"2025-03-01T00:00:00Z"`
	EndDate     string `json:"endDate" validate:"required" example:"2025-04-30T23:59:59Z"`
	IsActive    string `json:"isActive" validate:"required" example:"true"`
}

// UpdateEventRequest represents the request payload for updating an event
type UpdateEventRequest struct {
	IsActive string `json:"isActive" validate:"required" example:"false"`
}

// UpdateEventDetailRequest represents the request payload for updating event details
type UpdateEventDetailRequest struct {
	Name        string `json:"name" validate:"required" example:"Updated Ramadan Charity Drive"`
	Description string `json:"description" validate:"required" example:"Updated description for annual charity drive"`
	StartDate   string `json:"startDate" validate:"required" example:"2025-03-01T00:00:00Z"`
	EndDate     string `json:"endDate" validate:"required" example:"2025-04-30T23:59:59Z"`
	IsActive    string `json:"isActive" validate:"required" example:"true"`
}

// EventResponse represents the response for event operations
type EventResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Event created successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Error message"`
	Error   string `json:"error,omitempty" example:"Detailed error description"`
}
