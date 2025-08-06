package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func CategoryRoutes(r fiber.Router) {
	categoriesGroup := r.Group("/category")
	categoriesGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	categoriesGroup.Get("/", controllers.CategoryControllerGetAll)
	categoriesGroup.Post("/", controllers.CategoryControllerCreate)
	categoriesGroup.Put("/restore", controllers.CategoryControllerRestore)
	categoriesGroup.Delete("/delete", controllers.CategoryControllerDelete)
	categoriesGroup.Get("/:id", controllers.CategoryControllerGetByID)
	categoriesGroup.Put("/:id", controllers.CategoryControllerUpdate)
	} 