package controllers

import (
	"fmt"
	"time"

	"vafund-backend/models"
	"vafund-backend/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// WithdrawalController handles withdrawal-related HTTP requests
type WithdrawalController struct {
	fabricService *services.FabricService
	validator     *validator.Validate
}

// NewWithdrawalController creates a new withdrawal controller instance
func NewWithdrawalController() *WithdrawalController {
	return &WithdrawalController{
		fabricService: services.NewFabricService(),
		validator:     validator.New(),
	}
}

// CreateWithdrawal godoc
// @Summary Create withdrawal
// @Description Create a new withdrawal record
// @Tags withdrawals
// @Accept json
// @Produce json
// @Param withdrawal body models.WithdrawalRequest true "Withdrawal data"
// @Success 201 {object} models.WithdrawalResponse
// @Failure 400 {object} models.WithdrawalResponse
// @Failure 500 {object} models.WithdrawalResponse
// @Router /api/withdrawals [post]
func (wc *WithdrawalController) CreateWithdrawal(c *fiber.Ctx) error {
	var req models.WithdrawalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.WithdrawalResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := wc.validator.Struct(req); err != nil {
		return c.Status(400).JSON(models.WithdrawalResponse{
			Success: false,
			Message: "Validation error: " + err.Error(),
		})
	}

	// Generate ID and current time
	withdrawalID := fmt.Sprintf("WD-%d", time.Now().Unix())
	currentTime := time.Now().Format("2006-01-02T15:04:05Z07:00")

	// Call fabric service to create withdrawal
	result, err := wc.fabricService.CreateWithdrawal(withdrawalID, req.EventCode, req.Amount, req.WithdrawBy, currentTime)
	if err != nil {
		return c.Status(500).JSON(models.WithdrawalResponse{
			Success: false,
			Message: "Failed to create withdrawal: " + err.Error(),
		})
	}

	return c.Status(201).JSON(models.WithdrawalResponse{
		Success: true,
		Message: "Withdrawal created successfully",
		Data: &models.Withdrawal{
			ID:         withdrawalID,
			EventCode:  req.EventCode,
			Amount:     req.Amount,
			WithdrawBy: req.WithdrawBy,
			TxID:       result,
		},
	})
}

// GetWithdrawal godoc
// @Summary Get withdrawal by ID
// @Description Get a specific withdrawal by ID
// @Tags withdrawals
// @Produce json
// @Param id path string true "Withdrawal ID"
// @Success 200 {object} models.WithdrawalResponse
// @Failure 404 {object} models.WithdrawalResponse
// @Failure 500 {object} models.WithdrawalResponse
// @Router /api/withdrawals/{id} [get]
func (wc *WithdrawalController) GetWithdrawal(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(models.WithdrawalResponse{
			Success: false,
			Message: "Withdrawal ID is required",
		})
	}

	withdrawal, err := wc.fabricService.GetWithdrawal(id)
	if err != nil {
		return c.Status(404).JSON(models.WithdrawalResponse{
			Success: false,
			Message: "Withdrawal not found: " + err.Error(),
		})
	}

	return c.JSON(models.WithdrawalResponse{
		Success: true,
		Message: "Withdrawal retrieved successfully",
		Data:    withdrawal,
	})
}

// GetAllWithdrawals godoc
// @Summary Get all withdrawals
// @Description Get list of all withdrawals
// @Tags withdrawals
// @Produce json
// @Success 200 {object} models.WithdrawalsResponse
// @Failure 500 {object} models.WithdrawalsResponse
// @Router /api/withdrawals [get]
func (wc *WithdrawalController) GetAllWithdrawals(c *fiber.Ctx) error {
	withdrawals, err := wc.fabricService.GetAllWithdrawals()
	if err != nil {
		return c.Status(500).JSON(models.WithdrawalsResponse{
			Success: false,
			Message: "Failed to retrieve withdrawals: " + err.Error(),
		})
	}

	return c.JSON(models.WithdrawalsResponse{
		Success: true,
		Message: "Withdrawals retrieved successfully",
		Data:    withdrawals,
	})
}

// GetWithdrawalsByEventCode godoc
// @Summary Get withdrawals by event code
// @Description Get list of withdrawals for a specific event
// @Tags withdrawals
// @Produce json
// @Param event_code path string true "Event Code"
// @Success 200 {object} models.WithdrawalsResponse
// @Failure 500 {object} models.WithdrawalsResponse
// @Router /api/withdrawals/event/{event_code} [get]
func (wc *WithdrawalController) GetWithdrawalsByEventCode(c *fiber.Ctx) error {
	eventCode := c.Params("event_code")
	if eventCode == "" {
		return c.Status(400).JSON(models.WithdrawalsResponse{
			Success: false,
			Message: "Event code is required",
		})
	}

	withdrawals, err := wc.fabricService.GetWithdrawalsByEventCode(eventCode)
	if err != nil {
		return c.Status(500).JSON(models.WithdrawalsResponse{
			Success: false,
			Message: "Failed to retrieve withdrawals: " + err.Error(),
		})
	}

	return c.JSON(models.WithdrawalsResponse{
		Success: true,
		Message: "Withdrawals retrieved successfully",
		Data:    withdrawals,
	})
}

// GetTotalWithdrawals godoc
// @Summary Get total withdrawal amount
// @Description Get total amount of all withdrawals
// @Tags withdrawals
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/withdrawals/total [get]
func (wc *WithdrawalController) GetTotalWithdrawals(c *fiber.Ctx) error {
	total, err := wc.fabricService.GetTotalWithdrawals()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get total withdrawals: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Total withdrawals retrieved successfully",
		"data": fiber.Map{
			"total": total,
		},
	})
}

// GetTotalWithdrawalsByEventCode godoc
// @Summary Get total withdrawal amount by event code
// @Description Get total amount of withdrawals for a specific event
// @Tags withdrawals
// @Produce json
// @Param event_code path string true "Event Code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/withdrawals/event/{event_code}/total [get]
func (wc *WithdrawalController) GetTotalWithdrawalsByEventCode(c *fiber.Ctx) error {
	eventCode := c.Params("event_code")
	if eventCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Event code is required",
		})
	}

	total, err := wc.fabricService.GetTotalWithdrawalsByEventCode(eventCode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get total withdrawals: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Total withdrawals retrieved successfully",
		"data": fiber.Map{
			"event_code": eventCode,
			"total":      total,
		},
	})
}
