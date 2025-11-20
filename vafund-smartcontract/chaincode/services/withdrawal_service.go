package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"vafund-chaincode/entities"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// WithdrawalService handles withdrawal-related operations
type WithdrawalService struct{}

// NewWithdrawalService creates a new withdrawal service instance
func NewWithdrawalService() *WithdrawalService {
	return &WithdrawalService{}
}

// Create creates a new withdrawal
func (ws *WithdrawalService) Create(ctx contractapi.TransactionContextInterface, withdrawalID string, eventCode string, amountStr string, withdrawBy string, dateTimeStr string) error {
	exists, err := ws.Exists(ctx, withdrawalID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("withdrawal %s already exists", withdrawalID)
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount format: %s", amountStr)
	}

	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	// Parse dateTime
	dateTime, err := time.Parse("2006-01-02T15:04:05Z07:00", dateTimeStr)
	if err != nil {
		return fmt.Errorf("invalid dateTime format: %s. Use RFC3339 format (e.g., 2025-11-17T10:00:00Z)", dateTimeStr)
	}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}

	withdrawal := entities.NewWithdrawal(
		withdrawalID,
		eventCode,
		amount,
		withdrawBy,
		dateTime,
		time.Unix(timestamp.Seconds, int64(timestamp.Nanos)),
		ctx.GetStub().GetTxID(),
	)

	if !withdrawal.IsValid() {
		return fmt.Errorf("invalid withdrawal data")
	}

	withdrawalJSON, err := json.Marshal(withdrawal)
	if err != nil {
		return fmt.Errorf("failed to marshal withdrawal: %v", err)
	}

	// Use prefix to distinguish withdrawals from other entities
	return ctx.GetStub().PutState("WITHDRAWAL_"+withdrawalID, withdrawalJSON)
}

// GetByID retrieves a withdrawal by ID
func (ws *WithdrawalService) GetByID(ctx contractapi.TransactionContextInterface, id string) (*entities.Withdrawal, error) {
	withdrawalJSON, err := ctx.GetStub().GetState("WITHDRAWAL_" + id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if withdrawalJSON == nil {
		return nil, fmt.Errorf("withdrawal %s does not exist", id)
	}

	var withdrawal entities.Withdrawal
	err = json.Unmarshal(withdrawalJSON, &withdrawal)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal withdrawal: %v", err)
	}

	return &withdrawal, nil
}

// GetAll retrieves all withdrawals
func (ws *WithdrawalService) GetAll(ctx contractapi.TransactionContextInterface) ([]*entities.Withdrawal, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("WITHDRAWAL_", "WITHDRAWAL_\uffff")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var withdrawals []*entities.Withdrawal
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var withdrawal entities.Withdrawal
		err = json.Unmarshal(queryResponse.Value, &withdrawal)
		if err != nil {
			continue // Skip invalid entries
		}

		if withdrawal.IsValid() {
			withdrawals = append(withdrawals, &withdrawal)
		}
	}

	return withdrawals, nil
}

// GetByEventCode retrieves withdrawals associated with a specific event
func (ws *WithdrawalService) GetByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) ([]*entities.Withdrawal, error) {
	allWithdrawals, err := ws.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var eventWithdrawals []*entities.Withdrawal
	for _, withdrawal := range allWithdrawals {
		if withdrawal.EventCode == eventCode {
			eventWithdrawals = append(eventWithdrawals, withdrawal)
		}
	}

	return eventWithdrawals, nil
}

// GetTotalByEventCode calculates the total withdrawal amount for a specific event
func (ws *WithdrawalService) GetTotalByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) (float64, error) {
	eventWithdrawals, err := ws.GetByEventCode(ctx, eventCode)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, withdrawal := range eventWithdrawals {
		total += withdrawal.Amount
	}

	return total, nil
}

// GetTotal calculates the total withdrawal amount
func (ws *WithdrawalService) GetTotal(ctx contractapi.TransactionContextInterface) (float64, error) {
	allWithdrawals, err := ws.GetAll(ctx)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, withdrawal := range allWithdrawals {
		total += withdrawal.Amount
	}

	return total, nil
}

// Exists checks if a withdrawal with the given ID exists
func (ws *WithdrawalService) Exists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	withdrawalJSON, err := ctx.GetStub().GetState("WITHDRAWAL_" + id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return withdrawalJSON != nil, nil
}
