package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all uoms with pagination
// @Description Get all uoms with pagination, filtering, and search.
// @Tags UoM
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param uom_id query string false "Filter by uom ID"
// @Success 200 {array} models.ResponseGetUoM
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving customer types"
// @Router /api/v1/uom [get]
func UoMControllerGetAll(ctx *fiber.Ctx) error {
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

	uomRepo := repositories.NewUoMRepository(configs.DB)
	uomService := services.NewUoMService(uomRepo)
	
	uomsResponse, err := uomService.GetAllUoMsPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting customer types", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get all customer types", uomsResponse)
}

// @Summary Get By ID uom
// @Description Get By ID uom
// @Tags uoms
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetUoM
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting customer type"
// @Router /api/v1/uom/{id} [get]
func UoMControllerGetByID(c *fiber.Ctx) error {
	_, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Customer type info not found", nil)
	}

	id := c.Params("id")
	uomRepo := repositories.NewUoMRepository(configs.DB)
	uomService := services.NewUoMService(uomRepo)
	uomResponse, err := uomService.GetUoMByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting customer type", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Customer type fetched successfully", uomResponse)
}

	// @Summary Create uom
	// @Description Create uom. For now only accessible by users with DEVELOPER or SUPERADMIN uoms.
	// @Tags UoM
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param uomCreateRequest body models.UoMCreateRequest true "Customer type create request"
	// @Success 200 {object} models.ResponseGetUoM
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create customer types"
	// @Failure 409 {string} string "Customer type already exists"
	// @Failure 500 {string} string "Error creating customer type"
	// @Router /api/v1/uom [post]
	func UoMControllerCreate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		uom := new(models.UoMCreateRequest)
	
		if err := ctx.BodyParser(uom); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		if err := helpers.ValidateStruct(uom); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		uomRepo := repositories.NewUoMRepository(configs.DB)
		uomService := services.NewUoMService(uomRepo)
		
		if _ , err := uomService.CreateUoM(uom, ctx, userInfo); err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Customer type created successfully", nil)
	}

	// @Summary Update uom
	// @Description Update uom. For now only accessible by users with DEVELOPER or SUPERADMIN uoms.
	// @Tags UoM
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param id path string true "Customer type ID"
	// @Param uomUpdateRequest body models.UoMCreateRequest true "Customer type update request"
	// @Success 200 {object} models.ResponseGetUoM
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to update customer types"
	// @Failure 404 {string} string "Customer type not found"
	// @Failure 500 {string} string "Error updating customer type"
	// @Router /api/v1/uom/{id} [put]
	func UoMControllerUpdate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		uom := new(models.UoMCreateRequest)
	
		if err := ctx.BodyParser(uom); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		uomId := ctx.Params("id")
	
		if err := helpers.ValidateStruct(uom); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		uomRepo := repositories.NewUoMRepository(configs.DB)
		uomService := services.NewUoMService(uomRepo)
		if _ , err := uomService.UpdateUoM(uomId, uom, ctx, userInfo); err != nil {
			if err == repositories.ErrUoMNotFound {
				return helpers.Response(ctx, fiber.StatusNotFound, "Customer type not found", nil)
			}
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating customer type", nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Customer type updated successfully", nil)
	}

// UoMControllerDelete adalah handler untuk endpoint uom
// @Summary Delete uom
// @Description Delete uom. For now only accessible by uoms with DEVELOPER or SUPERADMIN uoms. If the uom is hard deleted, the uom's avatar will be deleted as well. If the uom is soft deleted, the uom's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags UoM
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.UoMIsHardDeleteRequest true "Customer type delete request body"
// @Success 200 {array} models.UoM
// @Failure 403 {string} string "Forbidden: You do not have access to delete customer type"
// @Failure 404 {string} string "Customer type not found"
// @Failure 500 {string} string "Error deleting customer type"
// @Router /api/v1/uom/delete [delete]
func UoMControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Customer type info not found", nil)
		}

	uomRequest := new(models.UoMIsHardDeleteRequest)
	if err := ctx.BodyParser(uomRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(uomRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	uomRepo := repositories.NewUoMRepository(configs.DB)
	uomService := services.NewUoMService(uomRepo)

	if err := uomService.DeleteUoMs(uomRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if uomRequest.IsHardDelete == "hardDelete"  {
		message = "Customer types deleted successfully"
	} else {
		message = "Customer types moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// UoMControllerRestore restores a soft-deleted uom
// @Summary Restore uom
// @Description Restore uom. For now only accessible by uoms with DEVELOPER or SUPERADMIN uoms.
// @Tags UoM
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.UoMRestoreRequest true "Customer type restore request body"
// @Success 200 {array} models.UoM
// @Failure 403 {string} string "Forbidden: You do not have access to restore customer type"
// @Failure 404 {string} string "Customer type not found"
// @Failure 500 {string} string "Error restoring customer type"
// @Router /api/v1/uom/restore [put]
func UoMControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Customer type info not found", nil)
		}

	uomRequest := new(models.UoMRestoreRequest)
	if err := ctx.BodyParser(uomRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(uomRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	uomRepo := repositories.NewUoMRepository(configs.DB)
	uomService := services.NewUoMService(uomRepo)

	restoredUoMs, err := uomService.RestoreUoMs(uomRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Customer types restored successfully", restoredUoMs)
}