package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SalesReportRoutes(r fiber.Router) {
	// Public endpoints
	public := r.Group("/sales-report")
	public.Get("/excel", controllers.ExportSalesReportExcel)
	public.Get("/:id/pdf", controllers.ExportSalesReportPDF)

	// Protected endpoints (middleware cukup sekali di sini)
	protected := r.Group("/sales-report", middlewares.JWTProtected, middlewares.RBACMiddleware)
	protected.Get("/summary", controllers.GetSalesReportSummary)
	protected.Get("/charts", controllers.GetSalesReportCharts)
	protected.Get("/details", controllers.GetSalesReportDetails)
	protected.Get("/insights", controllers.GetSalesReportInsights)
}
