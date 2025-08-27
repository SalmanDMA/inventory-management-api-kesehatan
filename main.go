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
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env:", err)
	}

	// (Opsional) pastikan timezone Asia/Jakarta untuk job yang pakai time.Now()
	if tz := os.Getenv("APP_TZ"); tz == "" {
		_ = os.Setenv("APP_TZ", "Asia/Jakarta")
	}
	loc, err := time.LoadLocation(os.Getenv("APP_TZ"))
	if err != nil {
		log.Printf("⚠️  Failed load location %q: %v (fallback local)", os.Getenv("APP_TZ"), err)
	} else {
		time.Local = loc
	}

	// Minio, DB, migration
	configs.InitMinio()
	configs.ConnectDB()
	migrations.RunMigration()

	// ====== Start background jobs here ======
	if os.Getenv("DISABLE_JOBS") != "1" {
		jobs.StartAll() 
		log.Println("🕒 Background jobs started")
	} else {
		log.Println("⏸  Background jobs disabled by env")
	}
	// ========================================

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

	// Graceful shutdown (CTRL+C / SIGTERM)
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🔻 Shutting down server...")

	// Tunggu beberapa detik bila perlu untuk job yang lagi jalan (opsional)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = app.Shutdown() 

	<-ctx.Done()
	log.Println("👋 Bye")
}
