package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all customerTypes with pagination
// @Description Get all customerTypes with pagination, filtering, and search.
// @Tags CustomerType
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param customerType_id query string false "Filter by customerType ID"
// @Success 200 {array} models.ResponseGetCustomerType
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving customer types"
// @Router /api/v1/customerType [get]
func CustomerTypeControllerGetAll(ctx *fiber.Ctx) error {
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

	customerTypeRepo := repositories.NewCustomerTypeRepository(configs.DB)
	customerTypeService := services.NewCustomerTypeService(customerTypeRepo)
	
	customerTypesResponse, err := customerTypeService.GetAllCustomerTypesPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting customer types", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get all customer types", customerTypesResponse)
}

// @Summary Get By ID customerType
// @Description Get By ID customerType
// @Tags customerTypes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetCustomerType
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting customer type"
// @Router /api/v1/customerType/{id} [get]
func CustomerTypeControllerGetByID(c *fiber.Ctx) error {
	_, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Customer type info not found", nil)
	}

	id := c.Params("id")
	customerTypeRepo := repositories.NewCustomerTypeRepository(configs.DB)
	customerTypeService := services.NewCustomerTypeService(customerTypeRepo)
	customerTypeResponse, err := customerTypeService.GetCustomerTypeByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting customer type", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Customer type fetched successfully", customerTypeResponse)
}

	// @Summary Create customerType
	// @Description Create customerType. For now only accessible by users with DEVELOPER or SUPERADMIN customerTypes.
	// @Tags CustomerType
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param customerTypeCreateRequest body models.CustomerTypeCreateRequest true "Customer type create request"
	// @Success 200 {object} models.ResponseGetCustomerType
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create customer types"
	// @Failure 409 {string} string "Customer type already exists"
	// @Failure 500 {string} string "Error creating customer type"
	// @Router /api/v1/customerType [post]
	func CustomerTypeControllerCreate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		customerType := new(models.CustomerTypeCreateRequest)
	
		if err := ctx.BodyParser(customerType); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		if err := helpers.ValidateStruct(customerType); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		customerTypeRepo := repositories.NewCustomerTypeRepository(configs.DB)
		customerTypeService := services.NewCustomerTypeService(customerTypeRepo)
		
		if _ , err := customerTypeService.CreateCustomerType(customerType, ctx, userInfo); err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Customer type created successfully", nil)
	}

	// @Summary Update customerType
	// @Description Update customerType. For now only accessible by users with DEVELOPER or SUPERADMIN customerTypes.
	// @Tags CustomerType
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param id path string true "Customer type ID"
	// @Param customerTypeUpdateRequest body models.CustomerTypeCreateRequest true "Customer type update request"
	// @Success 200 {object} models.ResponseGetCustomerType
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to update customer types"
	// @Failure 404 {string} string "Customer type not found"
	// @Failure 500 {string} string "Error updating customer type"
	// @Router /api/v1/customerType/{id} [put]
	func CustomerTypeControllerUpdate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		customerType := new(models.CustomerTypeCreateRequest)
	
		if err := ctx.BodyParser(customerType); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		customerTypeId := ctx.Params("id")
	
		if err := helpers.ValidateStruct(customerType); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		customerTypeRepo := repositories.NewCustomerTypeRepository(configs.DB)
		customerTypeService := services.NewCustomerTypeService(customerTypeRepo)
		if _ , err := customerTypeService.UpdateCustomerType(customerTypeId, customerType, ctx, userInfo); err != nil {
			if err == repositories.ErrCustomerTypeNotFound {
				return helpers.Response(ctx, fiber.StatusNotFound, "Customer type not found", nil)
			}
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating customer type", nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Customer type updated successfully", nil)
	}

// CustomerTypeControllerDelete adalah handler untuk endpoint customerType
// @Summary Delete customerType
// @Description Delete customerType. For now only accessible by customerTypes with DEVELOPER or SUPERADMIN customerTypes. If the customerType is hard deleted, the customerType's avatar will be deleted as well. If the customerType is soft deleted, the customerType's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags CustomerType
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.CustomerTypeIsHardDeleteRequest true "Customer type delete request body"
// @Success 200 {array} models.CustomerType
// @Failure 403 {string} string "Forbidden: You do not have access to delete customer type"
// @Failure 404 {string} string "Customer type not found"
// @Failure 500 {string} string "Error deleting customer type"
// @Router /api/v1/customerType/delete [delete]
func CustomerTypeControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Customer type info not found", nil)
		}

	customerTypeRequest := new(models.CustomerTypeIsHardDeleteRequest)
	if err := ctx.BodyParser(customerTypeRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(customerTypeRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	customerTypeRepo := repositories.NewCustomerTypeRepository(configs.DB)
	customerTypeService := services.NewCustomerTypeService(customerTypeRepo)

	if err := customerTypeService.DeleteCustomerTypes(customerTypeRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if customerTypeRequest.IsHardDelete == "hardDelete"  {
		message = "Customer types deleted successfully"
	} else {
		message = "Customer types moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// CustomerTypeControllerRestore restores a soft-deleted customerType
// @Summary Restore customerType
// @Description Restore customerType. For now only accessible by customerTypes with DEVELOPER or SUPERADMIN customerTypes.
// @Tags CustomerType
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.CustomerTypeRestoreRequest true "Customer type restore request body"
// @Success 200 {array} models.CustomerType
// @Failure 403 {string} string "Forbidden: You do not have access to restore customer type"
// @Failure 404 {string} string "Customer type not found"
// @Failure 500 {string} string "Error restoring customer type"
// @Router /api/v1/customerType/restore [put]
func CustomerTypeControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Customer type info not found", nil)
		}

	customerTypeRequest := new(models.CustomerTypeRestoreRequest)
	if err := ctx.BodyParser(customerTypeRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(customerTypeRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	customerTypeRepo := repositories.NewCustomerTypeRepository(configs.DB)
	customerTypeService := services.NewCustomerTypeService(customerTypeRepo)

	restoredCustomerTypes, err := customerTypeService.RestoreCustomerTypes(customerTypeRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Customer types restored successfully", restoredCustomerTypes)
}