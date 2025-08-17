package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all facilities with pagination
// @Description Get all facilities with pagination, filtering, and search.
// @Tags Facility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param facility_id query string false "Filter by facility ID"
// @Success 200 {array} models.ResponseGetFacility
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving facilities"
// @Router /api/v1/facility [get]
func FacilityControllerGetAll(ctx *fiber.Ctx) error {
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

	facilityRepo := repositories.NewFacilityRepository(configs.DB)
	facilityService := services.NewFacilityService(facilityRepo)
	
	facilitiesResponse, err := facilityService.GetAllFacilitiesPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting facilities", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get all facilities", facilitiesResponse)
}

// @Summary Get By ID facility
// @Description Get By ID facility
// @Tags facilities
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetFacility
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting facility"
// @Router /api/v1/facility/{id} [get]
func FacilityControllerGetByID(c *fiber.Ctx) error {
	_, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Facility info not found", nil)
	}

	id := c.Params("id")
	facilityRepo := repositories.NewFacilityRepository(configs.DB)
	facilityService := services.NewFacilityService(facilityRepo)
	facilityResponse, err := facilityService.GetFacilityByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting facility", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Facility fetched successfully", facilityResponse)
}

	// @Summary Create facility
	// @Description Create facility. For now only accessible by users with DEVELOPER or SUPERADMIN facilities.
	// @Tags Facility
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param facilityCreateRequest body models.FacilityCreateRequest true "Facility create request"
	// @Success 200 {object} models.ResponseGetFacility
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create facilities"
	// @Failure 409 {string} string "Facility already exists"
	// @Failure 500 {string} string "Error creating facility"
	// @Router /api/v1/facility [post]
	func FacilityControllerCreate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		facility := new(models.FacilityCreateRequest)
	
		if err := ctx.BodyParser(facility); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		if err := helpers.ValidateStruct(facility); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		facilityRepo := repositories.NewFacilityRepository(configs.DB)
		facilityService := services.NewFacilityService(facilityRepo)
		
		if _ , err := facilityService.CreateFacility(facility, ctx, userInfo); err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Facility created successfully", nil)
	}

	// @Summary Update facility
	// @Description Update facility. For now only accessible by users with DEVELOPER or SUPERADMIN facilities.
	// @Tags Facility
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param id path string true "Facility ID"
	// @Param facilityUpdateRequest body models.FacilityCreateRequest true "Facility update request"
	// @Success 200 {object} models.ResponseGetFacility
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to update facilities"
	// @Failure 404 {string} string "Facility not found"
	// @Failure 500 {string} string "Error updating facility"
	// @Router /api/v1/facility/{id} [put]
	func FacilityControllerUpdate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		facility := new(models.FacilityCreateRequest)
	
		if err := ctx.BodyParser(facility); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		facilityId := ctx.Params("id")
	
		if err := helpers.ValidateStruct(facility); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		facilityRepo := repositories.NewFacilityRepository(configs.DB)
		facilityService := services.NewFacilityService(facilityRepo)
		if _ , err := facilityService.UpdateFacility(facilityId, facility, ctx, userInfo); err != nil {
			if err == repositories.ErrFacilityNotFound {
				return helpers.Response(ctx, fiber.StatusNotFound, "Facility not found", nil)
			}
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating facility", nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Facility updated successfully", nil)
	}

// FacilityControllerDelete adalah handler untuk endpoint facility
// @Summary Delete facility
// @Description Delete facility. For now only accessible by facilities with DEVELOPER or SUPERADMIN facilities. If the facility is hard deleted, the facility's avatar will be deleted as well. If the facility is soft deleted, the facility's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags Facility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.FacilityIsHardDeleteRequest true "Facility delete request body"
// @Success 200 {array} models.Facility
// @Failure 403 {string} string "Forbidden: You do not have access to delete facility"
// @Failure 404 {string} string "Facility not found"
// @Failure 500 {string} string "Error deleting facility"
// @Router /api/v1/facility/delete [delete]
func FacilityControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Facility info not found", nil)
		}

	facilityRequest := new(models.FacilityIsHardDeleteRequest)
	if err := ctx.BodyParser(facilityRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(facilityRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	facilityRepo := repositories.NewFacilityRepository(configs.DB)
	facilityService := services.NewFacilityService(facilityRepo)

	if err := facilityService.DeleteFacilities(facilityRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if facilityRequest.IsHardDelete == "hardDelete"  {
		message = "Facilities deleted successfully"
	} else {
		message = "Facilities moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// FacilityControllerRestore restores a soft-deleted facility
// @Summary Restore facility
// @Description Restore facility. For now only accessible by facilities with DEVELOPER or SUPERADMIN facilities.
// @Tags Facility
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.FacilityRestoreRequest true "Facility restore request body"
// @Success 200 {array} models.Facility
// @Failure 403 {string} string "Forbidden: You do not have access to restore facility"
// @Failure 404 {string} string "Facility not found"
// @Failure 500 {string} string "Error restoring facility"
// @Router /api/v1/facility/restore [put]
func FacilityControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Facility info not found", nil)
		}

	facilityRequest := new(models.FacilityRestoreRequest)
	if err := ctx.BodyParser(facilityRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(facilityRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	facilityRepo := repositories.NewFacilityRepository(configs.DB)
	facilityService := services.NewFacilityService(facilityRepo)

	restoredFacilities, err := facilityService.RestoreFacilities(facilityRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Facilities restored successfully", restoredFacilities)
}