package models

import "time"

type Withdrawal struct {
	ID         string    `json:"id"`
	EventCode  string    `json:"event_code" validate:"required"`
	Amount     float64   `json:"amount" validate:"required,min=0.01"`
	WithdrawBy string    `json:"withdraw_by" validate:"required"`
	DateTime   time.Time `json:"date_time" validate:"required"`
	Timestamp  time.Time `json:"timestamp"`
	TxID       string    `json:"tx_id"`
}

type WithdrawalRequest struct {
	EventCode  string  `json:"event_code" validate:"required"`
	Amount     float64 `json:"amount" validate:"required,min=0.01"`
	WithdrawBy string  `json:"withdraw_by" validate:"required"`
}

type WithdrawalResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    *Withdrawal `json:"data,omitempty"`
}

type WithdrawalsResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    []*Withdrawal `json:"data,omitempty"`
}
