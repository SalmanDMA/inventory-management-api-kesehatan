package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// GetAllSuppliers
// @Summary List all suppliers
// @Description Retrieve all suppliers (non-paginated). Requires authentication.
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {array} models.Supplier "Suppliers fetched successfully"
// @Failure 401 {string} string "Unauthorized: Unable to retrieve user information"
// @Failure 500 {string} string "Failed to fetch suppliers"
// @Router /api/v1/supplier [get]
func GetAllSuppliers(ctx *fiber.Ctx) error {
	_, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	supplierService := services.NewSupplierService(supplierRepo)

	suppliers, err := supplierService.GetAllSuppliers()
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to fetch suppliers", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Suppliers fetched successfully", suppliers)
}

// GetAllSuppliersPaginated
// @Summary List suppliers (paginated)
// @Description Retrieve suppliers with pagination and optional query filters. Requires authentication.
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page  query int    false "Page number (default: 1)"
// @Param limit query int    false "Items per page (default: 10)"
// @Param search query string false "Search keyword"
// @Param sort  query string false "Sort field"
// @Param order query string false "Sort order (asc|desc)"
// @Success 200 {object} models.PaginatedResponse "Suppliers fetched successfully" // ⟵ ganti ke tipe paginated kamu bila berbeda
// @Failure 400 {string} string "Invalid query parameters"
// @Failure 401 {string} string "Unauthorized: Unable to retrieve user information"
// @Failure 500 {string} string "Failed to fetch suppliers"
// @Router /api/v1/supplier/paginated [get]
func GetAllSuppliersPaginated(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	supplierService := services.NewSupplierService(supplierRepo)

	result, err := supplierService.GetAllSuppliersPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to fetch suppliers", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Suppliers fetched successfully", result)
}

// GetSupplierByID
// @Summary Get supplier by ID
// @Description Retrieve a single supplier by its ID. Requires authentication.
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Supplier ID"
// @Success 200 {object} models.Supplier "Supplier fetched successfully"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 404 {string} string "Supplier not found"
// @Router /api/v1/supplier/{id} [get]
func GetSupplierByID(ctx *fiber.Ctx) error {
	_, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	supplierId := ctx.Params("id")
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	supplierService := services.NewSupplierService(supplierRepo)

	supplier, err := supplierService.GetSupplierByID(supplierId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusNotFound, "Supplier not found", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Supplier fetched successfully", supplier)
}

// CreateSupplier
// @Summary Create supplier
// @Description Create a new supplier. Requires authentication.
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.SupplierCreateRequest true "Supplier create request body"
// @Success 201 {object} models.Supplier "Supplier created successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to create supplier"
// @Router /api/v1/supplier [post]
func CreateSupplier(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	supplierRequest := new(models.SupplierCreateRequest)
	if err := ctx.BodyParser(supplierRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(supplierRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	supplierService := services.NewSupplierService(supplierRepo)

	supplier, err := supplierService.CreateSupplier(supplierRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to create supplier", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusCreated, "Supplier created successfully", supplier)
}

// UpdateSupplier
// @Summary Update supplier
// @Description Update an existing supplier by ID. Requires authentication.
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Supplier ID"
// @Param request body models.SupplierUpdateRequest true "Supplier update request body"
// @Success 200 {object} models.Supplier "Supplier updated successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to update supplier"
// @Router /api/v1/supplier/{id} [put]
func UpdateSupplier(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	supplierRequest := new(models.SupplierUpdateRequest)
	if err := ctx.BodyParser(supplierRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(supplierRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	supplierId := ctx.Params("id")
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	supplierService := services.NewSupplierService(supplierRepo)

	supplier, err := supplierService.UpdateSupplier(supplierId, supplierRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to update supplier", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Supplier updated successfully", supplier)
}

// DeleteSuppliers
// @Summary Delete suppliers (soft/hard)
// @Description Delete one or multiple suppliers. Use is_hard_delete to control hard/soft delete. Requires authentication.
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.SupplierIsHardDeleteRequest true "Delete suppliers request body"
// @Success 200 {string} string "Suppliers soft deleted successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to delete suppliers"
// @Router /api/v1/supplier/delete [delete]
func DeleteSuppliers(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	deleteRequest := new(models.SupplierIsHardDeleteRequest)
	if err := ctx.BodyParser(deleteRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(deleteRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	supplierService := services.NewSupplierService(supplierRepo)

	err := supplierService.DeleteSuppliers(deleteRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to delete suppliers", err.Error())
	}

	message := "Suppliers soft deleted successfully"
	if deleteRequest.IsHardDelete == "hardDelete" {
		message = "Suppliers permanently deleted successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}

// RestoreSuppliers
// @Summary Restore suppliers
// @Description Restore soft-deleted suppliers. Requires authentication.
// @Tags Supplier
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.SupplierRestoreRequest true "Restore suppliers request body"
// @Success 200 {string} string "Suppliers restored successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to restore suppliers"
// @Router /api/v1/supplier/restore [post]
func RestoreSuppliers(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	restoreRequest := new(models.SupplierRestoreRequest)
	if err := ctx.BodyParser(restoreRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(restoreRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	supplierService := services.NewSupplierService(supplierRepo)

	restoredSuppliers , err := supplierService.RestoreSuppliers(restoreRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to restore suppliers", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Suppliers restored successfully", restoredSuppliers)
}
