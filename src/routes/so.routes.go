package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SalesOrderRoutes(r fiber.Router) {
	public := r.Group("/sales-order")
	public.Get("/:id/document", controllers.GenerateDocumentDeliveryOrder)
	public.Get("/:id/invoice", controllers.GenerateInvoice)
	public.Get("/:id/receipt", controllers.GenerateReceipt)

	protected := r.Group("/sales-order", middlewares.JWTProtected, middlewares.RBACMiddleware)
	protected.Get("/", controllers.GetAllSalesOrdersPaginated)
	protected.Post("/", controllers.CreateSalesOrder)

	protected.Get("/all", controllers.GetAllSalesOrders)
	protected.Delete("/delete", controllers.DeleteSalesOrders)
	protected.Put("/restore", controllers.RestoreSalesOrders)

	protected.Get("/:id", controllers.GetSalesOrderByID)
	protected.Put("/:id", controllers.UpdateSalesOrder)
	protected.Put("/:id/status", controllers.UpdateSalesOrderStatus)
}
