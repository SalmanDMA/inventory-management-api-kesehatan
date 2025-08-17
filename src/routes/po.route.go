package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func PurchaseOrderRoutes(r fiber.Router) {
	public := r.Group("/purchase-order")
	public.Get("/:id/document", controllers.GenerateDocumentPurchaseOrder)

	protected := r.Group("/purchase-order", middlewares.JWTProtected, middlewares.RBACMiddleware)
	protected.Get("/", controllers.GetAllPurchaseOrdersPaginated)
	protected.Post("/", controllers.CreatePurchaseOrder)

	protected.Get("/all", controllers.GetAllPurchaseOrders)
	protected.Delete("/delete", controllers.DeletePurchaseOrders)
	protected.Put("/restore", controllers.RestorePurchaseOrders)

	protected.Get("/:id", controllers.GetPurchaseOrderByID)
	protected.Put("/:id", controllers.UpdatePurchaseOrder)
	protected.Put("/:id/status", controllers.UpdatePurchaseOrderStatus)
	protected.Put("/:id/receive", controllers.ReceiveItems)
}
