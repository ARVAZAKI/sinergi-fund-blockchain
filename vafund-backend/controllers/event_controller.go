package controllers

import (
	"strings"
	"time"

	"vafund-backend/models"
	"vafund-backend/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type EventController struct {
	fabricService *services.FabricService
	validator     *validator.Validate
}

func NewEventController(fabricService *services.FabricService) *EventController {
	return &EventController{
		fabricService: fabricService,
		validator:     validator.New(),
	}
}

// CreateEvent creates a new event
// @Summary Create a new event
// @Description Create a new event with code, name, start date, end date and active status
// @Tags events
// @Accept json
// @Produce json
// @Param event body models.CreateEventRequest true "Event data"
// @Success 201 {object} models.EventResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/events [post]
func (ec *EventController) CreateEvent(c *fiber.Ctx) error {
	var req models.CreateEventRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Validate required fields
	if req.Code == "" || req.Name == "" || req.StartDate == "" || req.EndDate == "" || req.IsActive == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Missing required fields: code, name, startDate, endDate, isActive",
		})
	}

	// Validate date formats
	startDate, err := time.Parse("2006-01-02T15:04:05Z07:00", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid start date format. Use RFC3339 format (e.g., 2025-01-01T00:00:00Z)",
			Error:   err.Error(),
		})
	}

	endDate, err := time.Parse("2006-01-02T15:04:05Z07:00", req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid end date format. Use RFC3339 format (e.g., 2025-12-31T23:59:59Z)",
			Error:   err.Error(),
		})
	}

	// Validate date range
	if endDate.Before(startDate) || endDate.Equal(startDate) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "End date must be after start date",
		})
	}

	// Validate isActive
	if req.IsActive != "true" && req.IsActive != "false" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "isActive must be 'true' or 'false'",
		})
	}

	// Call fabric service to create event
	event, err := ec.fabricService.CreateEvent(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to create event",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.EventResponse{
		Success: true,
		Message: "Event created successfully",
		Data:    event,
	})
}

// GetEvent retrieves a specific event by code
// @Summary Get event by code
// @Description Get a specific event by its code
// @Tags events
// @Produce json
// @Param code path string true "Event Code"
// @Success 200 {object} models.EventResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/events/{code} [get]
func (ec *EventController) GetEvent(c *fiber.Ctx) error {
	eventCode := c.Params("code")

	if eventCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event code is required",
		})
	}

	// Call fabric service to get event
	event, err := ec.fabricService.GetEvent(eventCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event not found",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.EventResponse{
		Success: true,
		Message: "Event retrieved successfully",
		Data:    event,
	})
}

// GetAllEvents retrieves all events
// @Summary Get all events
// @Description Get a list of all events
// @Tags events
// @Produce json
// @Success 200 {object} models.EventResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/events [get]
func (ec *EventController) GetAllEvents(c *fiber.Ctx) error {
	// Call fabric service to get all events
	events, err := ec.fabricService.GetAllEvents()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve events",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.EventResponse{
		Success: true,
		Message: "Events retrieved successfully",
		Data:    events,
	})
}

// GetActiveEvents retrieves only active events
// @Summary Get active events
// @Description Get a list of events that are currently active and within their date range
// @Tags events
// @Produce json
// @Success 200 {object} models.EventResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/events/active [get]
func (ec *EventController) GetActiveEvents(c *fiber.Ctx) error {
	// Call fabric service to get active events
	events, err := ec.fabricService.GetActiveEvents()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve active events",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.EventResponse{
		Success: true,
		Message: "Active events retrieved successfully",
		Data:    events,
	})
}

// UpdateEventStatus updates the active status of an event
// @Summary Update event status
// @Description Update the active status of an event (activate/deactivate)
// @Tags events
// @Accept json
// @Produce json
// @Param code path string true "Event Code"
// @Param status body models.UpdateEventRequest true "Event status update"
// @Success 200 {object} models.EventResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/events/{code}/status [put]
func (ec *EventController) UpdateEventStatus(c *fiber.Ctx) error {
	eventCode := c.Params("code")
	var req models.UpdateEventRequest

	if eventCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event code is required",
		})
	}

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Validate isActive field
	if req.IsActive == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "isActive field is required",
		})
	}

	if req.IsActive != "true" && req.IsActive != "false" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "isActive must be 'true' or 'false'",
		})
	}

	// Call fabric service to update event status
	event, err := ec.fabricService.UpdateEventStatus(eventCode, req.IsActive)
	if err != nil {
		if err.Error() == "event not found" {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Success: false,
				Message: "Event not found",
				Error:   err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to update event status",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.EventResponse{
		Success: true,
		Message: "Event status updated successfully",
		Data:    event,
	})
}

// UpdateEvent godoc
// @Summary Update event details
// @Description Update event details including name, dates, and status
// @Tags events
// @Accept json
// @Produce json
// @Param code path string true "Event Code"
// @Param event body models.UpdateEventDetailRequest true "Event update data"
// @Success 200 {object} models.EventResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/events/{code} [put]
func (ec *EventController) UpdateEvent(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event code is required",
		})
	}

	var req models.UpdateEventDetailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	// Validate request
	if err := ec.validator.Struct(req); err != nil {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Validation error: " + err.Error(),
		})
	}

	// Call fabric service
	updatedEvent, err := ec.fabricService.UpdateEvent(code, req)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return c.Status(404).JSON(models.ErrorResponse{
				Success: false,
				Message: "Event not found",
			})
		}
		return c.Status(500).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to update event: " + err.Error(),
		})
	}

	return c.JSON(models.EventResponse{
		Success: true,
		Message: "Event updated successfully",
		Data:    updatedEvent,
	})
}
