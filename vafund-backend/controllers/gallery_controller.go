package controllers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"vafund-backend/models"
	"vafund-backend/services"

	"github.com/gofiber/fiber/v2"
)

type GalleryController struct {
	fabricService *services.FabricService
	assetsPath    string
}

func NewGalleryController(fabricService *services.FabricService) *GalleryController {
	// Path to assets folder in smartcontract
	assetsPath := "../vafund-smartcontract/assets"
	
	// Create assets directory if it doesn't exist
	if err := os.MkdirAll(assetsPath, 0755); err != nil {
		fmt.Printf("Warning: Could not create assets directory: %v\n", err)
	}

	return &GalleryController{
		fabricService: fabricService,
		assetsPath:    assetsPath,
	}
}

// CreateGallery creates a new gallery entry with image upload
// @Summary Create a new gallery entry
// @Description Create a new gallery entry with image upload
// @Tags gallery
// @Accept multipart/form-data
// @Produce json
// @Param id formData string true "Gallery ID"
// @Param eventCode formData string true "Event Code"
// @Param description formData string false "Description"
// @Param image formData file true "Image file"
// @Success 201 {object} models.GalleryResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/gallery [post]
func (gc *GalleryController) CreateGallery(c *fiber.Ctx) error {
	// Get form data
	id := c.FormValue("id")
	eventCode := c.FormValue("eventCode")
	description := c.FormValue("description")

	// Validate required fields
	if id == "" || eventCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Missing required fields: id, eventCode",
		})
	}

	// Get uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Image file is required",
			Error:   err.Error(),
		})
	}

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

	// Generate filename: gallery_<eventCode>_<id>.<ext>
	filename := fmt.Sprintf("gallery_%s_%s%s", eventCode, id, ext)
	filePath := filepath.Join(gc.assetsPath, filename)

	// Save file
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to save image file",
			Error:   err.Error(),
		})
	}

	// Image URL will be relative path for blockchain storage
	imageURL := fmt.Sprintf("/api/gallery/%s/image", id)

	// Create gallery in blockchain
	gallery, err := gc.fabricService.CreateGallery(id, eventCode, imageURL, description)
	if err != nil {
		// Delete uploaded file if blockchain transaction fails
		os.Remove(filePath)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to create gallery in blockchain",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.GalleryResponse{
		Success: true,
		Message: "Gallery created successfully",
		Data:    gallery,
	})
}

// GetGallery retrieves a specific gallery by ID
// @Summary Get gallery by ID
// @Description Get a specific gallery by its ID
// @Tags gallery
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} models.GalleryResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/gallery/{id} [get]
func (gc *GalleryController) GetGallery(c *fiber.Ctx) error {
	galleryID := c.Params("id")

	if galleryID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Gallery ID is required",
		})
	}

	gallery, err := gc.fabricService.GetGallery(galleryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: "Gallery not found",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.GalleryResponse{
		Success: true,
		Message: "Gallery retrieved successfully",
		Data:    gallery,
	})
}

// GetAllGalleries retrieves all galleries
// @Summary Get all galleries
// @Description Get all galleries from the blockchain
// @Tags gallery
// @Produce json
// @Success 200 {object} models.GalleryResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/gallery [get]
func (gc *GalleryController) GetAllGalleries(c *fiber.Ctx) error {
	galleries, err := gc.fabricService.GetAllGalleries()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve galleries",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.GalleryResponse{
		Success: true,
		Message: "Galleries retrieved successfully",
		Data:    galleries,
	})
}

// GetGalleriesByEventCode retrieves galleries by event code
// @Summary Get galleries by event code
// @Description Get all galleries associated with a specific event
// @Tags gallery
// @Produce json
// @Param eventCode path string true "Event Code"
// @Success 200 {object} models.GalleryResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/gallery/event/{eventCode} [get]
func (gc *GalleryController) GetGalleriesByEventCode(c *fiber.Ctx) error {
	eventCode := c.Params("eventCode")

	if eventCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Event code is required",
		})
	}

	galleries, err := gc.fabricService.GetGalleriesByEventCode(eventCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve galleries for event",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.GalleryResponse{
		Success: true,
		Message: fmt.Sprintf("Galleries for event %s retrieved successfully", eventCode),
		Data:    galleries,
	})
}

// GetGalleryImage retrieves the image file for a gallery
// @Summary Get gallery image
// @Description Get the actual image file for a gallery entry
// @Tags gallery
// @Produce image/jpeg,image/png,image/gif
// @Param id path string true "Gallery ID"
// @Success 200 {file} binary
// @Failure 404 {object} models.ErrorResponse
// @Router /api/gallery/{id}/image [get]
func (gc *GalleryController) GetGalleryImage(c *fiber.Ctx) error {
	galleryID := c.Params("id")

	if galleryID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Gallery ID is required",
		})
	}

	// Get gallery data to find event code
	gallery, err := gc.fabricService.GetGallery(galleryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: "Gallery not found",
			Error:   err.Error(),
		})
	}

	// Find the image file
	patterns := []string{
		fmt.Sprintf("gallery_%s_%s.*", gallery.EventCode, galleryID),
	}

	var imagePath string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(gc.assetsPath, pattern))
		if err == nil && len(matches) > 0 {
			imagePath = matches[0]
			break
		}
	}

	if imagePath == "" {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: "Image file not found",
		})
	}

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

// UpdateGallery updates an existing gallery entry
// @Summary Update gallery entry
// @Description Update an existing gallery entry with optional new image
// @Tags gallery
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Gallery ID"
// @Param eventCode formData string true "Event Code"
// @Param description formData string false "Description"
// @Param image formData file false "New image file (optional)"
// @Success 200 {object} models.GalleryResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/gallery/{id} [put]
func (gc *GalleryController) UpdateGallery(c *fiber.Ctx) error {
	galleryID := c.Params("id")
	eventCode := c.FormValue("eventCode")
	description := c.FormValue("description")

	if galleryID == "" || eventCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Missing required fields: id, eventCode",
		})
	}

	// Get existing gallery
	existingGallery, err := gc.fabricService.GetGallery(galleryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: "Gallery not found",
			Error:   err.Error(),
		})
	}

	imageURL := existingGallery.ImageURL

	// Check if new image is uploaded
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
		oldPattern := fmt.Sprintf("gallery_%s_%s.*", existingGallery.EventCode, galleryID)
		oldMatches, _ := filepath.Glob(filepath.Join(gc.assetsPath, oldPattern))
		for _, oldFile := range oldMatches {
			os.Remove(oldFile)
		}

		// Save new image
		filename := fmt.Sprintf("gallery_%s_%s%s", eventCode, galleryID, ext)
		filePath := filepath.Join(gc.assetsPath, filename)

		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Success: false,
				Message: "Failed to save new image file",
				Error:   err.Error(),
			})
		}

		imageURL = fmt.Sprintf("/api/gallery/%s/image", galleryID)
	}

	// Update gallery in blockchain
	gallery, err := gc.fabricService.UpdateGallery(galleryID, eventCode, imageURL, description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to update gallery in blockchain",
			Error:   err.Error(),
		})
	}

	return c.JSON(models.GalleryResponse{
		Success: true,
		Message: "Gallery updated successfully",
		Data:    gallery,
	})
}

// DeleteGallery deletes a gallery entry
// @Summary Delete gallery entry
// @Description Delete a gallery entry and its associated image
// @Tags gallery
// @Produce json
// @Param id path string true "Gallery ID"
// @Success 200 {object} models.GalleryResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/gallery/{id} [delete]
func (gc *GalleryController) DeleteGallery(c *fiber.Ctx) error {
	galleryID := c.Params("id")

	if galleryID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Success: false,
			Message: "Gallery ID is required",
		})
	}

	// Get gallery to find event code
	gallery, err := gc.fabricService.GetGallery(galleryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Success: false,
			Message: "Gallery not found",
			Error:   err.Error(),
		})
	}

	// Delete from blockchain
	err = gc.fabricService.DeleteGallery(galleryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Success: false,
			Message: "Failed to delete gallery from blockchain",
			Error:   err.Error(),
		})
	}

	// Delete image files
	pattern := fmt.Sprintf("gallery_%s_%s.*", gallery.EventCode, galleryID)
	matches, _ := filepath.Glob(filepath.Join(gc.assetsPath, pattern))
	for _, filePath := range matches {
		os.Remove(filePath)
	}

	return c.JSON(models.GalleryResponse{
		Success: true,
		Message: "Gallery deleted successfully",
	})
}
