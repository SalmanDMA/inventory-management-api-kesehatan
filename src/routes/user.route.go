package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(r fiber.Router) {
	usersGroup := r.Group("/user")
	usersGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	usersGroup.Get("/me", controllers.UserControllerGetProfile)
	usersGroup.Put("/me", controllers.UserControllerUpdateProfile)
	usersGroup.Put("/restore", controllers.UserControllerRestore)
	usersGroup.Delete("/delete", controllers.UserControllerDelete)
	usersGroup.Get("/", controllers.UserControllerGetAll)
	usersGroup.Post("/", controllers.UserControllerCreate)
	usersGroup.Get("/:id", controllers.UserControllerGetById)
	usersGroup.Put("/:id", controllers.UserControllerUpdate)
}
