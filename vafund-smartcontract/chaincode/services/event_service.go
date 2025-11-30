package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"vafund-chaincode/entities"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type EventService struct{}

func NewEventService() *EventService {
	return &EventService{}
}

func (es *EventService) Create(ctx contractapi.TransactionContextInterface, code string, name string, description string, imgUrl string, startDateStr string, endDateStr string, isActiveStr string) error {
	exists, err := es.Exists(ctx, code)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("event %s already exists", code)
	}

	startDate, err := time.Parse("2006-01-02T15:04:05Z07:00", startDateStr)
	if err != nil {
		return fmt.Errorf("invalid start date format: %s. Use RFC3339 format (e.g., 2025-01-01T00:00:00Z)", startDateStr)
	}

	endDate, err := time.Parse("2006-01-02T15:04:05Z07:00", endDateStr)
	if err != nil {
		return fmt.Errorf("invalid end date format: %s. Use RFC3339 format (e.g., 2025-12-31T23:59:59Z)", endDateStr)
	}

	isActive, err := strconv.ParseBool(isActiveStr)
	if err != nil {
		return fmt.Errorf("invalid isActive format: %s. Use 'true' or 'false'", isActiveStr)
	}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}

	event := entities.NewEvent(
		code,
		name,
		description,
		imgUrl,
		startDate,
		endDate,
		isActive,
		time.Unix(timestamp.Seconds, int64(timestamp.Nanos)),
		ctx.GetStub().GetTxID(),
	)

	if !event.IsValid() {
		return fmt.Errorf("invalid event data: end date cannot be before start date")
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	return ctx.GetStub().PutState("EVENT_"+code, eventJSON)
}

func (es *EventService) GetByCode(ctx contractapi.TransactionContextInterface, code string) (*entities.Event, error) {
	eventJSON, err := ctx.GetStub().GetState("EVENT_" + code)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if eventJSON == nil {
		return nil, fmt.Errorf("event %s does not exist", code)
	}

	var event entities.Event
	err = json.Unmarshal(eventJSON, &event)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %v", err)
	}

	return &event, nil
}

func (es *EventService) GetAll(ctx contractapi.TransactionContextInterface) ([]*entities.Event, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("EVENT_", "EVENT_\uffff")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var events []*entities.Event
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var event entities.Event
		err = json.Unmarshal(queryResponse.Value, &event)
		if err != nil {
			continue
		}

		if event.IsValid() {
			events = append(events, &event)
		}
	}

	return events, nil
}

func (es *EventService) GetActive(ctx contractapi.TransactionContextInterface) ([]*entities.Event, error) {
	allEvents, err := es.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var activeEvents []*entities.Event
	for _, event := range allEvents {
		if event.IsCurrentlyActive() {
			activeEvents = append(activeEvents, event)
		}
	}

	return activeEvents, nil
}

func (es *EventService) Exists(ctx contractapi.TransactionContextInterface, code string) (bool, error) {
	eventJSON, err := ctx.GetStub().GetState("EVENT_" + code)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return eventJSON != nil, nil
}

func (es *EventService) UpdateStatus(ctx contractapi.TransactionContextInterface, code string, isActiveStr string) error {
	event, err := es.GetByCode(ctx, code)
	if err != nil {
		return err
	}

	isActive, err := strconv.ParseBool(isActiveStr)
	if err != nil {
		return fmt.Errorf("invalid isActive format: %s. Use 'true' or 'false'", isActiveStr)
	}

	event.IsActive = isActive

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	return ctx.GetStub().PutState("EVENT_"+code, eventJSON)
}

// Update updates an event with new information
func (es *EventService) Update(ctx contractapi.TransactionContextInterface, code string, name string, description string, imgUrl string, startDateStr string, endDateStr string, isActiveStr string) error {
	// Check if event exists
	exists, err := es.Exists(ctx, code)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("event %s does not exist", code)
	}

	// Parse dates with multiple format support
	var startDate, endDate time.Time

	// Try multiple date formats
	dateFormats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	// Parse start date
	for _, format := range dateFormats {
		if startDate, err = time.Parse(format, startDateStr); err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("invalid startDate format: %s. Supported formats: RFC3339 (2025-03-01T00:00:00Z), ISO8601, etc", startDateStr)
	}

	// Parse end date
	for _, format := range dateFormats {
		if endDate, err = time.Parse(format, endDateStr); err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("invalid endDate format: %s. Supported formats: RFC3339 (2025-04-30T23:59:59Z), ISO8601, etc", endDateStr)
	}

	// Parse isActive
	isActive, err := strconv.ParseBool(isActiveStr)
	if err != nil {
		return fmt.Errorf("invalid isActive value: %s. Use 'true' or 'false'", isActiveStr)
	}

	// Validate dates
	if endDate.Before(startDate) {
		return fmt.Errorf("end date (%s) cannot be before start date (%s)", endDate.Format("2006-01-02"), startDate.Format("2006-01-02"))
	}

	// Validate input parameters
	if name == "" {
		return fmt.Errorf("event name cannot be empty")
	}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}

	// Create updated event
	event := entities.NewEvent(
		code,
		name,
		description,
		imgUrl,
		startDate,
		endDate,
		isActive,
		time.Unix(timestamp.Seconds, int64(timestamp.Nanos)),
		ctx.GetStub().GetTxID(),
	)

	if !event.IsValid() {
		return fmt.Errorf("invalid event data: code=%s, name=%s, startDate=%s, endDate=%s", code, name, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	return ctx.GetStub().PutState("EVENT_"+code, eventJSON)
}
