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

// GetAllSalesOrders
// @Summary List all sales orders
// @Description Retrieve all sales orders (non-paginated). Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.SalesOrder "Sales orders fetched successfully"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to fetch sales orders"
// @Router /api/v1/sales-order/all [get]
func GetAllSalesOrders(ctx *fiber.Ctx) error {
	_, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	orders, err := soService.GetAllSalesOrders()
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to fetch sales orders", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales orders fetched successfully", orders)
}

// GetAllSalesOrdersPaginated
// @Summary List sales orders (paginated)
// @Description Retrieve sales orders with pagination and optional filters. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page   query int    false "Page number"
// @Param limit  query int    false "Items per page"
// @Param search query string false "Search keyword"
// @Param sort   query string false "Sort field"
// @Param order  query string false "Sort order (asc|desc)"
// @Success 200 {object} models.PaginatedResponse "Sales orders fetched successfully"
// @Failure 400 {string} string "Invalid query parameters"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to fetch sales orders"
// @Router /api/v1/sales-order [get]
func GetAllSalesOrdersPaginated(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	result, err := soService.GetAllSalesOrdersPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to fetch sales orders", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales orders fetched successfully", result)
}

// GetSalesOrderByID
// @Summary Get sales order by ID
// @Description Retrieve a single sales order by its ID. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Sales Order ID"
// @Success 200 {object} models.SalesOrder "Sales order fetched successfully"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Sales order not found"
// @Router /api/v1/sales-order/{id} [get]
func GetSalesOrderByID(ctx *fiber.Ctx) error {
	_, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	soId := ctx.Params("id")
	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	order, err := soService.GetSalesOrderByID(soId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusNotFound, "Sales order not found", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales order fetched successfully", order)
}

// GenerateDocumentDeliveryOrderSalesOrder
// @Summary Generate delivery order (PDF)
// @Description Stream sales order PDF directly. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce application/pdf
// @Security ApiKeyAuth
// @Param id path string true "Sales Order ID"
// @Success 200 {file} file "PDF stream"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Sales order not found"
// @Failure 500 {string} string "Failed to generate delivery order document"
// @Router /api/v1/sales-order/{id}/document [get]
func GenerateDocumentDeliveryOrder(ctx *fiber.Ctx) error {
	soId := ctx.Params("id")

	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	filename, pdfBytes, err := soService.GenerateDocumentDeliveryOrder(soId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to delivery order document", err.Error())
	}

	ctx.Set("Content-Type", "application/pdf")
	ctx.Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	return ctx.SendStream(bytes.NewReader(pdfBytes))
}

// Generate Invoice Sales Order
// @Summary Generate invoice (PDF)
// @Description Stream sales order PDF directly. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce application/pdf
// @Security ApiKeyAuth
// @Param id path string true "Sales Order ID"
// @Success 200 {file} file "PDF stream"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Sales order not found"
// @Failure 500 {string} string "Failed to generate invoice document"
// @Router /api/v1/sales-order/{id}/invoice [get]
func GenerateInvoice(ctx *fiber.Ctx) error {
	soId := ctx.Params("id")

	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	filename, pdfBytes, err := soService.GenerateInvoice(soId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to generate invoice document", err.Error())
	}

	ctx.Set("Content-Type", "application/pdf")
	ctx.Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	return ctx.SendStream(bytes.NewReader(pdfBytes))
}

// Generate Receipt Sales Order
// @Summary Generate receipt (PDF)
// @Description Stream sales order PDF directly. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce application/pdf
// @Security ApiKeyAuth
// @Param id path string true "Sales Order ID"
// @Success 200 {file} file "PDF stream"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Sales order not found"
// @Failure 500 {string} string "Failed to generate receipt document"
// @Router /api/v1/sales-order/{id}/receipt [get]
func GenerateReceipt(ctx *fiber.Ctx) error {
	soId := ctx.Params("id")

	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	filename, pdfBytes, err := soService.GenerateReceipt(soId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to generate receipt document", err.Error())
	}

	ctx.Set("Content-Type", "application/pdf")
	ctx.Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	return ctx.SendStream(bytes.NewReader(pdfBytes))
}

// CreateSalesOrder
// @Summary Create sales order
// @Description Create a new sales order. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SalesOrderCreateRequest true "Sales order create request body"
// @Success 201 {object} models.SalesOrder "Sales order created successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 400 {string} string "Failed to create sales order"
// @Router /api/v1/sales-order [post]
func CreateSalesOrder(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	soRequest := new(models.SalesOrderCreateRequest)
	if err := ctx.BodyParser(soRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(soRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	order, err := soService.CreateSalesOrder(soRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to create sales order", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusCreated, "Sales order created successfully", order)
}

// UpdateSalesOrder
// @Summary Update sales order
// @Description Update an existing sales order by ID. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Sales Order ID"
// @Param request body models.SalesOrderUpdateRequest true "Sales order update request body"
// @Success 200 {object} models.SalesOrder "Sales order updated successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 400 {string} string "Failed to update sales order"
// @Router /api/v1/sales-order/{id} [put]
func UpdateSalesOrder(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	soRequest := new(models.SalesOrderUpdateRequest)
	if err := ctx.BodyParser(soRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(soRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	soId := ctx.Params("id")
	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)
	order, err := soService.UpdateSalesOrder(soId, soRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to update sales order", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales order updated successfully", order)
}

// UpdateSalesOrderStatus
// @Summary Update sales order status
// @Description Update the status of a sales order by ID. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Sales Order ID"
// @Param request body models.SalesOrderStatusUpdateRequest true "Sales order update status request body"
// @Success 200 {object} models.SalesOrder "Sales order status updated successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 400 {string} string "Failed to update sales order status"
// @Router /api/v1/sales-order/{id}/status [put]
func UpdateSalesOrderStatus(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	statusRequest := new(models.SalesOrderStatusUpdateRequest)
	if err := ctx.BodyParser(statusRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(statusRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	soId := ctx.Params("id")
	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	err := soService.UpdateSalesOrderStatus(soId, statusRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to update sales order status", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales order status updated successfully", nil)
}

// DeleteSalesOrder
// @Summary Delete sales order (soft/hard)
// @Description Delete one or multiple sales order. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.SalesOrderIsHardDeleteRequest true "Delete sales orders request body"
// @Success 200 {string} string "Sales orders soft deleted successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to delete sales orders"
// @Router /api/v1/sales-order/delete/ [delete]
func DeleteSalesOrders(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	deleteRequest := new(models.SalesOrderIsHardDeleteRequest)
	if err := ctx.BodyParser(deleteRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}
		
	if err := helpers.ValidateStruct(deleteRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	err := soService.DeleteSalesOrders(deleteRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to delete sales order", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales order deleted successfully", nil)
}

// RestoreSalesOrders
// @Summary Restore sales orders
// @Description Restore soft-deleted sales orders. Requires authentication.
// @Tags SalesOrder
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.SalesOrderRestoreRequest true "Restore sales orders request body"
// @Success 200 {string} string "Sales orders restored successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 400 {string} string "Failed to restore sales orders"
// @Router /api/v1/sales-order/restore [post]
func RestoreSalesOrders(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	restoreRequest := new(models.SalesOrderRestoreRequest)
	if err := ctx.BodyParser(restoreRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}
		
	if err := helpers.ValidateStruct(restoreRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	spRepo := repositories.NewSalesPersonRepository(configs.DB)
	customerRepo := repositories.NewCustomerRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	soService := services.NewSalesOrderService(soRepo, spRepo, customerRepo, itemRepo, paymentRepo, itemHistoryRepo)

	err := soService.RestoreSalesOrders(restoreRequest, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to restore sales order", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Sales order restored successfully", nil)
}
