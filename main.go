package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/jobs"
	"github.com/SalmanDMA/inventory-app/backend/src/migrations"
	"github.com/SalmanDMA/inventory-app/backend/src/routes"
	"github.com/SalmanDMA/inventory-app/backend/src/websockets"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if _, err := os.Stat(".env"); err == nil {
    if err := godotenv.Load(".env"); err != nil {
        log.Printf("⚠️ Failed to load .env: %v (ignored, using system env)", err)
    } else {
        log.Println("📄 .env file loaded")
    }
	} else {
					log.Println("ℹ️ .env file not found, using system environment variables")
	}

	// Setup timezone
	tz := os.Getenv("APP_TZ")
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("⚠️ Failed load location %q: %v, fallback to WIB fixed offset", tz, err)
		loc = time.FixedZone("WIB", 7*60*60) // UTC+7
	}
	time.Local = loc // set global
	log.Printf("🌍 Timezone set to %s", loc)

	// Init services
	configs.InitMinio()
	configs.ConnectDB()
	migrations.RunMigration()

	// Start background jobs
	if os.Getenv("DISABLE_JOBS") != "1" {
		jobs.StartAll(loc)
		log.Println("🕒 Background jobs started")
	} else {
		log.Println("⏸ Background jobs disabled by env")
	}

	// Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB
	})
	app.Static("/uploads", "./public/uploads")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("ALLOWED_ORIGIN"),
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: true,
	}))
	app.Use(recover.New())

	// Routes & websockets
	routes.RouteInit(app)
	websockets.Register(app)

	// Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("✅ Server running at http://localhost:%s", port)
	log.Printf("📡 WebSocket server available at ws://localhost:%s/ws", port)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🔻 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = app.Shutdown()

	<-ctx.Done()
	log.Println("👋 Bye")
}
