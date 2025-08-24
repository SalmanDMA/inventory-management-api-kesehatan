package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func UoMRoutes(r fiber.Router) {
	uomsGroup := r.Group("/uom")
	uomsGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	uomsGroup.Get("/", controllers.UoMControllerGetAll)
	uomsGroup.Post("/", controllers.UoMControllerCreate)
	uomsGroup.Put("/restore", controllers.UoMControllerRestore)
	uomsGroup.Delete("/delete", controllers.UoMControllerDelete)
	uomsGroup.Get("/:id", controllers.UoMControllerGetByID)
	uomsGroup.Put("/:id", controllers.UoMControllerUpdate)
} 