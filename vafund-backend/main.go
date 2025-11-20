// @title VaFund Backend API
// @version 1.0
// @description Backend API untuk aplikasi VaFund menggunakan Go Fiber dan Hyperledger Fabric
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@vafund.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:3000
// @BasePath /
// @schemes http
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"vafund-backend/config"
	"vafund-backend/controllers"
	_ "vafund-backend/docs"
	"vafund-backend/routes"
	"vafund-backend/services"
)

func main() {
	cfg := config.LoadConfig()

	app := fiber.New(fiber.Config{
		AppName:      "VaFund Backend API",
		ServerHeader: "VaFund",
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	fabricService := services.NewFabricService()

	// Try to connect to Fabric network
	if err := fabricService.Connect(); err != nil {
		log.Printf("⚠️  Warning: Failed to connect to Fabric network: %v", err)
		log.Println("💡 Running in simulation mode. To connect to real Fabric:")
		log.Println("   1. Make sure Fabric network is running in Docker")
		log.Println("   2. Copy crypto materials to ./crypto/ folder")
		log.Println("   3. Check crypto/README.md for instructions")
	} else {
		log.Println("✅ Successfully connected to Fabric network!")
		// Ensure graceful shutdown
		defer fabricService.Close()
	}

	donationController := controllers.NewDonationController(fabricService)
	eventController := controllers.NewEventController(fabricService)
	withdrawalController := controllers.NewWithdrawalController()

	routes.SetupRoutes(app, donationController, eventController, withdrawalController)

	log.Printf("🚀 VaFund Backend API starting on port %s", cfg.Port)
	log.Printf("🌐 Access the API at: http://localhost:%s", cfg.Port)
	log.Printf("📚 Swagger UI: http://localhost:%s/swagger/", cfg.Port)
	log.Printf("❤️  Health check: http://localhost:%s/api/health", cfg.Port)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
