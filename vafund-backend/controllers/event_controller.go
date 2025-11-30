package controllers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	assetsPath    string
}

func NewEventController(fabricService *services.FabricService) *EventController {
	// Path to assets folder inside container
	assetsPath := "/app/assets"

	// Create assets directory if it doesn't exist
	if err := os.MkdirAll(assetsPath, 0755); err != nil {
		fmt.Printf("Warning: Could not create assets directory: %v\n", err)
	}

	return &EventController{
		fabricService: fabricService,
		validator:     validator.New(),
		assetsPath:    assetsPath,
	}
}

// CreateEvent creates a new event
// @Summary Create a new event
// @Description Create a new event with code, name, start date, end date, active status and image
// @Tags events
// @Accept multipart/form-data
// @Produce json
// @Param code formData string true "Event Code"
// @Param name formData string true "Event Name"
// @Param description formData string true "Event Description"
// @Param startDate formData string true "Start Date (RFC3339 format)"
// @Param endDate formData string true "End Date (RFC3339 format)"
// @Param isActive formData string true "Is Active (true/false)"
// @Param image formData file false "Event Image"
// @Success 201 {object} models.EventResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/events [post]
func (ec *EventController) CreateEvent(c *fiber.Ctx) error {
	// Get form data
	code := c.FormValue("code")
	name := c.FormValue("name")
	description := c.FormValue("description")
	startDate := c.FormValue("startDate")
	endDate := c.FormValue("endDate")
	isActive := c.FormValue("isActive")

	// Validate required fields
	if code == "" || name == "" || startDate == "" || endDate == "" || isActive == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Missing required fields: code, name, startDate, endDate, isActive",
		})
	}

	// Validate date formats
	_, err := time.Parse("2006-01-02T15:04:05Z07:00", startDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid start date format. Use RFC3339 format (e.g., 2025-01-01T00:00:00Z)",
			Error:   err.Error(),
		})
	}

	endDateTime, err := time.Parse("2006-01-02T15:04:05Z07:00", endDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Invalid end date format. Use RFC3339 format (e.g., 2025-12-31T23:59:59Z)",
			Error:   err.Error(),
		})
	}

	startDateTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", startDate)

	// Validate date range
	if endDateTime.Before(startDateTime) || endDateTime.Equal(startDateTime) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "End date must be after start date",
		})
	}

	// Validate isActive
	if isActive != "true" && isActive != "false" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "isActive must be 'true' or 'false'",
		})
	}

	imageURL := ""

	// Handle optional image upload
	file, err := c.FormFile("image")
	if err == nil {
		// Validate file type
		allowedTypes := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".gif":  true,
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedTypes[ext] {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Success: false,
				Message: "Invalid file type. Allowed: jpg, jpeg, png, gif",
			})
		}

		// Validate file size (5MB max)
		if file.Size > 5*1024*1024 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Success: false,
				Message: "File size exceeds 5MB limit",
			})
		}

		// Generate filename: event_<code>.<ext>
		filename := fmt.Sprintf("event_%s%s", code, ext)
		filePath := filepath.Join(ec.assetsPath, filename)

		// Save file
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Success: false,
				Message: "Failed to save image file",
				Error:   err.Error(),
			})
		}

		// Image URL will be relative path
		imageURL = fmt.Sprintf("/api/events/%s/image", code)
	}

	// Create request object
	req := models.CreateEventRequest{
		Code:        code,
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    isActive,
	}

	// Call fabric service to create event
	event, err := ec.fabricService.CreateEvent(req, imageURL)
	if err != nil {
		// Delete uploaded file if blockchain transaction fails
		if imageURL != "" {
			pattern := fmt.Sprintf("event_%s.*", code)
			matches, _ := filepath.Glob(filepath.Join(ec.assetsPath, pattern))
			for _, file := range matches {
				os.Remove(file)
			}
		}
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
// @Description Update event details including name, dates, status and optional image
// @Tags events
// @Accept multipart/form-data
// @Produce json
// @Param code path string true "Event Code"
// @Param name formData string true "Event Name"
// @Param description formData string true "Event Description"
// @Param startDate formData string true "Start Date (RFC3339 format)"
// @Param endDate formData string true "End Date (RFC3339 format)"
// @Param isActive formData string true "Is Active (true/false)"
// @Param image formData file false "New Event Image (optional)"
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

	// Get form data
	name := c.FormValue("name")
	description := c.FormValue("description")
	startDate := c.FormValue("startDate")
	endDate := c.FormValue("endDate")
	isActive := c.FormValue("isActive")

	// Validate required fields
	if name == "" || startDate == "" || endDate == "" || isActive == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Missing required fields: name, startDate, endDate, isActive",
		})
	}

	// Get existing event to get current imgUrl
	existingEvent, err := ec.fabricService.GetEvent(code)
	if err != nil {
		return c.Status(404).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event not found",
		})
	}

	imgUrl := existingEvent.ImgUrl

	// Handle optional image upload
	file, err := c.FormFile("image")
	if err == nil {
		// Validate file type
		allowedTypes := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".gif":  true,
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedTypes[ext] {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Success: false,
				Message: "Invalid file type. Allowed: jpg, jpeg, png, gif",
			})
		}

		// Validate file size
		if file.Size > 5*1024*1024 {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Success: false,
				Message: "File size exceeds 5MB limit",
			})
		}

		// Delete old image files
		oldPattern := fmt.Sprintf("event_%s.*", code)
		oldMatches, _ := filepath.Glob(filepath.Join(ec.assetsPath, oldPattern))
		for _, oldFile := range oldMatches {
			os.Remove(oldFile)
		}

		// Save new image
		filename := fmt.Sprintf("event_%s%s", code, ext)
		filePath := filepath.Join(ec.assetsPath, filename)

		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Success: false,
				Message: "Failed to save new image file",
				Error:   err.Error(),
			})
		}

		imgUrl = fmt.Sprintf("/api/events/%s/image", code)
	}

	// Create request object
	req := models.UpdateEventDetailRequest{
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    isActive,
	}

	// Validate request
	if err := ec.validator.Struct(req); err != nil {
		return c.Status(400).JSON(models.ErrorResponse{
			Success: false,
			Message: "Validation error: " + err.Error(),
		})
	}

	// Call fabric service
	updatedEvent, err := ec.fabricService.UpdateEvent(code, req, imgUrl)
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

// GetEventImage retrieves the image file for an event
// @Summary Get event image
// @Description Get the actual image file for an event
// @Tags events
// @Produce image/jpeg,image/png,image/gif
// @Param code path string true "Event Code"
// @Success 200 {file} binary
// @Failure 404 {object} models.ErrorResponse
// @Router /api/events/{code}/image [get]
func (ec *EventController) GetEventImage(c *fiber.Ctx) error {
	code := c.Params("code")

	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event code is required",
		})
	}

	// Get event data to verify it exists
	_, err := ec.fabricService.GetEvent(code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event not found",
			Error:   err.Error(),
		})
	}

	// Find the image file
	pattern := fmt.Sprintf("event_%s.*", code)
	matches, err := filepath.Glob(filepath.Join(ec.assetsPath, pattern))

	if err != nil || len(matches) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: "Image file not found",
		})
	}

	imagePath := matches[0]

	// Open and serve the file
	file, err := os.Open(imagePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to read image file",
			Error:   err.Error(),
		})
	}
	defer file.Close()

	// Get file info for content type
	ext := strings.ToLower(filepath.Ext(imagePath))
	contentType := "image/jpeg"
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	}

	c.Set("Content-Type", contentType)

	// Copy file to response
	_, err = io.Copy(c.Response().BodyWriter(), file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to send image",
			Error:   err.Error(),
		})
	}

	return nil
}
