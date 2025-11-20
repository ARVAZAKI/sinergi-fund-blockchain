package controllers

import (
	"fmt"
	"strconv"

	"vafund-backend/models"
	"vafund-backend/services"

	"github.com/gofiber/fiber/v2"
)

type DonationController struct {
	fabricService *services.FabricService
}

func NewDonationController(fabricService *services.FabricService) *DonationController {
	return &DonationController{
		fabricService: fabricService,
	}
}

// CreateDonation creates a new donation
// @Summary Create a new donation
// @Description Create a new donation with sender name, amount and optional message
// @Tags donations
// @Accept json
// @Produce json
// @Param donation body models.CreateDonationRequest true "Donation data"
// @Success 201 {object} models.DonationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/donations [post]
func (dc *DonationController) CreateDonation(c *fiber.Ctx) error {
	var req models.CreateDonationRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Validate required fields
	if req.DonationID == "" || req.SenderName == "" || req.Amount == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Missing required fields: donationId, senderName, amount",
		})
	}

	// Validate amount format
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid amount format",
			Error:   err.Error(),
		})
	}

	if amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Amount must be greater than 0",
		})
	}

	// Call fabric service to create donation
	donation, err := dc.fabricService.CreateDonation(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to create donation",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.DonationResponse{
		Success: true,
		Message: "Donation created successfully",
		Data:    donation,
	})
}

// GetDonation retrieves a specific donation by ID
// @Summary Get donation by ID
// @Description Get a specific donation by its ID
// @Tags donations
// @Produce json
// @Param id path string true "Donation ID"
// @Success 200 {object} models.DonationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/donations/{id} [get]
func (dc *DonationController) GetDonation(c *fiber.Ctx) error {
	donationID := c.Params("id")

	if donationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Donation ID is required",
		})
	}

	// Call fabric service to get donation
	donation, err := dc.fabricService.GetDonation(donationID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: "Donation not found",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.DonationResponse{
		Success: true,
		Message: "Donation retrieved successfully",
		Data:    donation,
	})
}

// GetAllDonations retrieves all donations
// @Summary Get all donations
// @Description Get a list of all donations
// @Tags donations
// @Produce json
// @Success 200 {object} models.DonationResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/donations [get]
func (dc *DonationController) GetAllDonations(c *fiber.Ctx) error {
	// Call fabric service to get all donations
	donations, err := dc.fabricService.GetAllDonations()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve donations",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.DonationResponse{
		Success: true,
		Message: "Donations retrieved successfully",
		Data:    donations,
	})
}

// GetTotalDonations retrieves total donation statistics
// @Summary Get total donation statistics
// @Description Get total amount and count of all donations
// @Tags donations
// @Produce json
// @Success 200 {object} models.TotalDonationResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/donations/total [get]
func (dc *DonationController) GetTotalDonations(c *fiber.Ctx) error {
	// Call fabric service to get total donations
	totalAmount, totalCount, err := dc.fabricService.GetTotalDonations()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve total donations",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.TotalDonationResponse{
		Success:     true,
		Message:     "Total donations retrieved successfully",
		TotalAmount: totalAmount,
		TotalCount:  totalCount,
	})
}

// GetDonationsByAmount retrieves donations by minimum amount
// @Summary Get donations by minimum amount
// @Description Get a list of donations that have at least the specified amount
// @Tags donations
// @Produce json
// @Param minAmount query number true "Minimum donation amount"
// @Success 200 {object} models.DonationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/donations/by-amount [get]
func (dc *DonationController) GetDonationsByAmount(c *fiber.Ctx) error {
	minAmountStr := c.Query("minAmount")
	if minAmountStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "minAmount query parameter is required",
		})
	}

	minAmount, err := strconv.ParseFloat(minAmountStr, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid minAmount format",
			Error:   err.Error(),
		})
	}

	donations, err := dc.fabricService.GetDonationsByAmount(minAmount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve donations by amount",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.DonationResponse{
		Success: true,
		Message: fmt.Sprintf("Donations with minimum amount %.2f retrieved successfully", minAmount),
		Data:    donations,
	})
}

// GetDonationsByEventCode retrieves donations by event code
// @Summary Get donations by event code
// @Description Get a list of donations associated with a specific event
// @Tags donations
// @Produce json
// @Param eventCode path string true "Event Code"
// @Success 200 {object} models.DonationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/donations/event/{eventCode} [get]
func (dc *DonationController) GetDonationsByEventCode(c *fiber.Ctx) error {
	eventCode := c.Params("eventCode")

	if eventCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event code is required",
		})
	}

	donations, err := dc.fabricService.GetDonationsByEventCode(eventCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve donations by event code",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.DonationResponse{
		Success: true,
		Message: fmt.Sprintf("Donations for event %s retrieved successfully", eventCode),
		Data:    donations,
	})
}

// GetTotalDonationsByEventCode retrieves total donation statistics for an event
// @Summary Get total donation statistics by event code with current amount (net of withdrawals)
// @Description Get total amount and count of donations for a specific event with current amount after withdrawals
// @Tags donations
// @Produce json
// @Param eventCode path string true "Event Code"
// @Success 200 {object} models.TotalDonationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/donations/event/{eventCode}/total [get]
func (dc *DonationController) GetTotalDonationsByEventCode(c *fiber.Ctx) error {
	eventCode := c.Params("eventCode")

	if eventCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event code is required",
		})
	}

	totalAmount, totalCount, err := dc.fabricService.GetTotalDonationsByEventCode(eventCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve total donations by event code",
			Error:   err.Error(),
		})
	}

	// Get current amount (net of withdrawals) for this event
	currentAmount, err := dc.fabricService.GetCurrentAmountByEventCode(eventCode)
	if err != nil {
		// If current amount service fails, fallback to manual calculation
		totalWithdrawals, withdrawalErr := dc.fabricService.GetTotalWithdrawalsByEventCode(eventCode)
		if withdrawalErr != nil {
			totalWithdrawals = 0
		}
		currentAmount = totalAmount - totalWithdrawals
	}

	return c.JSON(models.TotalDonationResponse{
		Success:       true,
		Message:       fmt.Sprintf("Total donations for event %s retrieved successfully", eventCode),
		TotalAmount:   totalAmount,
		CurrentAmount: currentAmount,
		TotalCount:    totalCount,
	})
}
