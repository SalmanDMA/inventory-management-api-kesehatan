package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func CustomerTypeRoutes(r fiber.Router) {
customerTypesGroup := r.Group("/customer-type")
customerTypesGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
customerTypesGroup.Get("/", controllers.CustomerTypeControllerGetAll)
customerTypesGroup.Post("/", controllers.CustomerTypeControllerCreate)
customerTypesGroup.Put("/restore", controllers.CustomerTypeControllerRestore)
customerTypesGroup.Delete("/delete", controllers.CustomerTypeControllerDelete)
customerTypesGroup.Get("/:id", controllers.CustomerTypeControllerGetByID)
customerTypesGroup.Put("/:id", controllers.CustomerTypeControllerUpdate)
} 