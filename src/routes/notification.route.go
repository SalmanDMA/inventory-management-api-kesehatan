package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func NotificationRoutes(r fiber.Router) {
	NotificationsGroup := r.Group("/notification")
	NotificationsGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	NotificationsGroup.Get("/", controllers.NotificationControllerGetAll)
	NotificationsGroup.Put("/mark-all-read", controllers.NotificationControllerMarkAllAsRead)
	NotificationsGroup.Put("/mark-multiple-read", controllers.NotificationControllerMarkMultipleAsRead)
	NotificationsGroup.Put("/restore", controllers.NotificationControllerRestore)
	NotificationsGroup.Delete("/delete", controllers.NotificationControllerDelete)
	NotificationsGroup.Put("/:id/read", controllers.NotificationControllerMarkAsRead)
	NotificationsGroup.Get("/:id", controllers.NotificationControllerGetByID)
} 