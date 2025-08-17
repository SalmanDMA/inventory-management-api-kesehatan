package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all facilityTypes with pagination
// @Description Get all facilityTypes with pagination, filtering, and search.
// @Tags FacilityType
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param facilityType_id query string false "Filter by facilityType ID"
// @Success 200 {array} models.ResponseGetFacilityType
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving facility types"
// @Router /api/v1/facilityType [get]
func FacilityTypeControllerGetAll(ctx *fiber.Ctx) error {
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

	facilityTypeRepo := repositories.NewFacilityTypeRepository(configs.DB)
	facilityTypeService := services.NewFacilityTypeService(facilityTypeRepo)
	
	facilityTypesResponse, err := facilityTypeService.GetAllFacilityTypesPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting facility types", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get all facility types", facilityTypesResponse)
}

// @Summary Get By ID facilityType
// @Description Get By ID facilityType
// @Tags facilityTypes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetFacilityType
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting facility type"
// @Router /api/v1/facilityType/{id} [get]
func FacilityTypeControllerGetByID(c *fiber.Ctx) error {
	_, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Facility type info not found", nil)
	}

	id := c.Params("id")
	facilityTypeRepo := repositories.NewFacilityTypeRepository(configs.DB)
	facilityTypeService := services.NewFacilityTypeService(facilityTypeRepo)
	facilityTypeResponse, err := facilityTypeService.GetFacilityTypeByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting facility type", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Facility type fetched successfully", facilityTypeResponse)
}

	// @Summary Create facilityType
	// @Description Create facilityType. For now only accessible by users with DEVELOPER or SUPERADMIN facilityTypes.
	// @Tags FacilityType
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param facilityTypeCreateRequest body models.FacilityTypeCreateRequest true "Facility type create request"
	// @Success 200 {object} models.ResponseGetFacilityType
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create facility types"
	// @Failure 409 {string} string "Facility type already exists"
	// @Failure 500 {string} string "Error creating facility type"
	// @Router /api/v1/facilityType [post]
	func FacilityTypeControllerCreate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		facilityType := new(models.FacilityTypeCreateRequest)
	
		if err := ctx.BodyParser(facilityType); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		if err := helpers.ValidateStruct(facilityType); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		facilityTypeRepo := repositories.NewFacilityTypeRepository(configs.DB)
		facilityTypeService := services.NewFacilityTypeService(facilityTypeRepo)
		
		if _ , err := facilityTypeService.CreateFacilityType(facilityType, ctx, userInfo); err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Facility type created successfully", nil)
	}

	// @Summary Update facilityType
	// @Description Update facilityType. For now only accessible by users with DEVELOPER or SUPERADMIN facilityTypes.
	// @Tags FacilityType
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param id path string true "Facility type ID"
	// @Param facilityTypeUpdateRequest body models.FacilityTypeCreateRequest true "Facility type update request"
	// @Success 200 {object} models.ResponseGetFacilityType
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to update facility types"
	// @Failure 404 {string} string "Facility type not found"
	// @Failure 500 {string} string "Error updating facility type"
	// @Router /api/v1/facilityType/{id} [put]
	func FacilityTypeControllerUpdate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		facilityType := new(models.FacilityTypeCreateRequest)
	
		if err := ctx.BodyParser(facilityType); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		facilityTypeId := ctx.Params("id")
	
		if err := helpers.ValidateStruct(facilityType); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		facilityTypeRepo := repositories.NewFacilityTypeRepository(configs.DB)
		facilityTypeService := services.NewFacilityTypeService(facilityTypeRepo)
		if _ , err := facilityTypeService.UpdateFacilityType(facilityTypeId, facilityType, ctx, userInfo); err != nil {
			if err == repositories.ErrFacilityTypeNotFound {
				return helpers.Response(ctx, fiber.StatusNotFound, "Facility type not found", nil)
			}
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating facility type", nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Facility type updated successfully", nil)
	}

// FacilityTypeControllerDelete adalah handler untuk endpoint facilityType
// @Summary Delete facilityType
// @Description Delete facilityType. For now only accessible by facilityTypes with DEVELOPER or SUPERADMIN facilityTypes. If the facilityType is hard deleted, the facilityType's avatar will be deleted as well. If the facilityType is soft deleted, the facilityType's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags FacilityType
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.FacilityTypeIsHardDeleteRequest true "Facility type delete request body"
// @Success 200 {array} models.FacilityType
// @Failure 403 {string} string "Forbidden: You do not have access to delete facility type"
// @Failure 404 {string} string "Facility type not found"
// @Failure 500 {string} string "Error deleting facility type"
// @Router /api/v1/facilityType/delete [delete]
func FacilityTypeControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Facility type info not found", nil)
		}

	facilityTypeRequest := new(models.FacilityTypeIsHardDeleteRequest)
	if err := ctx.BodyParser(facilityTypeRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(facilityTypeRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	facilityTypeRepo := repositories.NewFacilityTypeRepository(configs.DB)
	facilityTypeService := services.NewFacilityTypeService(facilityTypeRepo)

	if err := facilityTypeService.DeleteFacilityTypes(facilityTypeRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if facilityTypeRequest.IsHardDelete == "hardDelete"  {
		message = "Facility types deleted successfully"
	} else {
		message = "Facility types moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// FacilityTypeControllerRestore restores a soft-deleted facilityType
// @Summary Restore facilityType
// @Description Restore facilityType. For now only accessible by facilityTypes with DEVELOPER or SUPERADMIN facilityTypes.
// @Tags FacilityType
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.FacilityTypeRestoreRequest true "Facility type restore request body"
// @Success 200 {array} models.FacilityType
// @Failure 403 {string} string "Forbidden: You do not have access to restore facility type"
// @Failure 404 {string} string "Facility type not found"
// @Failure 500 {string} string "Error restoring facility type"
// @Router /api/v1/facilityType/restore [put]
func FacilityTypeControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Facility type info not found", nil)
		}

	facilityTypeRequest := new(models.FacilityTypeRestoreRequest)
	if err := ctx.BodyParser(facilityTypeRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(facilityTypeRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	facilityTypeRepo := repositories.NewFacilityTypeRepository(configs.DB)
	facilityTypeService := services.NewFacilityTypeService(facilityTypeRepo)

	restoredFacilityTypes, err := facilityTypeService.RestoreFacilityTypes(facilityTypeRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Facility types restored successfully", restoredFacilityTypes)
}