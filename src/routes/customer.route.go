package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func CustomerRoutes(r fiber.Router) {
customersGroup := r.Group("/customer")
customersGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
customersGroup.Get("/", controllers.CustomerControllerGetAll)
customersGroup.Post("/", controllers.CustomerControllerCreate)
customersGroup.Put("/restore", controllers.CustomerControllerRestore)
customersGroup.Delete("/delete", controllers.CustomerControllerDelete)
customersGroup.Get("/:id", controllers.CustomerControllerGetByID)
customersGroup.Put("/:id", controllers.CustomerControllerUpdate)
	} 