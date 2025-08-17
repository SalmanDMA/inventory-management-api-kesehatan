package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PaymentService struct {
	PaymentRepository       repositories.PaymentRepository
	PurchaseOrderRepository repositories.PurchaseOrderRepository
	SalesOrderRepository    repositories.SalesOrderRepository
	UploadRepository        repositories.UploadRepository
}

func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	poRepo repositories.PurchaseOrderRepository,
	soRepo repositories.SalesOrderRepository,
	uploadRepo repositories.UploadRepository,
) *PaymentService {
	return &PaymentService{
		PaymentRepository:       paymentRepo,
		PurchaseOrderRepository: poRepo,
		SalesOrderRepository:    soRepo,
		UploadRepository:        uploadRepo,
	}
}

func (service *PaymentService) CreatePayment(
	paymentRequest *models.PaymentCreateRequest,
	ctx *fiber.Ctx,
	userInfo *models.User,
) (*models.Payment, error) {
	if paymentRequest.OrderType != "PO" && paymentRequest.OrderType != "SO" {
		return nil, errors.New("order_type must be either PO or SO")
	}
	if paymentRequest.OrderType == "PO" && paymentRequest.PurchaseOrderID == uuid.Nil {
		return nil, errors.New("purchase_order_id is required for PO payments")
	}
	if paymentRequest.OrderType == "SO" && paymentRequest.SalesOrderID == uuid.Nil {
		return nil, errors.New("sales_order_id is required for SO payments")
	}
	if paymentRequest.Amount <= 0 {
		return nil, errors.New("payment amount must be greater than 0")
	}

	var (
		totalAmount    int
		paidAmount     int
		targetOrderID  uuid.UUID
		updateErr error
	)

	switch paymentRequest.OrderType {
	case "PO":
		po, err := service.PurchaseOrderRepository.FindById(paymentRequest.PurchaseOrderID.String(), true)
		if err != nil {
			return nil, errors.New("purchase order not found")
		}
		totalAmount = po.TotalAmount
		paidAmount = po.PaidAmount
		targetOrderID = po.ID

	case "SO":
		if service.SalesOrderRepository == nil {
			return nil, errors.New("sales order repository is not initialized")
		}
		so, err := service.SalesOrderRepository.FindById(paymentRequest.SalesOrderID.String(), true)
		if err != nil {
			return nil, errors.New("sales order not found")
		}
		totalAmount = so.TotalAmount
		paidAmount = so.PaidAmount
		targetOrderID = so.ID
	}

	remainingAmount := totalAmount - paidAmount
	if paymentRequest.Amount > remainingAmount {
		return nil, fmt.Errorf("payment amount (%d) exceeds remaining amount (%d)", paymentRequest.Amount, remainingAmount)
	}

	// --- Siapkan entitas Payment (set FK sesuai tipe) ---
	newPayment := &models.Payment{
		ID:              uuid.New(),
		OrderType:       paymentRequest.OrderType,
		PaymentType:     paymentRequest.PaymentType,
		Amount:          paymentRequest.Amount,
		PaymentDate:     paymentRequest.PaymentDate,
		PaymentMethod:   paymentRequest.PaymentMethod,
		ReferenceNumber: paymentRequest.ReferenceNumber,
		Notes:           paymentRequest.Notes,
	}
	if paymentRequest.OrderType == "PO" {
		newPayment.PurchaseOrderID = &paymentRequest.PurchaseOrderID
		newPayment.SalesOrderID = nil
	} else {
		newPayment.SalesOrderID = &paymentRequest.SalesOrderID
		newPayment.PurchaseOrderID = nil
	}

	// --- Transaksi ---
	var invoiceUUIDStr string
	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if invoiceUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
			}
		}
	}()

	if file, err := ctx.FormFile("invoice"); err == nil && file != nil {
		locationSubdir := "purchase-orders/invoices"
		if paymentRequest.OrderType == "SO" {
			locationSubdir = "sales-orders/invoices"
		}
		invoiceUUIDStr, updateErr = helpers.SaveFile(ctx, file, locationSubdir)
		if updateErr != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to save invoice: %w", updateErr)
		}
		invoiceUUID, parseErr := uuid.Parse(invoiceUUIDStr)
		if parseErr != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
			return nil, fmt.Errorf("invalid invoice UUID: %w", parseErr)
		}
		newPayment.InvoiceID = &invoiceUUID
	}

	if err := tx.Create(newPayment).Error; err != nil {
		tx.Rollback()
		if invoiceUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
		}
		return nil, fmt.Errorf("error creating payment: %w", err)
	}

	newPaidAmount := paidAmount + paymentRequest.Amount
	newPaymentStatus := "Partial"
	if newPaidAmount >= totalAmount {
		newPaymentStatus = "Paid"
	}

	switch paymentRequest.OrderType {
	case "PO":
		if err := tx.Model(&models.PurchaseOrder{}).
			Where("id = ?", targetOrderID).
			Updates(map[string]interface{}{
				"paid_amount":    newPaidAmount,
				"payment_status": newPaymentStatus,
			}).Error; err != nil {
			tx.Rollback()
			if invoiceUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
			}
			return nil, fmt.Errorf("error updating purchase order: %w", err)
		}

	case "SO":
		if err := tx.Model(&models.SalesOrder{}).
			Where("id = ?", targetOrderID).
			Updates(map[string]interface{}{
				"paid_amount":    newPaidAmount,
				"payment_status": newPaymentStatus,
			}).Error; err != nil {
			tx.Rollback()
			if invoiceUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
			}
			return nil, fmt.Errorf("error updating sales order: %w", err)
		}
	}

	// --- Commit ---
	if err := tx.Commit().Error; err != nil {
		if invoiceUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// --- Ambil payment terbaru dari repo ---
	createdPayment, err := service.PaymentRepository.FindById(newPayment.ID.String(), true)
	if err != nil {
		log.Printf("Warning: Payment created but failed to fetch created data: %v", err)
		return newPayment, nil
	}
	return createdPayment, nil
}

func (service *PaymentService) GetPaymentsByPurchaseOrder(poId string) ([]models.ResponseGetPayment, error) {
	payments, err := service.PaymentRepository.FindByPurchaseOrderId(poId)
	if err != nil {
		return nil, err
	}

	var paymentsResponse []models.ResponseGetPayment
	for _, payment := range payments {
		paymentsResponse = append(paymentsResponse, models.ResponseGetPayment{
			ID:              payment.ID,
			OrderType:       payment.OrderType,
			PurchaseOrderID: payment.PurchaseOrderID,
			SalesOrderID:    payment.SalesOrderID,
			PaymentType:     payment.PaymentType,
			Amount:          payment.Amount,
			PaymentDate:     payment.PaymentDate,
			PaymentMethod:   payment.PaymentMethod,
			ReferenceNumber: payment.ReferenceNumber,
			InvoiceID:       payment.InvoiceID,
			Notes:           payment.Notes,
			CreatedAt:       payment.CreatedAt,
			UpdatedAt:       payment.UpdatedAt,
			DeletedAt:       payment.DeletedAt,
			Invoice:         payment.Invoice,
			PurchaseOrder:   payment.PurchaseOrder,
			SalesOrder:      payment.SalesOrder,
		})
	}

	return paymentsResponse, nil
}

func (service *PaymentService) GetPaymentsBySalesOrder(soId string) ([]models.ResponseGetPayment, error) {
	payments, err := service.PaymentRepository.FindBySalesOrderId(soId)
	if err != nil {
		return nil, err
	}

	var paymentsResponse []models.ResponseGetPayment
	for _, payment := range payments {
		paymentsResponse = append(paymentsResponse, models.ResponseGetPayment{
			ID:              payment.ID,
			OrderType:       payment.OrderType,
			PurchaseOrderID: payment.PurchaseOrderID,
			SalesOrderID:    payment.SalesOrderID,
			PaymentType:     payment.PaymentType,
			Amount:          payment.Amount,
			PaymentDate:     payment.PaymentDate,
			PaymentMethod:   payment.PaymentMethod,
			ReferenceNumber: payment.ReferenceNumber,
			InvoiceID:       payment.InvoiceID,
			Notes:           payment.Notes,
			CreatedAt:       payment.CreatedAt,
			UpdatedAt:       payment.UpdatedAt,
			DeletedAt:       payment.DeletedAt,
			Invoice:         payment.Invoice,
			PurchaseOrder:   payment.PurchaseOrder,
			SalesOrder:      payment.SalesOrder,
		})
	}

	return paymentsResponse, nil
}

