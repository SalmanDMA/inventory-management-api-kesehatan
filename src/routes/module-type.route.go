package routes

import (
	"github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func ModuleTypeRoutes(r fiber.Router) {	
	moduleTypessGroup := r.Group("/module-type")
	moduleTypessGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	moduleTypessGroup.Get("/", controllers.ModuleTypeControllerGetAll)
	moduleTypessGroup.Post("/", controllers.ModuleTypeControllerCreate)
	moduleTypessGroup.Put("/restore", controllers.ModuleTypeControllerRestore)
	moduleTypessGroup.Delete("/delete", controllers.ModuleTypeControllerDelete)

	moduleTypessGroup.Get("/:id", controllers.ModuleTypeControllerGetByID)
	moduleTypessGroup.Put("/:id", controllers.ModuleTypeControllerUpdate)
} 