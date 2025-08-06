package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all module types
// @Description Get all module types
// @Tags module types
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetModuleType
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting module types"
// @Router /api/v1/moduleType [get]
func ModuleTypeControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	moduleTypeRepo := repositories.NewModuleTypeRepository(configs.DB)
	moduleTypeService := services.NewModuleTypeService(moduleTypeRepo)
	moduleTypesResponse, err := moduleTypeService.GetAllModuleTypes()
	
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting module types", nil)
	}
	
	return helpers.Response(ctx, fiber.StatusOK, "Module types fetched successfully", moduleTypesResponse)
}

// @Summary Get By ID moduleType
// @Description Get By ID moduleType
// @Tags moduleTypes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetModuleType
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting moduleType"
// @Router /api/v1/moduleType/{id} [get]
func ModuleTypeControllerGetByID(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: ModuleType info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	id := ctx.Params("id")
	moduleTypeRepo := repositories.NewModuleTypeRepository(configs.DB)
	moduleTypeService := services.NewModuleTypeService(moduleTypeRepo)
	moduleTypeResponse, err := moduleTypeService.GetModuleTypeByID(id)
	
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting moduleType", nil)
	}
	
	return helpers.Response(ctx, fiber.StatusOK, "ModuleType fetched successfully", moduleTypeResponse)
}

// @Summary Create moduleType
// @Description Create moduleType
// @Tags moduleTypes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param moduleType body models.ModuleTypeRequest true "ModuleType"
// @Success 200 {object} models.ResponseGetModuleType
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error creating moduleType"
// @Router /api/v1/moduleType [post]
func ModuleTypeControllerCreate(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to create users", nil)
	}

	moduleTypeRequest := new(models.ModuleTypeCreateRequest)
	if err := ctx.BodyParser(moduleTypeRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(moduleTypeRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	moduleTypeRepo := repositories.NewModuleTypeRepository(configs.DB)
	moduleTypeService := services.NewModuleTypeService(moduleTypeRepo)
	moduleTypeResponse, err := moduleTypeService.CreateModuleType(moduleTypeRequest, ctx, userInfo)
	
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error creating moduleType", nil)
	}
	
	return helpers.Response(ctx, fiber.StatusOK, "ModuleType created successfully", moduleTypeResponse)
}

// @Summary Update moduleType
// @Description Update moduleType
// @Tags moduleTypes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param moduleType body models.ModuleTypeRequest true "ModuleType"
// @Success 200 {object} models.ResponseGetModuleType
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error updating moduleType"
// @Router /api/v1/moduleType [put]
func ModuleTypeControllerUpdate(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	moduleTypeRequest := new(models.ModuleTypeUpdateRequest)
	if err := ctx.BodyParser(moduleTypeRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(moduleTypeRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	moduleTypeRepo := repositories.NewModuleTypeRepository(configs.DB)
	moduleTypeService := services.NewModuleTypeService(moduleTypeRepo)
	moduleTypeResponse, err := moduleTypeService.UpdateModuleType(ctx.Params("id"),moduleTypeRequest, ctx, userInfo)
	
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating moduleType", nil)
	}
	
	return helpers.Response(ctx, fiber.StatusOK, "ModuleType updated successfully", moduleTypeResponse)
}

// ModuleTypeControllerDelete adalah handler untuk endpoint ModuleType
// @Summary Delete ModuleType
// @Description Delete ModuleType. For now only accessible by ModuleTypes with DEVELOPER or SUPERADMIN roles. If the ModuleType is hard deleted, the ModuleType's avatar will be deleted as well. If the ModuleType is soft deleted, the ModuleType's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags ModuleType
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.ModuleTypeIsHardDeleteRequest true "ModuleType delete request body"
// @Success 200 {array} models.ModuleType
// @Failure 403 {string} string "Forbidden: You do not have access to delete ModuleType"
// @Failure 404 {string} string "ModuleType not found"
// @Failure 500 {string} string "Error deleting ModuleType"
// @Router /api/v1/ModuleType/delete [delete]
func ModuleTypeControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: ModuleType info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to delete module types", nil)
	}

	ModuleTypeRequest := new(models.ModuleTypeIsHardDeleteRequest)
	if err := ctx.BodyParser(ModuleTypeRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(ModuleTypeRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	ModuleTypeRepo := repositories.NewModuleTypeRepository(configs.DB)
	ModuleTypeService := services.NewModuleTypeService(ModuleTypeRepo)

	if err := ModuleTypeService.DeleteModuleTypes(ModuleTypeRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if ModuleTypeRequest.IsHardDelete == "hardDelete"  {
		message = "Module types deleted successfully"
	} else {
		message = "Module types moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// ModuleTypeControllerRestore restores a soft-deleted ModuleType
// @Summary Restore ModuleType
// @Description Restore ModuleType. For now only accessible by ModuleTypes with DEVELOPER or SUPERADMIN roles.
// @Tags ModuleType
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.ModuleTypeRestoreRequest true "ModuleType restore request body"
// @Success 200 {array} models.ModuleType
// @Failure 403 {string} string "Forbidden: You do not have access to restore ModuleType"
// @Failure 404 {string} string "ModuleType not found"
// @Failure 500 {string} string "Error restoring ModuleType"
// @Router /api/v1/ModuleType/restore [put]
func ModuleTypeControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: ModuleType info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to restore module types", nil)
	}

	ModuleTypeRequest := new(models.ModuleTypeRestoreRequest)
	if err := ctx.BodyParser(ModuleTypeRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(ModuleTypeRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	ModuleTypeRepo := repositories.NewModuleTypeRepository(configs.DB)
	ModuleTypeService := services.NewModuleTypeService(ModuleTypeRepo)

	restoredModuleTypes, err := ModuleTypeService.RestoreModuleTypes(ModuleTypeRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Module types restored successfully", restoredModuleTypes)
}
