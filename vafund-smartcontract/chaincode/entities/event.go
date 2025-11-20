package entities

import (
	"time"
)

type Event struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	IsActive    bool      `json:"isActive"`
	Timestamp   time.Time `json:"timestamp"`
	TxID        string    `json:"txId"`
}

func NewEvent(code, name, description string, startDate, endDate time.Time, isActive bool, timestamp time.Time, txID string) *Event {
	return &Event{
		Code:        code,
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    isActive,
		Timestamp:   timestamp,
		TxID:        txID,
	}
}

func (e *Event) IsValid() bool {
	return e.Code != "" && e.Name != "" && !e.EndDate.Before(e.StartDate)
}

func (e *Event) IsCurrentlyActive() bool {
	currentTime := time.Now()
	return e.IsActive && currentTime.After(e.StartDate) && currentTime.Before(e.EndDate)
}
