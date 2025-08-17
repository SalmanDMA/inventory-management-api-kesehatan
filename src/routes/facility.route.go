package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func FacilityRoutes(r fiber.Router) {
	facilitiesGroup := r.Group("/facility")
	facilitiesGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	facilitiesGroup.Get("/", controllers.FacilityControllerGetAll)
	facilitiesGroup.Post("/", controllers.FacilityControllerCreate)
	facilitiesGroup.Put("/restore", controllers.FacilityControllerRestore)
	facilitiesGroup.Delete("/delete", controllers.FacilityControllerDelete)
	facilitiesGroup.Get("/:id", controllers.FacilityControllerGetByID)
	facilitiesGroup.Put("/:id", controllers.FacilityControllerUpdate)
	} 