package routes

import (
	"vafund-backend/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func SetupRoutes(app *fiber.App, donationController *controllers.DonationController, eventController *controllers.EventController, withdrawalController *controllers.WithdrawalController, galleryController *controllers.GalleryController) {
	// Swagger endpoint
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api")

	// Donation routes
	donations := api.Group("/donations")
	donations.Post("/", donationController.CreateDonation)
	donations.Get("/", donationController.GetAllDonations)
	donations.Get("/total", donationController.GetTotalDonations)
	donations.Get("/by-amount", donationController.GetDonationsByAmount)
	donations.Get("/event/:eventCode", donationController.GetDonationsByEventCode)
	donations.Get("/event/:eventCode/total", donationController.GetTotalDonationsByEventCode)
	donations.Get("/:id", donationController.GetDonation)

	// Event routes
	events := api.Group("/events")
	events.Post("/", eventController.CreateEvent)
	events.Get("/", eventController.GetAllEvents)
	events.Get("/active", eventController.GetActiveEvents)
	events.Get("/:code", eventController.GetEvent)
	events.Get("/:code/image", eventController.GetEventImage)
	events.Put("/:code/status", eventController.UpdateEventStatus)
	events.Put("/:code", eventController.UpdateEvent)

	// Withdrawal routes
	withdrawals := api.Group("/withdrawals")
	withdrawals.Post("/", withdrawalController.CreateWithdrawal)
	withdrawals.Get("/", withdrawalController.GetAllWithdrawals)
	withdrawals.Get("/total", withdrawalController.GetTotalWithdrawals)
	withdrawals.Get("/event/:event_code", withdrawalController.GetWithdrawalsByEventCode)
	withdrawals.Get("/event/:event_code/total", withdrawalController.GetTotalWithdrawalsByEventCode)
	withdrawals.Get("/:id", withdrawalController.GetWithdrawal)

	// Gallery routes
	gallery := api.Group("/gallery")
	gallery.Post("/", galleryController.CreateGallery)
	gallery.Get("/", galleryController.GetAllGalleries)
	gallery.Get("/event/:eventCode", galleryController.GetGalleriesByEventCode)
	gallery.Get("/:id", galleryController.GetGallery)
	gallery.Get("/:id/image", galleryController.GetGalleryImage)
	gallery.Put("/:id", galleryController.UpdateGallery)
	gallery.Delete("/:id", galleryController.DeleteGallery)

	// Health check endpoint
	// @Summary Health Check
	// @Description Check if the API is running
	// @Tags health
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /api/health [get]
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "VaFund Backend API is running",
		})
	})
}
