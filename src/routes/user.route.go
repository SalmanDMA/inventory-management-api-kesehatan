package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(r fiber.Router) {
	usersGroup := r.Group("/user")

	usersGroup.Use(middlewares.JWTProtected)
	usersGroup.Get("/me", controllers.UserControllerGetProfile)
	usersGroup.Put("/me", controllers.UserControllerUpdateProfile)
	usersGroup.Put("/me/avatar", controllers.UserControllerUploadAvatar)

	protected := usersGroup.Group("/")
	protected.Use(middlewares.RBACMiddleware)
	protected.Put("/restore", controllers.UserControllerRestore)
	protected.Delete("/delete", controllers.UserControllerDelete)
	protected.Get("/", controllers.UserControllerGetAll)
	protected.Post("/", controllers.UserControllerCreate)
	protected.Get("/:id", controllers.UserControllerGetById)
	protected.Put("/:id", controllers.UserControllerUpdate)
}
