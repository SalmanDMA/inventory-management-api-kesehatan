package main

import (
	"log"
	"os"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/migrations"
	"github.com/SalmanDMA/inventory-app/backend/src/routes"
	"github.com/SalmanDMA/inventory-app/backend/src/websockets"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if not in PRODUCTION
	if os.Getenv("ENVIRONMENT") != "PRODUCTION" {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatal("Error loading .env:", err)
		}
	}

	// Database connection and migration
	configs.ConnectDB()
	migrations.RunMigration()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB
	})

	// Serve static files
	app.Static("/uploads", "./public/uploads")

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ALLOWED_ORIGIN"),
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: true,
	}))
	app.Use(recover.New())

	// Routes
	routes.RouteInit(app)

	// Register WebSocket route
	websockets.Register(app)

	// Get port from env or fallback
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Start server
	log.Printf("✅ Server running at http://localhost:%s", port)
	log.Printf("📡 WebSocket server available at ws://localhost:%s/ws", port)
	
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}