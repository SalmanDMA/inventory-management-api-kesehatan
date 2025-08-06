package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func ItemRoutes(r fiber.Router) {
	itemsGroup := r.Group("/item")
	itemsGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	itemsGroup.Get("/", controllers.ItemControllerGetAll)
	itemsGroup.Post("/", controllers.ItemControllerCreate)
	itemsGroup.Put("/restore", controllers.ItemControllerRestore)
	itemsGroup.Delete("/delete", controllers.ItemControllerDelete)
	itemsGroup.Get("/:id", controllers.ItemControllerGetByID)
	itemsGroup.Put("/:id", controllers.ItemControllerUpdate)
}