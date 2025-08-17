package controllers

import (
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetAreasWithSalesPersonInfo(ctx *fiber.Ctx) error {
	_, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	salesPersonIDStr := ctx.Params("salesPersonId")
	salesPersonID, err := uuid.Parse(salesPersonIDStr)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid sales person ID", nil)
	}

	areaRepo := repositories.NewAreaRepository(configs.DB)
	salesAssignmentRepo := repositories.NewSalesAssignmentRepository(configs.DB)
	salesAssignmentService := services.NewSalesAssignmentService(salesAssignmentRepo, areaRepo)
	salesAssignmentsResponse, err := salesAssignmentService.GetAllSalesAssignment(salesPersonID)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales assignments fetched successfully", salesAssignmentsResponse)
}

func CreateOrUpdateSalesAssignment(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	fmt.Println("Raw Body: ", string(ctx.Body()))

	areaRepo := repositories.NewAreaRepository(configs.DB)
	salesAssignmentRepo := repositories.NewSalesAssignmentRepository(configs.DB)
	salesAssignmentService := services.NewSalesAssignmentService(salesAssignmentRepo, areaRepo)
	salesAssignmentRequest := new(models.SalesAssignmentRequest)

	if err := ctx.BodyParser(salesAssignmentRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(salesAssignmentRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	stringSalesPersonID := ctx.Params("salesPersonId")
	salesPersonID, err := uuid.Parse(stringSalesPersonID)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid sales person ID", nil)
	}

	roleModuleResponse, err := salesAssignmentService.CreateOrUpdateSalesAssignment(salesPersonID, salesAssignmentRequest, ctx, userInfo)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return helpers.Response(ctx, fiber.StatusCreated, "Sales assignment created successfully", roleModuleResponse)
}
