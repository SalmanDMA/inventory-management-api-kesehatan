package routes

import (
	"github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func DashboardRoutes(r fiber.Router) {
	dashboardsGroup := r.Group("/dashboard")
	dashboardsGroup.Use(middlewares.JWTProtected)
	dashboardsGroup.Get("/summary", controllers.DashboardControllerGetSummary)
}