package entities

import (
	"time"
)

// Withdrawal represents a withdrawal record in the VaFund system
type Withdrawal struct {
	ID         string    `json:"id"`
	EventCode  string    `json:"eventCode"`
	Amount     float64   `json:"amount"`
	WithdrawBy string    `json:"withdrawBy"`
	DateTime   time.Time `json:"dateTime"`
	Timestamp  time.Time `json:"timestamp"`
	TxID       string    `json:"txId"`
}

// NewWithdrawal creates a new Withdrawal instance
func NewWithdrawal(id, eventCode string, amount float64, withdrawBy string, dateTime, timestamp time.Time, txID string) *Withdrawal {
	return &Withdrawal{
		ID:         id,
		EventCode:  eventCode,
		Amount:     amount,
		WithdrawBy: withdrawBy,
		DateTime:   dateTime,
		Timestamp:  timestamp,
		TxID:       txID,
	}
}

// IsValid validates the withdrawal data
func (w *Withdrawal) IsValid() bool {
	return w.EventCode != "" && w.Amount > 0 && w.WithdrawBy != ""
}
