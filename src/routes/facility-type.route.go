package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func FacilityTypeRoutes(r fiber.Router) {
	facilityTypesGroup := r.Group("/facility-type")
	facilityTypesGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	facilityTypesGroup.Get("/", controllers.FacilityTypeControllerGetAll)
	facilityTypesGroup.Post("/", controllers.FacilityTypeControllerCreate)
	facilityTypesGroup.Put("/restore", controllers.FacilityTypeControllerRestore)
	facilityTypesGroup.Delete("/delete", controllers.FacilityTypeControllerDelete)
	facilityTypesGroup.Get("/:id", controllers.FacilityTypeControllerGetByID)
	facilityTypesGroup.Put("/:id", controllers.FacilityTypeControllerUpdate)
} 