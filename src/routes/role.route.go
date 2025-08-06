package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func RoleRoutes(r fiber.Router) {
	rolesGroup := r.Group("/role")
	rolesGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	rolesGroup.Get("/", controllers.RoleControllerGetAll)
	rolesGroup.Post("/", controllers.RoleControllerCreate)
	rolesGroup.Put("/restore", controllers.RoleControllerRestore)
	rolesGroup.Delete("/delete", controllers.RoleControllerDelete)
	rolesGroup.Get("/:id", controllers.RoleControllerGetByID)
	rolesGroup.Put("/:id", controllers.RoleControllerUpdate)

	// role module
	rolesGroup.Get("/:roleId/module", controllers.GetModulesWithRoleInfo)
	rolesGroup.Post("/:roleId/module", controllers.CreateOrUpdateRoleModule)
	} 