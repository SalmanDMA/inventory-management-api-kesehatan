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

func GetModulesWithRoleInfo(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	roleIDStr := ctx.Params("roleID")
	roleID, err := uuid.Parse(roleIDStr)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid role ID", nil)
	}

	moduleRepo := repositories.NewModuleRepository(configs.DB)
	rolemoduleRepo := repositories.NewRoleModuleRepository(configs.DB)
	rolemoduleService := services.NewRoleModuleService(rolemoduleRepo, moduleRepo)
	rolemodulesResponse, err := rolemoduleService.GetAllRoleModule(roleID)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Modules fetched successfully", rolemodulesResponse)
}

func CreateOrUpdateRoleModule(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	fmt.Println("Raw Body: ", string(ctx.Body()))

	moduleRepo := repositories.NewModuleRepository(configs.DB)
	rolemoduleRepo := repositories.NewRoleModuleRepository(configs.DB)
	rolemoduleService := services.NewRoleModuleService(rolemoduleRepo, moduleRepo)
	roleModuleRequest := new(models.RoleModuleRequest)

	fmt.Println("Raw Body: ", string(ctx.Body()))

	if err := ctx.BodyParser(roleModuleRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(roleModuleRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	stringRoleId := ctx.Params("roleId")
	roleID, err := uuid.Parse(stringRoleId)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid role ID", nil)
	}

	roleModuleResponse, err := rolemoduleService.CreateOrUpdateRoleModule(roleID, roleModuleRequest, ctx, userInfo)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	return helpers.Response(ctx, fiber.StatusCreated, "Role module created successfully", roleModuleResponse)
}
