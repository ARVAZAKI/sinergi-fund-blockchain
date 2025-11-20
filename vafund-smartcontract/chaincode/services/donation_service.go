package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"vafund-chaincode/entities"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type DonationService struct{}

func NewDonationService() *DonationService {
	return &DonationService{}
}

func (ds *DonationService) Create(ctx contractapi.TransactionContextInterface, donationID string, senderName string, amountStr string, message string, eventCode string) error {
	exists, err := ds.Exists(ctx, donationID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("donation %s already exists", donationID)
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount format: %s", amountStr)
	}

	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}

	donation := entities.NewDonation(
		donationID,
		senderName,
		amount,
		message,
		eventCode,
		time.Unix(timestamp.Seconds, int64(timestamp.Nanos)),
		ctx.GetStub().GetTxID(),
	)

	if !donation.IsValid() {
		return fmt.Errorf("invalid donation data")
	}

	donationJSON, err := json.Marshal(donation)
	if err != nil {
		return fmt.Errorf("failed to marshal donation: %v", err)
	}

	return ctx.GetStub().PutState(donationID, donationJSON)
}

func (ds *DonationService) GetByID(ctx contractapi.TransactionContextInterface, id string) (*entities.Donation, error) {
	donationJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if donationJSON == nil {
		return nil, fmt.Errorf("donation %s does not exist", id)
	}

	var donation entities.Donation
	err = json.Unmarshal(donationJSON, &donation)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal donation: %v", err)
	}

	return &donation, nil
}

func (ds *DonationService) GetAll(ctx contractapi.TransactionContextInterface) ([]*entities.Donation, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var donations []*entities.Donation
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		// Skip non-donation keys
		key := queryResponse.Key
		if len(key) > 6 && key[:6] == "EVENT_" {
			continue // Skip events
		}
		if len(key) > 11 && key[:11] == "WITHDRAWAL_" {
			continue // Skip withdrawals
		}

		var donation entities.Donation
		err = json.Unmarshal(queryResponse.Value, &donation)
		if err != nil {
			continue // Skip invalid JSON
		}

		// Additional validation to ensure it's a donation
		if donation.SenderName != "" && donation.Amount > 0 {
			donations = append(donations, &donation)
		}
	}

	return donations, nil
}

func (ds *DonationService) GetByMinAmount(ctx contractapi.TransactionContextInterface, minAmount float64) ([]*entities.Donation, error) {
	allDonations, err := ds.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var filteredDonations []*entities.Donation
	for _, donation := range allDonations {
		if donation.Amount >= minAmount {
			filteredDonations = append(filteredDonations, donation)
		}
	}

	return filteredDonations, nil
}

// GetByEventCode retrieves donations associated with a specific event
func (ds *DonationService) GetByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) ([]*entities.Donation, error) {
	allDonations, err := ds.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var eventDonations []*entities.Donation
	for _, donation := range allDonations {
		if donation.EventCode == eventCode {
			eventDonations = append(eventDonations, donation)
		}
	}

	return eventDonations, nil
}

// GetTotalByEventCode calculates the total donation amount for a specific event
func (ds *DonationService) GetTotalByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) (float64, error) {
	eventDonations, err := ds.GetByEventCode(ctx, eventCode)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, donation := range eventDonations {
		total += donation.Amount
	}

	return total, nil
}

func (ds *DonationService) GetTotal(ctx contractapi.TransactionContextInterface) (float64, error) {
	allDonations, err := ds.GetAll(ctx)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, donation := range allDonations {
		total += donation.Amount
	}

	return total, nil
}

func (ds *DonationService) Exists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	donationJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return donationJSON != nil, nil
}
