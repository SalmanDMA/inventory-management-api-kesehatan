package controllers

import (
	"bytes"
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// GetAllPurchaseOrders
// @Summary List all purchase orders
// @Description Retrieve all purchase orders (non-paginated). Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {array} models.PurchaseOrder "Purchase orders fetched successfully"
// @Failure 401 {string} string "Unauthorized: Unable to retrieve user information"
// @Failure 500 {string} string "Failed to fetch purchase orders"
// @Router /api/v1/purchase-order [get]
func GetAllPurchaseOrders(ctx *fiber.Ctx) error {
	_, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	pos, err := poService.GetAllPurchaseOrders()
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to fetch purchase orders", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Purchase orders fetched successfully", pos)
}

// GetAllPurchaseOrdersPaginated
// @Summary List purchase orders (paginated)
// @Description Retrieve purchase orders with pagination and optional filters. Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page   query int    false "Page number (default: 1)"
// @Param limit  query int    false "Items per page (default: 10)"
// @Param search query string false "Search keyword"
// @Param sort   query string false "Sort field"
// @Param order  query string false "Sort order (asc|desc)"
// @Success 200 {object} models.PaginatedResponse "Purchase orders fetched successfully"
// @Failure 400 {string} string "Invalid query parameters"
// @Failure 401 {string} string "Unauthorized: Unable to retrieve user information"
// @Failure 500 {string} string "Failed to fetch purchase orders"
// @Router /api/v1/purchase-order/paginated [get]
func GetAllPurchaseOrdersPaginated(ctx *fiber.Ctx) error {
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

	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	result, err := poService.GetAllPurchaseOrdersPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to fetch purchase orders", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Purchase orders fetched successfully", result)
}

// GetPurchaseOrderByID
// @Summary Get purchase order by ID
// @Description Retrieve a single purchase order by its ID. Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Purchase Order ID"
// @Success 200 {object} models.PurchaseOrder "Purchase order fetched successfully"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 404 {string} string "Purchase order not found"
// @Router /api/v1/purchase-order/{id} [get]
func GetPurchaseOrderByID(ctx *fiber.Ctx) error {
	_, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	poId := ctx.Params("id")
	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	po, err := poService.GetPurchaseOrderByID(poId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusNotFound, "Purchase order not found", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Purchase order fetched successfully", po)
}

// GenerateDocumentPurchaseOrder
// @Summary Generate purchase order document (PDF)
// @Description Stream purchase order PDF directly (no server-side storage). Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce application/pdf
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Purchase Order ID"
// @Success 200 {file} file "PDF stream"
// @Failure 401 {string} string "Unauthorized: User info not found"
// @Failure 404 {string} string "Purchase order not found"
// @Failure 500 {string} string "Failed to generate purchase order document"
// @Router /api/v1/purchase-order/{id}/document [get]
func GenerateDocumentPurchaseOrder(ctx *fiber.Ctx) error {
	poId := ctx.Params("id")

	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	filename, pdfBytes, err := poService.GenerateDocumentPurchaseOrder(poId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to generate purchase order document", err.Error())
	}

	ctx.Set("Content-Type", "application/pdf")
	ctx.Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	return ctx.SendStream(bytes.NewReader(pdfBytes))
}

// CreatePurchaseOrder
// @Summary Create purchase order
// @Description Create a new purchase order. Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.PurchaseOrderCreateRequest true "Purchase order create request body"
// @Success 201 {object} models.PurchaseOrder "Purchase order created successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to create purchase order"
// @Router /api/v1/purchase-order [post]
func CreatePurchaseOrder(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	poRequest := new(models.PurchaseOrderCreateRequest)
	if err := ctx.BodyParser(poRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(poRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	po, err := poService.CreatePurchaseOrder(poRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to create purchase order", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusCreated, "Purchase order created successfully", po)
}

// UpdatePurchaseOrder
// @Summary Update purchase order
// @Description Update an existing purchase order by ID. Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Purchase Order ID"
// @Param request body models.PurchaseOrderUpdateRequest true "Purchase order update request body"
// @Success 200 {object} models.PurchaseOrder "Purchase order updated successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to update purchase order"
// @Router /api/v1/purchase-order/{id} [put]
func UpdatePurchaseOrder(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	poRequest := new(models.PurchaseOrderUpdateRequest)
	if err := ctx.BodyParser(poRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}
		
	if err := helpers.ValidateStruct(poRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	poId := ctx.Params("id")
	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	po, err := poService.UpdatePurchaseOrder(poId, poRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to update purchase order", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Purchase order updated successfully", po)
}

// UpdatePurchaseOrderStatus
// @Summary Update purchase order status
// @Description Update the status of a purchase order (e.g., draft → ordered → partially_received → received → cancelled). Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Purchase Order ID"
// @Param request body models.PurchaseOrderStatusUpdateRequest true "Purchase order status update request body"
// @Success 200 {string} string "Purchase order status updated successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to update purchase order status"
// @Router /api/v1/purchase-order/{id}/status [put]
func UpdatePurchaseOrderStatus(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	statusRequest := new(models.PurchaseOrderStatusUpdateRequest)
	if err := ctx.BodyParser(statusRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}
		
	if err := helpers.ValidateStruct(statusRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	poId := ctx.Params("id")
	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	err := poService.UpdatePurchaseOrderStatus(poId, statusRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to update purchase order status", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Purchase order status updated successfully", nil)
}

// ReceiveItems
// @Summary Receive items for a purchase order (GRN)
// @Description Post receiving (GRN) for a purchase order: update received/accepted/rejected and stock moves for accepted qty. Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Purchase Order ID"
// @Param request body models.ReceiveItemsRequest true "Receive items request body"
// @Success 200 {string} string "Items received successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to receive items"
// @Router /api/v1/purchase-order/{id}/receive [post]
func ReceiveItems(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	receiveRequest := new(models.ReceiveItemsRequest)
	if err := ctx.BodyParser(receiveRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}
		
	if err := helpers.ValidateStruct(receiveRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	poId := ctx.Params("id")
	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	err := poService.ReceiveItems(poId, receiveRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to receive items", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Items received successfully", nil)
}

// DeletePurchaseOrders
// @Summary Delete purchase orders (soft/hard)
// @Description Delete one or multiple purchase orders. Use is_hard_delete to control hard/soft delete. Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.PurchaseOrderIsHardDeleteRequest true "Delete purchase orders request body"
// @Success 200 {string} string "Purchase orders soft deleted successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to delete purchase orders"
// @Router /api/v1/purchase-order/delete [delete]
func DeletePurchaseOrders(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	deleteRequest := new(models.PurchaseOrderIsHardDeleteRequest)
	if err := ctx.BodyParser(deleteRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}
		
	if err := helpers.ValidateStruct(deleteRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	err := poService.DeletePurchaseOrders(deleteRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to delete purchase orders", err.Error())
	}

	message := "Purchase orders soft deleted successfully"
	if deleteRequest.IsHardDelete == "hardDelete" {
		message = "Purchase orders permanently deleted successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}

// RestorePurchaseOrders
// @Summary Restore purchase orders
// @Description Restore soft-deleted purchase orders. Requires authentication.
// @Tags PurchaseOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.PurchaseOrderRestoreRequest true "Restore purchase orders request body"
// @Success 200 {string} string "Purchase orders restored successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to restore purchase orders"
// @Router /api/v1/purchase-order/restore [post]
func RestorePurchaseOrders(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	restoreRequest := new(models.PurchaseOrderRestoreRequest)
	if err := ctx.BodyParser(restoreRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}
		
	if err := helpers.ValidateStruct(restoreRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	supplierRepo := repositories.NewSupplierRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poService := services.NewPurchaseOrderService(poRepo, supplierRepo, itemRepo, paymentRepo)

	err := poService.RestorePurchaseOrders(restoreRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to restore purchase orders", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Purchase orders restored successfully", nil)
}
