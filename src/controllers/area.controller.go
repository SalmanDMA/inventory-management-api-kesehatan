package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all areas with pagination
// @Description Get all areas with pagination, filtering, and search.
// @Tags Area
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param area_id query string false "Filter by area ID"
// @Success 200 {array} models.ResponseGetArea
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving areas"
// @Router /api/v1/area [get]
func AreaControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	areaRepo := repositories.NewAreaRepository(configs.DB)
	areaService := services.NewAreaService(areaRepo)
	
	areasResponse, err := areaService.GetAllAreasPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting areas", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get all areas", areasResponse)
}

// @Summary Get By ID area
// @Description Get By ID area
// @Tags areas
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetArea
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting area"
// @Router /api/v1/area/{id} [get]
func AreaControllerGetByID(c *fiber.Ctx) error {
	_, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Area info not found", nil)
	}

	id := c.Params("id")
	areaRepo := repositories.NewAreaRepository(configs.DB)
	areaService := services.NewAreaService(areaRepo)
	areaResponse, err := areaService.GetAreaByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting area", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Area fetched successfully", areaResponse)
}

	// @Summary Create area
	// @Description Create area. For now only accessible by users with DEVELOPER or SUPERADMIN areas.
	// @Tags Area
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param areaCreateRequest body models.AreaCreateRequest true "Area create request"
	// @Success 200 {object} models.ResponseGetArea
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create areas"
	// @Failure 409 {string} string "Area already exists"
	// @Failure 500 {string} string "Error creating area"
	// @Router /api/v1/area [post]
	func AreaControllerCreate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		area := new(models.AreaCreateRequest)
	
		if err := ctx.BodyParser(area); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		if err := helpers.ValidateStruct(area); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		areaRepo := repositories.NewAreaRepository(configs.DB)
		areaService := services.NewAreaService(areaRepo)
		
		if _ , err := areaService.CreateArea(area, ctx, userInfo); err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Area created successfully", nil)
	}

	// @Summary Update area
	// @Description Update area. For now only accessible by users with DEVELOPER or SUPERADMIN areas.
	// @Tags Area
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param id path string true "Area ID"
	// @Param areaUpdateRequest body models.AreaCreateRequest true "Area update request"
	// @Success 200 {object} models.ResponseGetArea
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to update areas"
	// @Failure 404 {string} string "Area not found"
	// @Failure 500 {string} string "Error updating area"
	// @Router /api/v1/area/{id} [put]
	func AreaControllerUpdate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		area := new(models.AreaCreateRequest)
	
		if err := ctx.BodyParser(area); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		areaId := ctx.Params("id")
	
		if err := helpers.ValidateStruct(area); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		areaRepo := repositories.NewAreaRepository(configs.DB)
		areaService := services.NewAreaService(areaRepo)
		if _ , err := areaService.UpdateArea(areaId, area, ctx, userInfo); err != nil {
			if err == repositories.ErrAreaNotFound {
				return helpers.Response(ctx, fiber.StatusNotFound, "Area not found", nil)
			}
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating area", nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Area updated successfully", nil)
	}

// AreaControllerDelete adalah handler untuk endpoint area
// @Summary Delete area
// @Description Delete area. For now only accessible by areas with DEVELOPER or SUPERADMIN areas. If the area is hard deleted, the area's avatar will be deleted as well. If the area is soft deleted, the area's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags Area
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.AreaIsHardDeleteRequest true "Area delete request body"
// @Success 200 {array} models.Area
// @Failure 403 {string} string "Forbidden: You do not have access to delete area"
// @Failure 404 {string} string "Area not found"
// @Failure 500 {string} string "Error deleting area"
// @Router /api/v1/area/delete [delete]
func AreaControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Area info not found", nil)
		}

	areaRequest := new(models.AreaIsHardDeleteRequest)
	if err := ctx.BodyParser(areaRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(areaRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	areaRepo := repositories.NewAreaRepository(configs.DB)
	areaService := services.NewAreaService(areaRepo)

	if err := areaService.DeleteAreas(areaRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if areaRequest.IsHardDelete == "hardDelete"  {
		message = "Areas deleted successfully"
	} else {
		message = "Areas moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// AreaControllerRestore restores a soft-deleted area
// @Summary Restore area
// @Description Restore area. For now only accessible by areas with DEVELOPER or SUPERADMIN areas.
// @Tags Area
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.AreaRestoreRequest true "Area restore request body"
// @Success 200 {array} models.Area
// @Failure 403 {string} string "Forbidden: You do not have access to restore area"
// @Failure 404 {string} string "Area not found"
// @Failure 500 {string} string "Error restoring area"
// @Router /api/v1/area/restore [put]
func AreaControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Area info not found", nil)
		}

	areaRequest := new(models.AreaRestoreRequest)
	if err := ctx.BodyParser(areaRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(areaRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	areaRepo := repositories.NewAreaRepository(configs.DB)
	areaService := services.NewAreaService(areaRepo)

	restoredAreas, err := areaService.RestoreAreas(areaRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Areas restored successfully", restoredAreas)
}