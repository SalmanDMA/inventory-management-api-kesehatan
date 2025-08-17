package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// SalesPersonControllerGetAll adalah handler untuk endpoint sales person dengan pagination
// @Summary Get all sales persons with pagination
// @Description Get all sales persons with pagination, filtering, and search.
// @Tags SalesPerson
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or salesPersonname"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param role_id query string false "Filter by role ID"
// @Success 200 {object} models.SalesPersonPaginatedResponse
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting sales persons"
// @Router /api/v1/sales-person [get]
func SalesPersonControllerGetAll(ctx *fiber.Ctx) error {
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

	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesPersonService := services.NewSalesPersonService(salesPersonRepo)
	
	salesPersonsResponse, err := salesPersonService.GetAllSalesPersonsPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting sales persons", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get all sales persons", salesPersonsResponse)
}

// SalesPersonControllerCreate adalah handler untuk endpoint sales person
// @Summary Create sales person
// @Description Create sales person.
// @Tags SalesPerson
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param salesPerson body models.SalesPersonCreate true "Sales Person"
// @Success 200 {string} string "Success creating salesPerson"
// @Failure 400 {string} string "Invalid request body"
// @Failure 403 {string} string "Forbidden: You do not have access to create salesPersons"
// @Failure 409 {string} string "Email already exists"
// @Failure 500 {string} string "Error creating salesPerson"
// @Router /api/v1/sales-person [post]
func SalesPersonControllerCreate(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	salesPersonRequest := new(models.SalesPersonCreateRequest)
	if err := ctx.BodyParser(salesPersonRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(salesPersonRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesPersonService := services.NewSalesPersonService(salesPersonRepo)

	newSalesPerson, err := salesPersonService.CreateSalesPerson(salesPersonRequest, ctx, userInfo)
	if err != nil {
		if err.Error() == "email already exists" {
			return helpers.Response(ctx, fiber.StatusConflict, err.Error(), nil)
		}
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error creating sales person: "+err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusCreated, "Success creating sales person", newSalesPerson)
}

// SalesPersonControllerGetById adalah handler untuk endpoint sales person
// @Summary Get sales person by ID
// @Description Get sales person by ID
// @Tags SalesPerson
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Sales person ID"
// @Success 200 {object} models.ResponseGetSalesPerson
// @Failure 403 {string} string "Forbidden: You do not have access to view salesPerson details"
// @Failure 404 {string} string "SalesPerson not found"
// @Failure 500 {string} string "Error getting salesPerson"
// @Router /api/v1/sales-person/{id} [get]
func SalesPersonControllerGetById(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	salesPersonId := ctx.Params("id")
	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesPersonService := services.NewSalesPersonService(salesPersonRepo)
	roleResponse, err := salesPersonService.GetSalesPersonByID(salesPersonId, userInfo)
	
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting sales person", nil)
	}
	
	return helpers.Response(ctx, fiber.StatusOK, "Sales person fetched successfully", roleResponse)
}

// SalesPersonControllerUpdate adalah handler untuk endpoint sales person
// @Summary Update sales person
// @Description Update sales person.
// @Tags SalesPerson
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Sales Person ID"
// @Param request body models.SalesPersonUpdate true "Sales Person update request body"
// @Success 200 {string} string "Success update salesPerson"
// @Failure 403 {string} string "Forbidden: You do not have access to update salesPerson"
// @Failure 404 {string} string "Sales Person not found"
// @Failure 500 {string} string "Error updating salesPerson"
// @Router /api/v1/sales-person/{id} [put]
func SalesPersonControllerUpdate(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	salesPersonRequest := new(models.SalesPersonUpdateRequest)
	if err := ctx.BodyParser(salesPersonRequest); err != nil {
					return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(salesPersonRequest); err != nil {
					errorMessage := helpers.ExtractErrorMessages(err)
					return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesPersonService := services.NewSalesPersonService(salesPersonRepo)
	updatedSalesPerson, err := salesPersonService.UpdateSalesPerson(salesPersonRequest, ctx.Params("id"), ctx, userInfo)
	
	if err != nil {
					if err.Error() == "email already exists" {
									return helpers.Response(ctx, fiber.StatusConflict, "Email already exists", nil)
					}
					return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating salesPerson: "+err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success updating salesPerson", updatedSalesPerson)
}

// SalesPersonControllerDelete adalah handler untuk endpoint salesPerson
// @Summary Delete salesPerson
// @Description Delete salesPerson. For now only accessible by salesPersons with DEVELOPER or SUPERADMIN roles. If the salesPerson is hard deleted, the salesPerson's avatar will be deleted as well. If the salesPerson is soft deleted, the salesPerson's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags SalesPerson
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.SalesPersonIsHardDeleteRequest true "SalesPerson delete request body"
// @Success 200 {array} models.SalesPerson
// @Failure 403 {string} string "Forbidden: You do not have access to delete salesPerson"
// @Failure 404 {string} string "SalesPerson not found"
// @Failure 500 {string} string "Error deleting salesPerson"
// @Router /api/v1/sales-person/delete [delete]
func SalesPersonControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

	salesPersonRequest := new(models.SalesPersonIsHardDeleteRequest)
	if err := ctx.BodyParser(salesPersonRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(salesPersonRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesPersonService := services.NewSalesPersonService(salesPersonRepo)

	if err := salesPersonService.DeleteSalesPersons(salesPersonRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if salesPersonRequest.IsHardDelete == "hardDelete"  {
		message = "Sales persons deleted successfully"
	} else {
		message = "Sales persons moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// SalesPersonControllerRestore restores a soft-deleted sales person
// @Summary Restore sales person
// @Description Restore sales person.
// @Tags SalesPerson
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.SalesPersonRestoreRequest true "SalesPerson restore request body"
// @Success 200 {array} models.SalesPerson
// @Failure 403 {string} string "Forbidden: You do not have access to restore salesPerson"
// @Failure 404 {string} string "SalesPerson not found"
// @Failure 500 {string} string "Error restoring salesPerson"
// @Router /api/v1/sales-person/restore [put]
func SalesPersonControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

	salesPersonRequest := new(models.SalesPersonRestoreRequest)
	if err := ctx.BodyParser(salesPersonRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(salesPersonRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	salesPersonRepo := repositories.NewSalesPersonRepository(configs.DB)
	salesPersonService := services.NewSalesPersonService(salesPersonRepo)	

	restoredSalesPersons, err := salesPersonService.RestoreSalesPersons(salesPersonRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales persons restored successfully", restoredSalesPersons)
}
