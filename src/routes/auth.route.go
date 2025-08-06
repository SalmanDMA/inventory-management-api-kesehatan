package routes

import (
	"github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(r fiber.Router) {
	r.Post("auth/check-identifier", controllers.CheckIdentifierController)
	r.Post("auth/forgot-password", controllers.ForgotPasswordController)

	authGroup := r.Group("/auth")
	authGroup.Post("/login", controllers.LoginController)

	// using middleware
	authGroup.Use(middlewares.JWTProtected)
	authGroup.Put("/reset-password", controllers.ResetPasswordController)
}
