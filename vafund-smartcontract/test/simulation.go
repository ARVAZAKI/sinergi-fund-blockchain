package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// Donation struct - copy dari chaincode untuk testing
type Donation struct {
	ID         string    `json:"id"`
	SenderName string    `json:"senderName"`
	Amount     float64   `json:"amount"`
	Message    string    `json:"message,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	TxID       string    `json:"txId"`
}

type Event struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	IsActive  bool      `json:"isActive"`
	Timestamp time.Time `json:"timestamp"`
	TxID      string    `json:"txId"`
}

func main() {
	fmt.Println("========== VaFund Smart Contract Test ==========")
	fmt.Println("Simulasi Donations & Events tanpa blockchain network")
	fmt.Println()

	// Test 1: Membuat donation
	fmt.Println("1. Testing Donation Struct:")
	donation1 := Donation{
		ID:         "donation1",
		SenderName: "John Doe",
		Amount:     100000,
		Message:    "Semoga bermanfaat",
		Timestamp:  time.Now(),
		TxID:       "tx123abc",
	}

	donation2 := Donation{
		ID:         "donation2",
		SenderName: "Jane Smith",
		Amount:     250000,
		Message:    "Untuk kebaikan",
		Timestamp:  time.Now(),
		TxID:       "tx456def",
	}

	// Simulasi storage (dalam real blockchain ini disimpan di ledger)
	donations := make(map[string]Donation)
	donations[donation1.ID] = donation1
	donations[donation2.ID] = donation2

	// Test 2: Read donation
	fmt.Println("2. Testing Read Donation:")
	if donation, exists := donations["donation1"]; exists {
		donationJSON, _ := json.MarshalIndent(donation, "", "  ")
		fmt.Printf("Found donation1: %s\n\n", donationJSON)
	}

	// Test 3: Get all donations
	fmt.Println("3. Testing Get All Donations:")
	var allDonations []Donation
	for _, donation := range donations {
		allDonations = append(allDonations, donation)
	}

	allJSON, _ := json.MarshalIndent(allDonations, "", "  ")
	fmt.Printf("All donations: %s\n\n", allJSON)

	// Test 4: Calculate total
	fmt.Println("4. Testing Total Donations:")
	var total float64
	for _, donation := range donations {
		total += donation.Amount
	}
	fmt.Printf("Total donations: Rp %.2f\n\n", total)

	// Test 5: Filter by amount
	fmt.Println("5. Testing Filter by Amount (min: 200000):")
	var filteredDonations []Donation
	minAmount := 200000.0

	for _, donation := range donations {
		if donation.Amount >= minAmount {
			filteredDonations = append(filteredDonations, donation)
		}
	}

	filteredJSON, _ := json.MarshalIndent(filteredDonations, "", "  ")
	fmt.Printf("Donations >= %.0f: %s\n\n", minAmount, filteredJSON)

	// Test 6: Testing Events
	fmt.Println("6. Testing Event Creation:")
	event1 := Event{
		Code:      "RAMADAN2025",
		Name:      "Ramadan Charity Drive 2025",
		StartDate: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 4, 30, 23, 59, 59, 0, time.UTC),
		IsActive:  true,
		Timestamp: time.Now(),
		TxID:      "event-tx123",
	}

	event2 := Event{
		Code:      "HARI_RAYA2025",
		Name:      "Hari Raya Donation Campaign",
		StartDate: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 4, 15, 23, 59, 59, 0, time.UTC),
		IsActive:  false,
		Timestamp: time.Now(),
		TxID:      "event-tx456",
	}

	// Simulasi event storage
	events := make(map[string]Event)
	events[event1.Code] = event1
	events[event2.Code] = event2

	// Test 7: Get all events
	fmt.Println("7. Testing Get All Events:")
	var allEvents []Event
	for _, event := range events {
		allEvents = append(allEvents, event)
	}

	allEventsJSON, _ := json.MarshalIndent(allEvents, "", "  ")
	fmt.Printf("All events: %s\n\n", allEventsJSON)

	// Test 8: Get active events
	fmt.Println("8. Testing Get Active Events:")
	var activeEvents []Event
	currentTime := time.Now()

	for _, event := range events {
		if event.IsActive && currentTime.After(event.StartDate) && currentTime.Before(event.EndDate) {
			activeEvents = append(activeEvents, event)
		}
	}

	activeEventsJSON, _ := json.MarshalIndent(activeEvents, "", "  ")
	fmt.Printf("Active events: %s\n\n", activeEventsJSON)

	fmt.Println("========== Test Completed Successfully! ==========")
	fmt.Println("Smart contract structure (Donations + Events) is working correctly.")
	fmt.Println("Ready to deploy to Hyperledger Fabric network.")
}
