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
	_ = userInfo

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
		totalAmount   int
		paidAmount    int
		targetOrderID uuid.UUID
	)

	switch paymentRequest.OrderType {
	case "PO":
		po, err := service.PurchaseOrderRepository.FindById(nil, paymentRequest.PurchaseOrderID.String(), true)
		if err != nil {
			return nil, errors.New("purchase order not found")
		}
		totalAmount = po.TotalAmount
		paidAmount = po.PaidAmount
		targetOrderID = po.ID

	case "SO":
		so, err := service.SalesOrderRepository.FindById(nil, paymentRequest.SalesOrderID.String(), true)
		if err != nil {
			return nil, errors.New("sales order not found")
		}
		totalAmount = so.TotalAmount
		paidAmount = so.PaidAmount
		targetOrderID = so.ID
	}

	remaining := totalAmount - paidAmount
	if paymentRequest.Amount > remaining {
		return nil, fmt.Errorf("payment amount (%d) exceeds remaining amount (%d)", paymentRequest.Amount, remaining)
	}

	// ---- Siapkan entity Payment
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
	} else {
		newPayment.SalesOrderID = &paymentRequest.SalesOrderID
	}

	// ---- Transaksi
	var invoiceUUIDStr string

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			if invoiceUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
			}
		}
	}()

	if file, err := ctx.FormFile("invoice"); err == nil && file != nil {
		subdir := "purchase-orders/invoices"
		if paymentRequest.OrderType == "SO" {
			subdir = "sales-orders/invoices"
		}

		uuidStr, err := helpers.SaveFile(ctx, file, subdir)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to save invoice: %w", err)
		}
		invoiceUUIDStr = uuidStr

		invoiceID, err := uuid.Parse(uuidStr)
		if err != nil {
			_ = tx.Rollback()
			helpers.DeleteLocalFileImmediate(uuidStr)
			return nil, fmt.Errorf("invalid invoice UUID: %w", err)
		}
		newPayment.InvoiceID = &invoiceID
	}

	createdPayment, err := service.PaymentRepository.Insert(tx, newPayment)
	if err != nil {
		_ = tx.Rollback()
		if invoiceUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
		}
		return nil, fmt.Errorf("error creating payment: %w", err)
	}

	newPaid := paidAmount + paymentRequest.Amount
	newPayStat := "Partial"
	if newPaid >= totalAmount {
		newPayStat = "Paid"
	}

	switch paymentRequest.OrderType {
	case "PO":
		if err := tx.Model(&models.PurchaseOrder{}).
			Where("id = ?", targetOrderID).
			Updates(map[string]interface{}{
				"paid_amount":    newPaid,
				"payment_status": newPayStat,
			}).Error; err != nil {
			_ = tx.Rollback()
			if invoiceUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
			}
			return nil, fmt.Errorf("error updating purchase order: %w", err)
		}

	case "SO":
		if err := tx.Model(&models.SalesOrder{}).
			Where("id = ?", targetOrderID).
			Updates(map[string]interface{}{
				"paid_amount":    newPaid,
				"payment_status": newPayStat,
			}).Error; err != nil {
			_ = tx.Rollback()
			if invoiceUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
			}
			return nil, fmt.Errorf("error updating sales order: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		if invoiceUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(invoiceUUIDStr)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	fresh, err := service.PaymentRepository.FindById(nil, createdPayment.ID.String(), true)
	if err != nil {
		log.Printf("Warning: Payment created but failed to fetch created data: %v", err)
		return createdPayment, nil
	}
	return fresh, nil
}

func (service *PaymentService) GetPaymentsByPurchaseOrder(poId string) ([]models.ResponseGetPayment, error) {
	payments, err := service.PaymentRepository.FindByPurchaseOrderId(nil, poId)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetPayment, 0, len(payments))
	for _, p := range payments {
		resp = append(resp, models.ResponseGetPayment{
			ID:              p.ID,
			OrderType:       p.OrderType,
			PurchaseOrderID: p.PurchaseOrderID,
			SalesOrderID:    p.SalesOrderID,
			PaymentType:     p.PaymentType,
			Amount:          p.Amount,
			PaymentDate:     p.PaymentDate,
			PaymentMethod:   p.PaymentMethod,
			ReferenceNumber: p.ReferenceNumber,
			InvoiceID:       p.InvoiceID,
			Notes:           p.Notes,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
			DeletedAt:       p.DeletedAt,
			Invoice:         p.Invoice,
			PurchaseOrder:   p.PurchaseOrder,
			SalesOrder:      p.SalesOrder,
		})
	}
	return resp, nil
}

func (service *PaymentService) GetPaymentsBySalesOrder(soId string) ([]models.ResponseGetPayment, error) {
	payments, err := service.PaymentRepository.FindBySalesOrderId(nil,soId)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetPayment, 0, len(payments))
	for _, p := range payments {
		resp = append(resp, models.ResponseGetPayment{
			ID:              p.ID,
			OrderType:       p.OrderType,
			PurchaseOrderID: p.PurchaseOrderID,
			SalesOrderID:    p.SalesOrderID,
			PaymentType:     p.PaymentType,
			Amount:          p.Amount,
			PaymentDate:     p.PaymentDate,
			PaymentMethod:   p.PaymentMethod,
			ReferenceNumber: p.ReferenceNumber,
			InvoiceID:       p.InvoiceID,
			Notes:           p.Notes,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
			DeletedAt:       p.DeletedAt,
			Invoice:         p.Invoice,
			PurchaseOrder:   p.PurchaseOrder,
			SalesOrder:      p.SalesOrder,
		})
	}
	return resp, nil
}
