package entities

import (
	"time"
)

type Donation struct {
	ID         string    `json:"id"`
	SenderName string    `json:"senderName"`
	Amount     float64   `json:"amount"`
	Message    string    `json:"message,omitempty"`
	EventCode  string    `json:"eventCode,omitempty"` // Event that this donation is associated with
	Timestamp  time.Time `json:"timestamp"`
	TxID       string    `json:"txId"`
}

func NewDonation(id, senderName string, amount float64, message, eventCode string, timestamp time.Time, txID string) *Donation {
	return &Donation{
		ID:         id,
		SenderName: senderName,
		Amount:     amount,
		Message:    message,
		EventCode:  eventCode,
		Timestamp:  timestamp,
		TxID:       txID,
	}
}

func (d *Donation) IsValid() bool {
	return d.SenderName != "" && d.Amount > 0
}
