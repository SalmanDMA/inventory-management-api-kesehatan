package routes

import (
	controllers "github.com/SalmanDMA/inventory-app/backend/src/controllers"
	"github.com/SalmanDMA/inventory-app/backend/src/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SalesPersonRoutes(r fiber.Router) {
	salesPersonsGroup := r.Group("/sales-person")
	salesPersonsGroup.Use(middlewares.JWTProtected, middlewares.RBACMiddleware)
	salesPersonsGroup.Get("/", controllers.SalesPersonControllerGetAll)
	salesPersonsGroup.Post("/", controllers.SalesPersonControllerCreate)
	salesPersonsGroup.Put("/restore", controllers.SalesPersonControllerRestore)
	salesPersonsGroup.Delete("/delete", controllers.SalesPersonControllerDelete)
	salesPersonsGroup.Get("/:id", controllers.SalesPersonControllerGetById)
	salesPersonsGroup.Put("/:id", controllers.SalesPersonControllerUpdate)

		// sales assignment 
	salesPersonsGroup.Get("/:salesPersonId/area", controllers.GetAreasWithSalesPersonInfo)
	salesPersonsGroup.Post("/:salesPersonId/area", controllers.CreateOrUpdateSalesAssignment)
} 