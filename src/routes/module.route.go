package routes

import (
	"github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func ModuleRoutes(r fiber.Router) {	
	modulesGroup := r.Group("/module")
	modulesGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	modulesGroup.Put("/restore", controllers.ModuleControllerRestore)
	modulesGroup.Delete("/delete", controllers.ModuleControllerDelete)
	modulesGroup.Get("/", controllers.ModuleControllerGetAll)
	// modulesGroup.Get("/root", controllers.ModuleControllerGetModuleRoot)
	modulesGroup.Post("/", controllers.ModuleControllerCreate)
	modulesGroup.Get("/:id", controllers.ModuleControllerGetById)
	modulesGroup.Put("/:id", controllers.ModuleControllerUpdate)
} 