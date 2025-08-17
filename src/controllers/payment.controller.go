package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// CreatePayment membuat pembayaran untuk Purchase Order tertentu
// @Summary Create payment
// @Description Create a new payment for a purchase order. Requires authentication.
// @Tags Payment
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.PaymentCreateRequest true "Payment create request body"
// @Success 201 {object} models.Payment "Payment created successfully"
// @Failure 400 {string} string "Failed to create payment"
// @Failure 401 {string} string "Unauthorized: Unable to retrieve user information"
// @Router /api/v1/payment [post]
func CreatePayment(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	paymentRequest := new(models.PaymentCreateRequest)
	if err := ctx.BodyParser(paymentRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(paymentRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	paymentService := services.NewPaymentService(paymentRepo, poRepo, soRepo, uploadRepo)

	payment, err := paymentService.CreatePayment(paymentRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Failed to create payment", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusCreated, "Payment created successfully", payment)
}

// GetPaymentsByPurchaseOrder mendapatkan daftar pembayaran berdasarkan Purchase Order ID
// @Summary Get payments by purchase order
// @Description Retrieve all payments for a given purchase order ID. Requires authentication.
// @Tags Payment
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param po_id path string true "Purchase Order ID"
// @Success 200 {array} models.Payment "Payments fetched successfully"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 500 {string} string "Failed to fetch payments"
// @Router /api/v1/payment/purchase-order/{po_id} [get]
func GetPaymentsByPurchaseOrder(ctx *fiber.Ctx) error {
	_, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	poId := ctx.Params("po_id")
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	paymentService := services.NewPaymentService(paymentRepo, poRepo, soRepo, uploadRepo)

	payments, err := paymentService.GetPaymentsByPurchaseOrder(poId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to fetch payments", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Payments fetched successfully", payments)
}

// GetPaymentsBySalesOrder mendapatkan daftar pembayaran berdasarkan Sales Order ID
// @Summary Get payments by sales order
// @Description Retrieve all payments for a given sales order ID. Requires authentication.
// @Tags Payment
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param so_id path string true "Sales Order ID"
// @Success 200 {array} models.Payment "Payments fetched successfully"
// @Failure 401 {string} string "Unauthorized: Item info not found"
// @Failure 500 {string} string "Failed to fetch payments"
// @Router /api/v1/payment/sales-order/{so_id} [get]
func GetPaymentsBySalesOrder(ctx *fiber.Ctx) error {
	_, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	soId := ctx.Params("so_id")
	paymentRepo := repositories.NewPaymentRepository(configs.DB)
	poRepo := repositories.NewPurchaseOrderRepository(configs.DB)
	soRepo := repositories.NewSalesOrderRepository(configs.DB)
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	paymentService := services.NewPaymentService(paymentRepo, poRepo, soRepo, uploadRepo)

	payments, err := paymentService.GetPaymentsBySalesOrder(soId)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Failed to fetch payments", err.Error())
	}

	return helpers.Response(ctx, fiber.StatusOK, "Payments fetched successfully", payments)
}
