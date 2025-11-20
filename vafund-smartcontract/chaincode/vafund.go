package main

import (
	"fmt"

	"vafund-chaincode/entities"
	"vafund-chaincode/services"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
	donationService   *services.DonationService
	eventService      *services.EventService
	withdrawalService *services.WithdrawalService
}

func NewSmartContract() *SmartContract {
	return &SmartContract{
		donationService:   services.NewDonationService(),
		eventService:      services.NewEventService(),
		withdrawalService: services.NewWithdrawalService(),
	}
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	return nil
}

func (s *SmartContract) MakeDonation(ctx contractapi.TransactionContextInterface, donationID string, senderName string, amountStr string, message string, eventCode string) error {
	if s.donationService == nil {
		s.donationService = services.NewDonationService()
	}
	return s.donationService.Create(ctx, donationID, senderName, amountStr, message, eventCode)
}

func (s *SmartContract) ReadDonation(ctx contractapi.TransactionContextInterface, id string) (*entities.Donation, error) {
	if s.donationService == nil {
		s.donationService = services.NewDonationService()
	}
	return s.donationService.GetByID(ctx, id)
}

func (s *SmartContract) GetAllDonations(ctx contractapi.TransactionContextInterface) ([]*entities.Donation, error) {
	if s.donationService == nil {
		s.donationService = services.NewDonationService()
	}
	return s.donationService.GetAll(ctx)
}

func (s *SmartContract) GetDonationsByAmount(ctx contractapi.TransactionContextInterface, minAmount float64) ([]*entities.Donation, error) {
	if s.donationService == nil {
		s.donationService = services.NewDonationService()
	}
	return s.donationService.GetByMinAmount(ctx, minAmount)
}

func (s *SmartContract) GetTotalDonations(ctx contractapi.TransactionContextInterface) (float64, error) {
	if s.donationService == nil {
		s.donationService = services.NewDonationService()
	}
	return s.donationService.GetTotal(ctx)
}

func (s *SmartContract) DonationExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	if s.donationService == nil {
		s.donationService = services.NewDonationService()
	}
	return s.donationService.Exists(ctx, id)
}

// GetDonationsByEventCode retrieves donations associated with a specific event
func (s *SmartContract) GetDonationsByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) ([]*entities.Donation, error) {
	if s.donationService == nil {
		s.donationService = services.NewDonationService()
	}
	return s.donationService.GetByEventCode(ctx, eventCode)
}

// GetTotalDonationsByEventCode calculates total donation amount for a specific event
func (s *SmartContract) GetTotalDonationsByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) (float64, error) {
	if s.donationService == nil {
		s.donationService = services.NewDonationService()
	}
	return s.donationService.GetTotalByEventCode(ctx, eventCode)
}

// GetCurrentAmountByEventCode calculates current amount (donations - withdrawals) for a specific event
func (s *SmartContract) GetCurrentAmountByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) (float64, error) {
	// Get total donations for this event
	if s.donationService == nil {
		s.donationService = services.NewDonationService()
	}
	totalDonations, err := s.donationService.GetTotalByEventCode(ctx, eventCode)
	if err != nil {
		return 0, err
	}

	// Get total withdrawals for this event
	if s.withdrawalService == nil {
		s.withdrawalService = services.NewWithdrawalService()
	}
	totalWithdrawals, err := s.withdrawalService.GetTotalByEventCode(ctx, eventCode)
	if err != nil {
		// If no withdrawals or error, assume 0 withdrawals
		totalWithdrawals = 0
	}

	// Calculate current amount (donations - withdrawals)
	currentAmount := totalDonations - totalWithdrawals
	return currentAmount, nil
}

// ===== EVENT FUNCTIONS =====

func (s *SmartContract) CreateEvent(ctx contractapi.TransactionContextInterface, code string, name string, description string, startDateStr string, endDateStr string, isActiveStr string) error {
	if s.eventService == nil {
		s.eventService = services.NewEventService()
	}
	return s.eventService.Create(ctx, code, name, description, startDateStr, endDateStr, isActiveStr)
}

func (s *SmartContract) ReadEvent(ctx contractapi.TransactionContextInterface, code string) (*entities.Event, error) {
	if s.eventService == nil {
		s.eventService = services.NewEventService()
	}
	return s.eventService.GetByCode(ctx, code)
}

func (s *SmartContract) GetAllEvents(ctx contractapi.TransactionContextInterface) ([]*entities.Event, error) {
	if s.eventService == nil {
		s.eventService = services.NewEventService()
	}
	return s.eventService.GetAll(ctx)
}

func (s *SmartContract) GetActiveEvents(ctx contractapi.TransactionContextInterface) ([]*entities.Event, error) {
	if s.eventService == nil {
		s.eventService = services.NewEventService()
	}
	return s.eventService.GetActive(ctx)
}

func (s *SmartContract) EventExists(ctx contractapi.TransactionContextInterface, code string) (bool, error) {
	if s.eventService == nil {
		s.eventService = services.NewEventService()
	}
	return s.eventService.Exists(ctx, code)
}

func (s *SmartContract) UpdateEventStatus(ctx contractapi.TransactionContextInterface, code string, isActiveStr string) error {
	if s.eventService == nil {
		s.eventService = services.NewEventService()
	}
	return s.eventService.UpdateStatus(ctx, code, isActiveStr)
}

func (s *SmartContract) UpdateEvent(ctx contractapi.TransactionContextInterface, code string, name string, description string, startDateStr string, endDateStr string, isActiveStr string) error {
	if s.eventService == nil {
		s.eventService = services.NewEventService()
	}
	return s.eventService.Update(ctx, code, name, description, startDateStr, endDateStr, isActiveStr)
}

// ===== WITHDRAWAL FUNCTIONS =====

func (s *SmartContract) CreateWithdrawal(ctx contractapi.TransactionContextInterface, withdrawalID string, eventCode string, amountStr string, withdrawBy string, dateTimeStr string) error {
	if s.withdrawalService == nil {
		s.withdrawalService = services.NewWithdrawalService()
	}
	return s.withdrawalService.Create(ctx, withdrawalID, eventCode, amountStr, withdrawBy, dateTimeStr)
}

func (s *SmartContract) ReadWithdrawal(ctx contractapi.TransactionContextInterface, id string) (*entities.Withdrawal, error) {
	if s.withdrawalService == nil {
		s.withdrawalService = services.NewWithdrawalService()
	}
	return s.withdrawalService.GetByID(ctx, id)
}

func (s *SmartContract) GetAllWithdrawals(ctx contractapi.TransactionContextInterface) ([]*entities.Withdrawal, error) {
	if s.withdrawalService == nil {
		s.withdrawalService = services.NewWithdrawalService()
	}
	return s.withdrawalService.GetAll(ctx)
}

func (s *SmartContract) GetWithdrawalsByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) ([]*entities.Withdrawal, error) {
	if s.withdrawalService == nil {
		s.withdrawalService = services.NewWithdrawalService()
	}
	return s.withdrawalService.GetByEventCode(ctx, eventCode)
}

func (s *SmartContract) GetTotalWithdrawalsByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) (float64, error) {
	if s.withdrawalService == nil {
		s.withdrawalService = services.NewWithdrawalService()
	}
	return s.withdrawalService.GetTotalByEventCode(ctx, eventCode)
}

func (s *SmartContract) GetTotalWithdrawals(ctx contractapi.TransactionContextInterface) (float64, error) {
	if s.withdrawalService == nil {
		s.withdrawalService = services.NewWithdrawalService()
	}
	return s.withdrawalService.GetTotal(ctx)
}

func (s *SmartContract) WithdrawalExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	if s.withdrawalService == nil {
		s.withdrawalService = services.NewWithdrawalService()
	}
	return s.withdrawalService.Exists(ctx, id)
}

// ===== TEST FUNCTIONS =====

func (s *SmartContract) TestWrite(ctx contractapi.TransactionContextInterface, key string, value string) error {
	return ctx.GetStub().PutState(key, []byte(value))
}

func (s *SmartContract) TestRead(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	data, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", fmt.Errorf("key %s does not exist", key)
	}
	return string(data), nil
}

func init() {

}
