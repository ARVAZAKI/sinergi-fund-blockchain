package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"time"

	"vafund-backend/config"
	"vafund-backend/models"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type FabricService struct {
	gateway     *client.Gateway
	contract    *client.Contract
	connection  *grpc.ClientConn
	isConnected bool
	// Fallback storage for simulation mode
	donations   map[string]*models.Donation
	events      map[string]*models.Event
	withdrawals map[string]*models.Withdrawal
	galleries   map[string]*models.Gallery
}

func NewFabricService() *FabricService {
	return &FabricService{
		donations:   make(map[string]*models.Donation),
		events:      make(map[string]*models.Event),
		withdrawals: make(map[string]*models.Withdrawal),
		galleries:   make(map[string]*models.Gallery),
	}
}

func (fs *FabricService) Connect() error {
	fmt.Println("🔄 Starting Fabric network connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := config.AppConfig
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	fmt.Printf("📋 Using configuration:\n")
	fmt.Printf("   Channel: %s\n", cfg.FabricChannel)
	fmt.Printf("   Chaincode: %s\n", cfg.FabricChaincode)
	fmt.Printf("   MSP: %s\n", cfg.FabricOrgMSP)
	fmt.Printf("   TLS Name: %s\n", cfg.FabricPeerTLSName)

	// Load crypto materials
	fmt.Println("🔐 Loading crypto materials...")
	cert, err := loadCertificate(cfg.FabricCertPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %v", err)
	}
	fmt.Printf("✅ Certificate loaded: %s\n", cfg.FabricCertPath)

	key, err := loadPrivateKey(cfg.FabricKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load private key: %v", err)
	}
	fmt.Printf("✅ Private key loaded: %s\n", cfg.FabricKeyPath)

	tlsCert, err := loadTLSCertificate(cfg.FabricTLSCAPath)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %v", err)
	}
	fmt.Printf("✅ TLS certificate loaded: %s\n", cfg.FabricTLSCAPath)

	// Create certificate pool for TLS with multiple CA certificates
	certPool := x509.NewCertPool()
	certPool.AddCert(tlsCert)

	// Also try to load additional TLS CA certificates
	additionalTLSCerts := []string{
		"./crypto/tlsca-org1.pem",
		"./crypto/peer-tls-ca.pem",
		"./crypto/orderer-tls-ca.pem",
	}

	for _, certPath := range additionalTLSCerts {
		if additionalCert, err := loadTLSCertificate(certPath); err == nil {
			certPool.AddCert(additionalCert)
			fmt.Printf("✅ Additional TLS certificate loaded: %s\n", certPath)
		} else {
			fmt.Printf("⚠️  Could not load additional TLS cert %s: %v\n", certPath, err)
		}
	}

	// Try multiple connection approaches for WSL
	var connection *grpc.ClientConn

	// Get all possible hosts from config
	allHosts := cfg.GetAllHosts()
	fmt.Printf("🔗 Will try connecting to hosts: %v\n", allHosts)

	var lastErr error
	for i, host := range allHosts {
		fmt.Printf("🔗 [Attempt %d/%d] Trying to connect to: %s\n", i+1, len(allHosts), host)

		// Try with strict TLS config first
		tlsConfig := &tls.Config{
			RootCAs:            certPool,
			ServerName:         cfg.FabricPeerTLSName,
			InsecureSkipVerify: false,
		}
		creds := credentials.NewTLS(tlsConfig)

		conn, dialErr := grpc.DialContext(ctx, host,
			grpc.WithTransportCredentials(creds),
		)

		if dialErr == nil {
			fmt.Printf("✅ Successfully connected to: %s (with strict TLS verification)\n", host)
			connection = conn
			break
		}

		fmt.Printf("❌ Strict TLS connection failed to %s: %v\n", host, dialErr)

		// Try with relaxed TLS config (skip hostname verification)
		fmt.Printf("🔄 Trying with relaxed TLS verification...\n")
		tlsConfig.InsecureSkipVerify = true
		creds = credentials.NewTLS(tlsConfig)

		conn, dialErr = grpc.DialContext(ctx, host,
			grpc.WithTransportCredentials(creds),
		)

		if dialErr == nil {
			fmt.Printf("⚠️  Connected to %s with relaxed TLS verification!\n", host)
			connection = conn
			break
		}

		fmt.Printf("❌ Relaxed TLS connection also failed to %s: %v\n", host, dialErr)
		lastErr = dialErr
	}

	if connection == nil {
		return fmt.Errorf("failed to connect to any Fabric peer: %v", lastErr)
	}

	fs.connection = connection

	// Create identity
	fmt.Printf("🆔 Creating identity with MSP: %s\n", cfg.FabricOrgMSP)
	id, err := identity.NewX509Identity(cfg.FabricOrgMSP, cert)
	if err != nil {
		return fmt.Errorf("failed to create identity: %v", err)
	}
	fmt.Println("✅ Identity created successfully")

	// Create sign function
	fmt.Println("🔑 Creating sign function...")
	sign, err := identity.NewPrivateKeySign(key)
	if err != nil {
		return fmt.Errorf("failed to create sign function: %v", err)
	}
	fmt.Printf("✅ Sign function created successfully (type: %T)\n", sign)

	// Create gateway
	fmt.Println("🚪 Creating Fabric gateway...")
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(connection),
		client.WithEvaluateTimeout(30*time.Second),
		client.WithEndorseTimeout(30*time.Second),
		client.WithSubmitTimeout(30*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %v", err)
	}
	fmt.Println("✅ Gateway created successfully")

	fs.gateway = gateway
	fs.isConnected = true

	// Get network and contract
	fmt.Printf("📡 Getting network: %s\n", cfg.FabricChannel)
	network := gateway.GetNetwork(cfg.FabricChannel)
	fmt.Printf("📦 Getting contract: %s\n", cfg.FabricChaincode)
	fs.contract = network.GetContract(cfg.FabricChaincode)

	fmt.Println("🎉 Successfully connected to Fabric network!")
	return nil
}

func (fs *FabricService) Close() {
	if fs.gateway != nil {
		fs.gateway.Close()
	}
	if fs.connection != nil {
		fs.connection.Close()
	}
}

func (fs *FabricService) CreateDonation(req models.CreateDonationRequest) (*models.Donation, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		_, err := fs.contract.SubmitTransaction("MakeDonation",
			req.DonationID,
			req.SenderName,
			req.Amount,
			req.Message,
			req.EventCode,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to submit transaction: %v", err)
		}

		// Read back the created donation
		result, err := fs.contract.EvaluateTransaction("ReadDonation", req.DonationID)
		if err != nil {
			return nil, fmt.Errorf("failed to read created donation: %v", err)
		}

		var donation models.Donation
		err = json.Unmarshal(result, &donation)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal donation: %v", err)
		}

		return &donation, nil
	} else {
		// Simulation mode
		if _, exists := fs.donations[req.DonationID]; exists {
			return nil, fmt.Errorf("donation %s already exists", req.DonationID)
		}

		amount, err := strconv.ParseFloat(req.Amount, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount format: %s", req.Amount)
		}

		if amount <= 0 {
			return nil, fmt.Errorf("amount must be greater than 0")
		}

		donation := &models.Donation{
			ID:         req.DonationID,
			SenderName: req.SenderName,
			Amount:     amount,
			Message:    req.Message,
			EventCode:  req.EventCode,
			Timestamp:  time.Now(),
			TxID:       fmt.Sprintf("sim-tx-%d", time.Now().UnixNano()),
		}

		fs.donations[req.DonationID] = donation
		return donation, nil
	}
}

func (fs *FabricService) GetDonation(donationID string) (*models.Donation, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("ReadDonation", donationID)
		if err != nil {
			return nil, fmt.Errorf("failed to read donation: %v", err)
		}

		var donation models.Donation
		err = json.Unmarshal(result, &donation)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal donation: %v", err)
		}

		return &donation, nil
	} else {
		// Simulation mode
		if donationID == "" {
			return nil, fmt.Errorf("donation ID is required")
		}

		donation, exists := fs.donations[donationID]
		if !exists {
			return nil, fmt.Errorf("donation %s does not exist", donationID)
		}

		return donation, nil
	}
}

func (fs *FabricService) GetAllDonations() ([]*models.Donation, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetAllDonations")
		if err != nil {
			return nil, fmt.Errorf("failed to get all donations: %v", err)
		}

		// Debug: print raw response
		fmt.Printf("🔍 Raw chaincode response: '%s'\n", string(result))

		// Check if result is empty
		if len(result) == 0 || string(result) == "" {
			fmt.Println("⚠️  Empty response from chaincode, returning empty list")
			return []*models.Donation{}, nil
		}

		// Check if result is "null" or similar
		resultStr := string(result)
		if resultStr == "null" || resultStr == "{}" || resultStr == "[]" {
			fmt.Println("⚠️  Null/empty response from chaincode, returning empty list")
			return []*models.Donation{}, nil
		}

		var donations []*models.Donation
		err = json.Unmarshal(result, &donations)
		if err != nil {
			fmt.Printf("❌ JSON unmarshal error: %v\n", err)
			fmt.Printf("❌ Raw data: '%s'\n", string(result))
			return nil, fmt.Errorf("failed to unmarshal donations: %v", err)
		}

		fmt.Printf("✅ Successfully parsed %d donations\n", len(donations))
		return donations, nil
	} else {
		// Simulation mode
		var donations []*models.Donation
		for _, donation := range fs.donations {
			donations = append(donations, donation)
		}
		return donations, nil
	}
}

func (fs *FabricService) GetTotalDonations() (float64, int, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetTotalDonations")
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get total donations: %v", err)
		}

		// Debug: print raw response
		fmt.Printf("🔍 Raw GetTotalDonations response: '%s'\n", string(result))

		// Check if result is empty or "0"
		if len(result) == 0 || string(result) == "" || string(result) == "0" {
			fmt.Println("⚠️  Empty/zero response from GetTotalDonations")
			return 0.0, 0, nil
		}

		totalAmount, err := strconv.ParseFloat(string(result), 64)
		if err != nil {
			fmt.Printf("❌ Failed to parse total amount '%s': %v\n", string(result), err)
			// If we can't parse, assume 0
			totalAmount = 0.0
		}

		// Get count by fetching all donations
		donations, err := fs.GetAllDonations()
		if err != nil {
			return 0, 0, err
		}

		return totalAmount, len(donations), nil
	} else {
		// Simulation mode
		donations, err := fs.GetAllDonations()
		if err != nil {
			return 0, 0, err
		}

		var totalAmount float64
		for _, donation := range donations {
			totalAmount += donation.Amount
		}

		return totalAmount, len(donations), nil
	}
}

func (fs *FabricService) GetDonationsByAmount(minAmount float64) ([]*models.Donation, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetDonationsByAmount", fmt.Sprintf("%.2f", minAmount))
		if err != nil {
			return nil, fmt.Errorf("failed to get donations by amount: %v", err)
		}

		var donations []*models.Donation
		err = json.Unmarshal(result, &donations)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal donations: %v", err)
		}

		return donations, nil
	} else {
		// Simulation mode
		allDonations, err := fs.GetAllDonations()
		if err != nil {
			return nil, err
		}

		var filteredDonations []*models.Donation
		for _, donation := range allDonations {
			if donation.Amount >= minAmount {
				filteredDonations = append(filteredDonations, donation)
			}
		}

		return filteredDonations, nil
	}
}

// GetDonationsByEventCode retrieves donations associated with a specific event
func (fs *FabricService) GetDonationsByEventCode(eventCode string) ([]*models.Donation, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetDonationsByEventCode", eventCode)
		if err != nil {
			return nil, fmt.Errorf("failed to get donations by event code: %v", err)
		}

		// Debug: print raw response
		fmt.Printf("🔍 Raw GetDonationsByEventCode response: '%s'\n", string(result))

		// Check if result is empty
		if len(result) == 0 || string(result) == "" {
			fmt.Println("⚠️  Empty response from GetDonationsByEventCode, returning empty list")
			return []*models.Donation{}, nil
		}

		// Check if result is "null" or similar
		resultStr := string(result)
		if resultStr == "null" || resultStr == "{}" || resultStr == "[]" {
			fmt.Println("⚠️  Null/empty response from GetDonationsByEventCode, returning empty list")
			return []*models.Donation{}, nil
		}

		var donations []*models.Donation
		err = json.Unmarshal(result, &donations)
		if err != nil {
			fmt.Printf("❌ JSON unmarshal error for donations by event: %v\n", err)
			fmt.Printf("❌ Raw data: '%s'\n", string(result))
			return nil, fmt.Errorf("failed to unmarshal donations by event: %v", err)
		}

		fmt.Printf("✅ Successfully parsed %d donations for event %s\n", len(donations), eventCode)
		return donations, nil
	} else {
		// Simulation mode
		var eventDonations []*models.Donation
		for _, donation := range fs.donations {
			if donation.EventCode == eventCode {
				eventDonations = append(eventDonations, donation)
			}
		}
		return eventDonations, nil
	}
}

// GetTotalDonationsByEventCode calculates total donation amount for a specific event
func (fs *FabricService) GetTotalDonationsByEventCode(eventCode string) (float64, int, error) {
	donations, err := fs.GetDonationsByEventCode(eventCode)
	if err != nil {
		return 0, 0, err
	}

	var total float64
	for _, donation := range donations {
		total += donation.Amount
	}

	return total, len(donations), nil
}

// GetCurrentAmountByEventCode calculates current amount (donations - withdrawals) for a specific event
func (fs *FabricService) GetCurrentAmountByEventCode(eventCode string) (float64, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode - use smart contract function
		result, err := fs.contract.EvaluateTransaction("GetCurrentAmountByEventCode", eventCode)
		if err != nil {
			return 0, fmt.Errorf("failed to get current amount by event code: %v", err)
		}

		currentAmount, err := strconv.ParseFloat(string(result), 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse current amount: %v", err)
		}

		return currentAmount, nil
	} else {
		// Simulation mode - calculate manually
		totalDonations := 0.0
		for _, donation := range fs.donations {
			if donation.EventCode == eventCode {
				totalDonations += donation.Amount
			}
		}

		totalWithdrawals := 0.0
		for _, withdrawal := range fs.withdrawals {
			if withdrawal.EventCode == eventCode {
				totalWithdrawals += withdrawal.Amount
			}
		}

		return totalDonations - totalWithdrawals, nil
	}
}

// ===== EVENT FUNCTIONS =====

func (fs *FabricService) CreateEvent(req models.CreateEventRequest) (*models.Event, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		_, err := fs.contract.SubmitTransaction("CreateEvent",
			req.Code,
			req.Name,
			req.Description,
			req.StartDate,
			req.EndDate,
			req.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to submit transaction: %v", err)
		}

		// Read back the created event
		result, err := fs.contract.EvaluateTransaction("ReadEvent", req.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to read created event: %v", err)
		}

		var event models.Event
		err = json.Unmarshal(result, &event)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %v", err)
		}

		return &event, nil
	} else {
		// Simulation mode
		if _, exists := fs.events[req.Code]; exists {
			return nil, fmt.Errorf("event %s already exists", req.Code)
		}

		// Parse dates
		startDate, err := time.Parse("2006-01-02T15:04:05Z07:00", req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date format: %s", req.StartDate)
		}

		endDate, err := time.Parse("2006-01-02T15:04:05Z07:00", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format: %s", req.EndDate)
		}

		// Validate date range
		if endDate.Before(startDate) || endDate.Equal(startDate) {
			return nil, fmt.Errorf("end date must be after start date")
		}

		// Parse isActive
		isActive := req.IsActive == "true"

		event := &models.Event{
			Code:      req.Code,
			Name:      req.Name,
			StartDate: startDate,
			EndDate:   endDate,
			IsActive:  isActive,
			Timestamp: time.Now(),
			TxID:      fmt.Sprintf("sim-event-tx-%d", time.Now().UnixNano()),
		}

		fs.events[req.Code] = event
		return event, nil
	}
}

func (fs *FabricService) GetEvent(eventCode string) (*models.Event, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("ReadEvent", eventCode)
		if err != nil {
			return nil, fmt.Errorf("failed to read event: %v", err)
		}

		var event models.Event
		err = json.Unmarshal(result, &event)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %v", err)
		}

		return &event, nil
	} else {
		// Simulation mode
		if eventCode == "" {
			return nil, fmt.Errorf("event code is required")
		}

		event, exists := fs.events[eventCode]
		if !exists {
			return nil, fmt.Errorf("event %s does not exist", eventCode)
		}

		return event, nil
	}
}

func (fs *FabricService) GetAllEvents() ([]*models.Event, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetAllEvents")
		if err != nil {
			return nil, fmt.Errorf("failed to get all events: %v", err)
		}

		// Debug: print raw response
		fmt.Printf("🔍 Raw GetAllEvents response: '%s'\n", string(result))

		// Check if result is empty
		if len(result) == 0 || string(result) == "" {
			fmt.Println("⚠️  Empty response from GetAllEvents, returning empty list")
			return []*models.Event{}, nil
		}

		// Check if result is "null" or similar
		resultStr := string(result)
		if resultStr == "null" || resultStr == "{}" || resultStr == "[]" {
			fmt.Println("⚠️  Null/empty response from GetAllEvents, returning empty list")
			return []*models.Event{}, nil
		}

		var events []*models.Event
		err = json.Unmarshal(result, &events)
		if err != nil {
			fmt.Printf("❌ JSON unmarshal error for events: %v\n", err)
			fmt.Printf("❌ Raw data: '%s'\n", string(result))
			return nil, fmt.Errorf("failed to unmarshal events: %v", err)
		}

		fmt.Printf("✅ Successfully parsed %d events\n", len(events))
		return events, nil
	} else {
		// Simulation mode
		var events []*models.Event
		for _, event := range fs.events {
			events = append(events, event)
		}
		return events, nil
	}
}

func (fs *FabricService) GetActiveEvents() ([]*models.Event, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetActiveEvents")
		if err != nil {
			return nil, fmt.Errorf("failed to get active events: %v", err)
		}

		// Debug: print raw response
		fmt.Printf("🔍 Raw GetActiveEvents response: '%s'\n", string(result))

		// Check if result is empty
		if len(result) == 0 || string(result) == "" {
			fmt.Println("⚠️  Empty response from GetActiveEvents, returning empty list")
			return []*models.Event{}, nil
		}

		// Check if result is "null" or similar
		resultStr := string(result)
		if resultStr == "null" || resultStr == "{}" || resultStr == "[]" {
			fmt.Println("⚠️  Null/empty response from GetActiveEvents, returning empty list")
			return []*models.Event{}, nil
		}

		var events []*models.Event
		err = json.Unmarshal(result, &events)
		if err != nil {
			fmt.Printf("❌ JSON unmarshal error for active events: %v\n", err)
			fmt.Printf("❌ Raw data: '%s'\n", string(result))
			return nil, fmt.Errorf("failed to unmarshal active events: %v", err)
		}

		fmt.Printf("✅ Successfully parsed %d active events\n", len(events))
		return events, nil
	} else {
		// Simulation mode
		allEvents, err := fs.GetAllEvents()
		if err != nil {
			return nil, err
		}

		var activeEvents []*models.Event
		currentTime := time.Now()

		for _, event := range allEvents {
			// Check if event is active and within date range
			if event.IsActive && currentTime.After(event.StartDate) && currentTime.Before(event.EndDate) {
				activeEvents = append(activeEvents, event)
			}
		}

		return activeEvents, nil
	}
}

func (fs *FabricService) UpdateEventStatus(eventCode string, isActiveStr string) (*models.Event, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		_, err := fs.contract.SubmitTransaction("UpdateEventStatus", eventCode, isActiveStr)
		if err != nil {
			return nil, fmt.Errorf("failed to update event status: %v", err)
		}

		// Read back the updated event
		result, err := fs.contract.EvaluateTransaction("ReadEvent", eventCode)
		if err != nil {
			return nil, fmt.Errorf("failed to read updated event: %v", err)
		}

		var event models.Event
		err = json.Unmarshal(result, &event)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal updated event: %v", err)
		}

		return &event, nil
	} else {
		// Simulation mode
		event, exists := fs.events[eventCode]
		if !exists {
			return nil, fmt.Errorf("event not found")
		}

		// Update the status
		event.IsActive = (isActiveStr == "true")

		fs.events[eventCode] = event
		return event, nil
	}
}

func (fs *FabricService) UpdateEvent(eventCode string, req models.UpdateEventDetailRequest) (*models.Event, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		_, err := fs.contract.SubmitTransaction("UpdateEvent", eventCode, req.Name, req.Description, req.StartDate, req.EndDate, req.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to update event: %v", err)
		}

		// Read back the updated event
		result, err := fs.contract.EvaluateTransaction("ReadEvent", eventCode)
		if err != nil {
			return nil, fmt.Errorf("failed to read updated event: %v", err)
		}

		var event models.Event
		err = json.Unmarshal(result, &event)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal updated event: %v", err)
		}

		return &event, nil
	} else {
		// Simulation mode
		event, exists := fs.events[eventCode]
		if !exists {
			return nil, fmt.Errorf("event not found")
		}

		// Parse dates for simulation
		if startDate, err := time.Parse("2006-01-02T15:04:05Z07:00", req.StartDate); err == nil {
			event.StartDate = startDate
		}
		if endDate, err := time.Parse("2006-01-02T15:04:05Z07:00", req.EndDate); err == nil {
			event.EndDate = endDate
		}

		// Update other fields
		event.Name = req.Name
		event.Description = req.Description
		event.IsActive = (req.IsActive == "true")
		event.Timestamp = time.Now()

		fs.events[eventCode] = event
		return event, nil
	}
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %v", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

func loadPrivateKey(filename string) (*ecdsa.PrivateKey, error) {
	privateKeyPEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	// Parse PEM block
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	// Parse PKCS#8 private key
	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#8 private key: %v", err)
	}

	// Convert to ECDSA private key
	privateKey, ok := privateKeyInterface.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("expected ECDSA private key, got %T", privateKeyInterface)
	}

	return privateKey, nil
}

func loadTLSCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS certificate file: %v", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// ===== WITHDRAWAL FUNCTIONS =====

func (fs *FabricService) CreateWithdrawal(withdrawalID, eventCode string, amount float64, withdrawBy, dateTime string) (string, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		amountStr := fmt.Sprintf("%.2f", amount)

		result, err := fs.contract.SubmitTransaction("CreateWithdrawal", withdrawalID, eventCode, amountStr, withdrawBy, dateTime)
		if err != nil {
			return "", fmt.Errorf("failed to create withdrawal: %v", err)
		}

		fmt.Printf("✅ Withdrawal created successfully. TxID: %s\n", string(result))
		return string(result), nil
	} else {
		// Simulation mode
		withdrawal := &models.Withdrawal{
			ID:         withdrawalID,
			EventCode:  eventCode,
			Amount:     amount,
			WithdrawBy: withdrawBy,
			Timestamp:  time.Now(),
			TxID:       fmt.Sprintf("sim-tx-%d", time.Now().Unix()),
		}

		// Parse datetime
		if dt, err := time.Parse("2006-01-02T15:04:05Z07:00", dateTime); err == nil {
			withdrawal.DateTime = dt
		} else {
			withdrawal.DateTime = time.Now()
		}

		fs.withdrawals[withdrawalID] = withdrawal
		fmt.Printf("📝 Withdrawal created in simulation mode: %s\n", withdrawalID)
		return withdrawal.TxID, nil
	}
}

func (fs *FabricService) GetWithdrawal(withdrawalID string) (*models.Withdrawal, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("ReadWithdrawal", withdrawalID)
		if err != nil {
			return nil, fmt.Errorf("failed to read withdrawal: %v", err)
		}

		var withdrawal models.Withdrawal
		err = json.Unmarshal(result, &withdrawal)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal withdrawal: %v", err)
		}

		return &withdrawal, nil
	} else {
		// Simulation mode
		withdrawal, exists := fs.withdrawals[withdrawalID]
		if !exists {
			return nil, fmt.Errorf("withdrawal not found")
		}
		return withdrawal, nil
	}
}

func (fs *FabricService) GetAllWithdrawals() ([]*models.Withdrawal, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetAllWithdrawals")
		if err != nil {
			return nil, fmt.Errorf("failed to get all withdrawals: %v", err)
		}

		var withdrawals []*models.Withdrawal
		err = json.Unmarshal(result, &withdrawals)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal withdrawals: %v", err)
		}

		return withdrawals, nil
	} else {
		// Simulation mode
		var withdrawals []*models.Withdrawal
		for _, withdrawal := range fs.withdrawals {
			withdrawals = append(withdrawals, withdrawal)
		}
		return withdrawals, nil
	}
}

func (fs *FabricService) GetWithdrawalsByEventCode(eventCode string) ([]*models.Withdrawal, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetWithdrawalsByEventCode", eventCode)
		if err != nil {
			return nil, fmt.Errorf("failed to get withdrawals by event code: %v", err)
		}

		var withdrawals []*models.Withdrawal
		err = json.Unmarshal(result, &withdrawals)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal withdrawals: %v", err)
		}

		return withdrawals, nil
	} else {
		// Simulation mode
		var eventWithdrawals []*models.Withdrawal
		for _, withdrawal := range fs.withdrawals {
			if withdrawal.EventCode == eventCode {
				eventWithdrawals = append(eventWithdrawals, withdrawal)
			}
		}
		return eventWithdrawals, nil
	}
}

func (fs *FabricService) GetTotalWithdrawals() (float64, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetTotalWithdrawals")
		if err != nil {
			return 0, fmt.Errorf("failed to get total withdrawals: %v", err)
		}

		total, err := strconv.ParseFloat(string(result), 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse total withdrawals: %v", err)
		}

		return total, nil
	} else {
		// Simulation mode
		var total float64
		for _, withdrawal := range fs.withdrawals {
			total += withdrawal.Amount
		}
		return total, nil
	}
}

func (fs *FabricService) GetTotalWithdrawalsByEventCode(eventCode string) (float64, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetTotalWithdrawalsByEventCode", eventCode)
		if err != nil {
			return 0, fmt.Errorf("failed to get total withdrawals by event code: %v", err)
		}

		total, err := strconv.ParseFloat(string(result), 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse total withdrawals: %v", err)
		}

		return total, nil
	} else {
		// Simulation mode
		var total float64
		for _, withdrawal := range fs.withdrawals {
			if withdrawal.EventCode == eventCode {
				total += withdrawal.Amount
			}
		}
		return total, nil
	}
}

// ===== GALLERY FUNCTIONS =====

func (fs *FabricService) CreateGallery(id string, eventCode string, imageURL string, description string) (*models.Gallery, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		_, err := fs.contract.SubmitTransaction("CreateGallery", id, eventCode, imageURL, description)
		if err != nil {
			return nil, fmt.Errorf("failed to submit transaction: %v", err)
		}

		// Read back the created gallery
		result, err := fs.contract.EvaluateTransaction("ReadGallery", id)
		if err != nil {
			return nil, fmt.Errorf("failed to read created gallery: %v", err)
		}

		var gallery models.Gallery
		err = json.Unmarshal(result, &gallery)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal gallery: %v", err)
		}

		return &gallery, nil
	} else {
		// Simulation mode
		if _, exists := fs.galleries[id]; exists {
			return nil, fmt.Errorf("gallery %s already exists", id)
		}

		gallery := &models.Gallery{
			ID:          id,
			EventCode:   eventCode,
			ImageURL:    imageURL,
			Description: description,
			Timestamp:   time.Now(),
			TxID:        fmt.Sprintf("sim-tx-%d", time.Now().UnixNano()),
		}

		fs.galleries[id] = gallery
		return gallery, nil
	}
}

func (fs *FabricService) GetGallery(galleryID string) (*models.Gallery, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("ReadGallery", galleryID)
		if err != nil {
			return nil, fmt.Errorf("failed to read gallery: %v", err)
		}

		var gallery models.Gallery
		err = json.Unmarshal(result, &gallery)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal gallery: %v", err)
		}

		return &gallery, nil
	} else {
		// Simulation mode
		if galleryID == "" {
			return nil, fmt.Errorf("gallery ID is required")
		}

		gallery, exists := fs.galleries[galleryID]
		if !exists {
			return nil, fmt.Errorf("gallery %s does not exist", galleryID)
		}

		return gallery, nil
	}
}

func (fs *FabricService) GetAllGalleries() ([]*models.Gallery, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetAllGalleries")
		if err != nil {
			return nil, fmt.Errorf("failed to get all galleries: %v", err)
		}

		// Check if result is empty
		if len(result) == 0 || string(result) == "" || string(result) == "null" || string(result) == "[]" {
			return []*models.Gallery{}, nil
		}

		var galleries []*models.Gallery
		err = json.Unmarshal(result, &galleries)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal galleries: %v", err)
		}

		return galleries, nil
	} else {
		// Simulation mode
		var galleries []*models.Gallery
		for _, gallery := range fs.galleries {
			galleries = append(galleries, gallery)
		}
		return galleries, nil
	}
}

func (fs *FabricService) GetGalleriesByEventCode(eventCode string) ([]*models.Gallery, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		result, err := fs.contract.EvaluateTransaction("GetGalleriesByEventCode", eventCode)
		if err != nil {
			return nil, fmt.Errorf("failed to get galleries by event code: %v", err)
		}

		// Check if result is empty
		if len(result) == 0 || string(result) == "" || string(result) == "null" || string(result) == "[]" {
			return []*models.Gallery{}, nil
		}

		var galleries []*models.Gallery
		err = json.Unmarshal(result, &galleries)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal galleries: %v", err)
		}

		return galleries, nil
	} else {
		// Simulation mode
		var eventGalleries []*models.Gallery
		for _, gallery := range fs.galleries {
			if gallery.EventCode == eventCode {
				eventGalleries = append(eventGalleries, gallery)
			}
		}
		return eventGalleries, nil
	}
}

func (fs *FabricService) UpdateGallery(id string, eventCode string, imageURL string, description string) (*models.Gallery, error) {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		_, err := fs.contract.SubmitTransaction("UpdateGallery", id, eventCode, imageURL, description)
		if err != nil {
			return nil, fmt.Errorf("failed to submit transaction: %v", err)
		}

		// Read back the updated gallery
		result, err := fs.contract.EvaluateTransaction("ReadGallery", id)
		if err != nil {
			return nil, fmt.Errorf("failed to read updated gallery: %v", err)
		}

		var gallery models.Gallery
		err = json.Unmarshal(result, &gallery)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal gallery: %v", err)
		}

		return &gallery, nil
	} else {
		// Simulation mode
		gallery, exists := fs.galleries[id]
		if !exists {
			return nil, fmt.Errorf("gallery %s does not exist", id)
		}

		gallery.EventCode = eventCode
		gallery.ImageURL = imageURL
		gallery.Description = description
		gallery.Timestamp = time.Now()
		gallery.TxID = fmt.Sprintf("sim-tx-%d", time.Now().UnixNano())

		return gallery, nil
	}
}

func (fs *FabricService) DeleteGallery(galleryID string) error {
	if fs.isConnected && fs.contract != nil {
		// Real Fabric mode
		_, err := fs.contract.SubmitTransaction("DeleteGallery", galleryID)
		if err != nil {
			return fmt.Errorf("failed to delete gallery: %v", err)
		}
		return nil
	} else {
		// Simulation mode
		if _, exists := fs.galleries[galleryID]; !exists {
			return fmt.Errorf("gallery %s does not exist", galleryID)
		}

		delete(fs.galleries, galleryID)
		return nil
	}
}

