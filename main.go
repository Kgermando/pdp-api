package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/pdp-api/config"
	"github.com/kgermando/pdp-api/database"
	"github.com/kgermando/pdp-api/middleware"
	"github.com/kgermando/pdp-api/routes"
)

func main() {
	// Load configuration
	config.Load()

	// Connect to database
	database.Connect()
	database.Migrate()
	database.SeedSuperAdmin()

	// Ensure upload directory exists
	if err := os.MkdirAll(config.AppConfig.UploadDir, 0750); err != nil {
		log.Printf("Warning: Could not create upload dir: %v", err)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Parlement Digital du Peuple API v1.0",
		BodyLimit:    int(config.AppConfig.MaxUpload),
		ErrorHandler: customErrorHandler,
	})

	// Setup global middlewares
	middleware.SetupMiddlewares(app)

	// Setup routes
	routes.Setup(app)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Route non trouvée",
		})
	})

	addr := fmt.Sprintf(":%s", config.AppConfig.AppPort)
	log.Printf("🚀 Parlement Digital du Peuple API démarré sur %s", addr)
	log.Fatal(app.Listen(addr))
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Erreur interne du serveur"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}
