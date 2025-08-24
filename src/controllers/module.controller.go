package controllers

import (
	"strconv"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all modules
// @Description Get all modules
// @Tags modules
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {array} models.ResponseGetModule
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting users"
// @Router /api/v1/module [get]
func ModuleControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	moduleRepo := repositories.NewModuleRepository(configs.DB)
	moduleService := services.NewModuleService(moduleRepo)
	modulesResponse, err := moduleService.GetAllModules()
	
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting modules", nil)
	}
	
	return helpers.Response(ctx, fiber.StatusOK, "Modules fetched successfully", modulesResponse)
}

// Handler for creating new module
// @Summary Create new module
// @Description Create new module
// @Tags modules
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param module body models.ModuleCreateRequest true "Module"
// @Success 200 {object} models.ResponseGetModule
// @Failure 400 {string} string "Invalid request body"
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error creating module"
// @Router /api/v1/module [post]
func ModuleControllerCreate(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to create users", nil)
	}

	moduleRequest := new(models.ModuleCreateRequest)
	if err := ctx.BodyParser(moduleRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(moduleRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	moduleRepo := repositories.NewModuleRepository(configs.DB)
	moduleService := services.NewModuleService(moduleRepo)
	moduleResponse, err := moduleService.CreateModule(moduleRequest, ctx, userInfo)
	
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error creating module"+err.Error(), nil)
	}
	
	return helpers.Response(ctx, fiber.StatusOK, "Module created successfully", moduleResponse)
}

// ModuleControllerGetById 
// @Summary Get module by ID
// @Description Get module by ID
// @Tags modules
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Module ID"
// @Success 200 {object} models.ResponseGetModule
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 404 {string} string "Module not found"
// @Failure 500 {string} string "Error getting module"
// @Router /api/v1/module/{id} [get]
func ModuleControllerGetById(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	moduleIdStr := ctx.Params("id")
	moduleId, err := strconv.Atoi(moduleIdStr)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid module ID", nil)
	}

	moduleRepo := repositories.NewModuleRepository(configs.DB)
	module, err := moduleRepo.FindById(nil, moduleId, false)

	if err != nil {
		if err == repositories.ErrModuleNotFound {
			return helpers.Response(ctx, fiber.StatusNotFound, "Module not found", nil)
		}
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting module", nil)
	}

	moduleResponse := models.ResponseGetModule{
		ID:           module.ID,
		Name:         module.Name,
		ModuleTypeID: module.ModuleTypeID,
		ModuleType:   module.ModuleType,
		ParentID:     module.ParentID,
		Parent:       module.Parent,
		Path:         module.Path,
		Route:        module.Route,
		Icon:         module.Icon,
		Children:     module.Children,
		Description:  module.Description,
		RoleModules:  module.RoleModules,
	}

	return helpers.Response(ctx, fiber.StatusOK, "Module fetched successfully", moduleResponse)
}


// ModuleControllerGetModuleRoot
// @Summary Get module root
// @Description Get module root
// @Tags modules
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetModule
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting module root"
// @Router /api/v1/module/root [get]
// func ModuleControllerGetModuleRoot(ctx *fiber.Ctx) error {
// 	userInfo, ok := ctx.Locals("userInfo").(*models.User)
// 	if !ok {
// 					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
// 	}

// 	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
// 		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
// 	}

// 	moduleRepo := repositories.NewModuleRepository(configs.DB)
// 	module, err := moduleRepo.FindRootModule()
	
// 	if err != nil {
// 		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting module root", nil)
// 	}
	
// 	return helpers.Response(ctx, fiber.StatusOK, "Module root fetched successfully", module)
// }

// ModuleControllerUpdate 
// @Summary Update module
// @Description Update module
// @Tags modules
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Module ID"
// @Param module body models.ModuleUpdateRequest true "Module"
// @Success 200 {object} models.ResponseGetModule
// @Failure 400 {string} string "Invalid request body"
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 404 {string} string "Module not found"
// @Failure 500 {string} string "Error updating module"
// @Router /api/v1/module/{id} [put]
func ModuleControllerUpdate(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	moduleRequest := new(models.ModuleUpdateRequest)

	if err := ctx.BodyParser(moduleRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(moduleRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	moduleIdStr := ctx.Params("id")
	moduleId, err := strconv.Atoi(moduleIdStr)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid module ID", nil)
	}
	moduleRepo := repositories.NewModuleRepository(configs.DB)
	moduleService := services.NewModuleService(moduleRepo)
	moduleResponse, err := moduleService.UpdateModule(moduleId,moduleRequest, ctx, userInfo)
	
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating module"+err.Error(), nil)
	}
	
	return helpers.Response(ctx, fiber.StatusOK, "Module updated successfully", moduleResponse)
}

// ModuleControllerDelete 
// @Summary Delete module
// @Description Delete module
// @Tags modules
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Module ID"
// @Success 200 {object} models.ResponseGetModule
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 404 {string} string "Module not found"
// @Failure 500 {string} string "Error deleting module"
// @Router /api/v1/module/{id} [delete]
func ModuleControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	moduleRequest := new(models.ModuleIsHardDeleteRequest)

	if err := ctx.BodyParser(moduleRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(moduleRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	moduleRepo := repositories.NewModuleRepository(configs.DB)
	moduleService := services.NewModuleService(moduleRepo)
	
	if err := moduleService.DeleteModule(moduleRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error deleting module"+err.Error(), nil)
	}

	var message string
	if moduleRequest.IsHardDelete == "hardDelete" {
		message = "Module deleted successfully"
	} else {
		message = "Module move to trash successfully"
	}
	 
	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}

// ModuleControllerRestore
// @Summary Restore module
// @Description Restore module
// @Tags modules
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Body models.ModuleRestoreRequest true "Module"
// @Success 200 {object} models.ResponseGetModule
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 404 {string} string "Module not found"
// @Failure 500 {string} string "Error restoring module"
// @Router /api/v1/module/restore [put]
func ModuleControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

		if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
			return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to restore users", nil)
		}

		moduleRequest := new(models.ModuleRestoreRequest)
		if err := ctx.BodyParser(moduleRequest); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}

		if err := helpers.ValidateStruct(moduleRequest); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}

		moduleRepo := repositories.NewModuleRepository(configs.DB)
		moduleService := services.NewModuleService(moduleRepo)
		restoredModule, err := moduleService.RestoreModule(moduleRequest, ctx, userInfo)
		
		if err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error restoring module"+err.Error(), nil)
		}
		
		return helpers.Response(ctx, fiber.StatusOK, "Module restored successfully", restoredModule)
}
