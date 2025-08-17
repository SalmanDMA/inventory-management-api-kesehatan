package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SupplierRoutes(r fiber.Router) {
	suppliers := r.Group("/supplier")
	suppliers.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	suppliers.Get("/", controllers.GetAllSuppliersPaginated)
	suppliers.Post("/", controllers.CreateSupplier)
	
	suppliers.Get("/all", controllers.GetAllSuppliers)
	suppliers.Delete("/delete", controllers.DeleteSuppliers)
	suppliers.Put("/restore", controllers.RestoreSuppliers)

	suppliers.Get("/:id", controllers.GetSupplierByID)
	suppliers.Put("/:id", controllers.UpdateSupplier)
}

