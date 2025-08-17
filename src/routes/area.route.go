package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func AreaRoutes(r fiber.Router) {
	areasGroup := r.Group("/area")
	areasGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	areasGroup.Get("/", controllers.AreaControllerGetAll)
	areasGroup.Post("/", controllers.AreaControllerCreate)
	areasGroup.Put("/restore", controllers.AreaControllerRestore)
	areasGroup.Delete("/delete", controllers.AreaControllerDelete)
	areasGroup.Get("/:id", controllers.AreaControllerGetByID)
	areasGroup.Put("/:id", controllers.AreaControllerUpdate)
	} 