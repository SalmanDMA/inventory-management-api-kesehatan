package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func PaymentRoutes(r fiber.Router) {
	payments := r.Group("/payment")
	payments.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	payments.Post("/", controllers.CreatePayment)
	payments.Get("/purchase-order/:po_id", controllers.GetPaymentsByPurchaseOrder)
	payments.Get("/sales-order/:so_id", controllers.GetPaymentsBySalesOrder)
}

