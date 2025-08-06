package routes

import (
	"github.com/SalmanDMA/inventory-app/backend/src/controllers"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
)

func RouteInit(app *fiber.App) {
	api := app.Group("/api")

	v1 := api.Group("/v1", func(c *fiber.Ctx) error {
					c.Set("Version", "v1")
					return c.Next()
	})

	v1.Get("/healthCheck", HealthCheck)
	v1.Get("/swagger/*", swagger.HandlerDefault)

	v1.Get("/upload", controllers.UserControllerUpdateProfile)

	DashboardRoutes(v1)
	UserRoutes(v1)
	AuthRoutes(v1)
	RoleRoutes(v1)
	ModuleRoutes(v1)
	ModuleTypeRoutes(v1)
	CategoryRoutes(v1)
	ItemRoutes(v1)
	ItemHistoryRoutes(v1)
}


// HealthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/healthCheck [get]
func HealthCheck(c *fiber.Ctx) error {
	res := map[string]string{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
}

