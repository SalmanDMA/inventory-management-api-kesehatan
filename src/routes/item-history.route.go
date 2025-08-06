package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func ItemHistoryRoutes(r fiber.Router) {
	itemhistoriesGroup := r.Group("/item-history")
	itemhistoriesGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	itemhistoriesGroup.Get("/", controllers.ItemHistoryControllerGetAll)
	itemhistoriesGroup.Post("/", controllers.ItemHistoryControllerCreate)
	itemhistoriesGroup.Put("/restore", controllers.ItemHistoryControllerRestore)
	itemhistoriesGroup.Delete("/delete", controllers.ItemHistoryControllerDelete)
}