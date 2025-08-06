package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get dashboard summary
// @Description Get aggregated dashboard data like total items, orders, value, etc.
// @Tags dashboards
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} models.DashboardSummary
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/dashboard/summary [get]
func DashboardControllerGetSummary(c *fiber.Ctx) error {
	dashboardService := services.NewDashboardService(
		repositories.NewItemRepository(configs.DB),
	)

	summary, err := dashboardService.GetDashboardSummary()
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Failed to fetch dashboard summary", nil)
	}

	return helpers.Response(c, fiber.StatusOK, "Dashboard summary fetched successfully", summary)
}