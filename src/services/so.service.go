package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers/documents"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SalesOrderService struct {
	SalesOrderRepository    repositories.SalesOrderRepository
	SalesPersonRepository   repositories.SalesPersonRepository
	CustomerRepository      repositories.CustomerRepository
	ItemRepository          repositories.ItemRepository
	PaymentRepository       repositories.PaymentRepository
	ItemHistoryRepository   repositories.ItemHistoryRepository
}

func NewSalesOrderService(
	soRepo repositories.SalesOrderRepository,
	spRepo repositories.SalesPersonRepository,
	customerRepo repositories.CustomerRepository,
	itemRepo repositories.ItemRepository,
	paymentRepo repositories.PaymentRepository,
	itemHistoryRepo repositories.ItemHistoryRepository,
) *SalesOrderService {
	return &SalesOrderService{
		SalesOrderRepository:  soRepo,
		SalesPersonRepository: spRepo,
		CustomerRepository:    customerRepo,
		ItemRepository:        itemRepo,
		PaymentRepository:     paymentRepo,
		ItemHistoryRepository: itemHistoryRepo,
	}
}

func (service *SalesOrderService) GetAllSalesOrders() ([]models.ResponseGetSalesOrder, error) {
	sos, err := service.SalesOrderRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	sosResponse := []models.ResponseGetSalesOrder{}
	for _, so := range sos {
		sosResponse = append(sosResponse, service.mapSOToResponse(so))
	}
	return sosResponse, nil
}

func (service *SalesOrderService) GetAllSalesOrdersPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.SalesOrderPaginatedResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Status == "" {
		req.Status = "active"
	}

	sos, totalCount, err := service.SalesOrderRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	sosResponse := []models.ResponseGetSalesOrder{}
	for _, so := range sos {
		sosResponse = append(sosResponse, service.mapSOToResponse(so))
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	paginationResponse := models.PaginationResponse{
		CurrentPage:  req.Page,
		PerPage:      req.Limit,
		TotalPages:   totalPages,
		TotalRecords: totalCount,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
	}

	return &models.SalesOrderPaginatedResponse{
		Data:       sosResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *SalesOrderService) GetSalesOrderByID(soId string) (*models.ResponseGetSalesOrder, error) {
	so, err := service.SalesOrderRepository.FindById(nil, soId, false)
	if err != nil {
		return nil, err
	}
	resp := service.mapSOToResponse(*so)
	return &resp, nil
}

func (service *SalesOrderService) GenerateDocumentDeliveryOrder(soId string) (string, []byte, error) {
	so, err := service.SalesOrderRepository.FindById(nil, soId, true)
	if err != nil {
		return "", nil, fmt.Errorf("Sales order not found: %w", err)
	}

	filename, data, err := documents.GenerateDeliveryOrderPDF(so)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate delivery order PDF: %w", err)
	}

	if so.SOStatus == "Confirmed" {
		if err := service.SalesOrderRepository.UpdateStatus(nil, soId, "Shipped", ""); err != nil {
			log.Printf("Failed to update Sales order status: %v", err)
		}
	}

	return filename, data, nil
}

func (service *SalesOrderService) GenerateInvoice(soId string) (string, []byte, error) {
	so, err := service.SalesOrderRepository.FindById(nil, soId, true)
	if err != nil {
		return "", nil, fmt.Errorf("Sales order not found: %w", err)
	}

	filename, data, err := documents.GenerateInvoicePDF(so)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate invoice PDF: %w", err)
	}

	return filename, data, nil
}

func (service *SalesOrderService) GenerateReceipt(soId string) (string, []byte, error) {
	so, err := service.SalesOrderRepository.FindById(nil, soId, true)
	if err != nil {
		return "", nil, fmt.Errorf("Sales order not found: %w", err)
	}

	filename, data, err := documents.GenerateReceiptPDF(so)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate receipt PDF: %w", err)
	}

	return filename, data, nil
}

func (service *SalesOrderService) CreateSalesOrder(
	soRequest *models.SalesOrderCreateRequest,
	userInfo *models.User,
) (*models.SalesOrder, error) {
	_ = userInfo

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// validate master data di dalam tx
	if _, err := service.SalesPersonRepository.FindById(tx, soRequest.SalesPersonID.String(), false); err != nil {
		tx.Rollback()
		return nil, errors.New("sales person not found")
	}
	if _, err := service.CustomerRepository.FindById(tx, soRequest.CustomerID.String(), false); err != nil {
		tx.Rollback()
		return nil, errors.New("customer not found")
	}

	// lock & cek stok
	if err := service.validateAndLockStock(tx, soRequest.Items); err != nil {
		tx.Rollback()
		return nil, err
	}

	var totalAmount int
	soItems := make([]models.SalesOrderItem, 0, len(soRequest.Items))

	for _, itemReq := range soRequest.Items {
		itemData, err := service.ItemRepository.FindById(tx, itemReq.ItemID.String(), false)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("item %s not found", itemReq.ItemID.String())
		}

		totalPrice := itemReq.Quantity * itemReq.UnitPrice
		totalAmount += totalPrice

		soItems = append(soItems, models.SalesOrderItem{
			ID:         uuid.New(),
			ItemID:     itemReq.ItemID,
			Quantity:   itemReq.Quantity,
			UoMID:      itemData.UoMID,
			UnitPrice:  itemReq.UnitPrice,
			TotalPrice: totalPrice,
		})
	}

	soNumber, err := service.SalesOrderRepository.GenerateNextSONumber(tx)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error generating SO number: %w", err)
	}

	paymentStatus := "Unpaid"
	if soRequest.TermOfPayment == "DP" && soRequest.DPAmount > 0 {
		paymentStatus = "Partial"
	}

	newSO := &models.SalesOrder{
		ID:               uuid.New(),
		SONumber:         soNumber,
		SalesPersonID:    soRequest.SalesPersonID,
		CustomerID:       soRequest.CustomerID,
		SODate:           soRequest.SODate,
		EstimatedArrival: soRequest.EstimatedArrival,
		TermOfPayment:    soRequest.TermOfPayment,
		SOStatus:         soRequest.SOStatus,
		PaymentStatus:    paymentStatus,
		TotalAmount:      totalAmount,
		DPAmount:         soRequest.DPAmount,
		DueDate:          soRequest.DueDate,
		Notes:            soRequest.Notes,
		SalesOrderItems:  soItems,
	}

	if _, err := service.SalesOrderRepository.Insert(tx, newSO); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating sales order: %w", err)
	}

	if soRequest.TermOfPayment == "DP" && soRequest.DPAmount > 0 {
		dpPayment := &models.Payment{
			ID:           uuid.New(),
			SalesOrderID: &newSO.ID,
			PaymentType:  "DP",
			Amount:       soRequest.DPAmount,
			PaymentDate:  time.Now(),
			PaymentMethod:"Pending",
			Notes:        "Initial DP payment",
		}
		if _, err := service.PaymentRepository.Insert(tx, dpPayment); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error creating DP payment: %w", err)
		}

		newSO.PaidAmount = soRequest.DPAmount
		if _, err := service.SalesOrderRepository.Update(tx, newSO); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error updating SO paid amount: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	createdSO, err := service.SalesOrderRepository.FindById(nil, newSO.ID.String(), false)
	if err != nil {
		log.Printf("Warning: SO created but failed to fetch created data: %v", err)
		return newSO, nil
	}
	return createdSO, nil
}

func (service *SalesOrderService) UpdateSalesOrder(
	soId string,
	soRequest *models.SalesOrderUpdateRequest,
	userInfo *models.User,
) (*models.SalesOrder, error) {
	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	so, err := service.SalesOrderRepository.FindById(tx, soId, false)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if so.SOStatus != "Draft" {
		tx.Rollback()
		return nil, errors.New("can only update sales orders in Draft status")
	}

	if len(soRequest.Items) > 0 {
		if err := service.validateAndLockStock(tx, soRequest.Items); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	updates := map[string]interface{}{}

	if soRequest.SalesPersonID != uuid.Nil && soRequest.SalesPersonID != so.SalesPersonID {
		if _, err := service.SalesPersonRepository.FindById(tx, soRequest.SalesPersonID.String(), false); err != nil {
			tx.Rollback()
			return nil, errors.New("sales person not found")
		}
		updates["sales_person_id"] = soRequest.SalesPersonID
	}

	if soRequest.CustomerID != uuid.Nil && soRequest.CustomerID != so.CustomerID {
		if _, err := service.CustomerRepository.FindById(tx, soRequest.CustomerID.String(), false); err != nil {
			tx.Rollback()
			return nil, errors.New("customer not found")
		}
		updates["customer_id"] = soRequest.CustomerID
	}

	if !soRequest.SODate.IsZero() {
		updates["so_date"] = soRequest.SODate
	}
	if soRequest.EstimatedArrival != nil {
		updates["estimated_arrival"] = soRequest.EstimatedArrival
	}
	if soRequest.TermOfPayment != "" {
		updates["term_of_payment"] = soRequest.TermOfPayment
	}
	if soRequest.DPAmount >= 0 {
		updates["dp_amount"] = soRequest.DPAmount
	}
	if soRequest.DueDate != nil {
		updates["due_date"] = soRequest.DueDate
	}
	if soRequest.Notes != "" {
		updates["notes"] = soRequest.Notes
	}

	// items
	if len(soRequest.Items) > 0 {
		var existingItems []models.SalesOrderItem
		if err := tx.Where("sales_order_id = ?", so.ID).Find(&existingItems).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error fetching existing items: %w", err)
		}

		existingByItemID := make(map[uuid.UUID]*models.SalesOrderItem, len(existingItems))
		for i := range existingItems {
			existingByItemID[existingItems[i].ItemID] = &existingItems[i]
		}

		finalItems := make([]models.SalesOrderItem, 0, len(soRequest.Items))
		seen := make(map[uuid.UUID]bool)

		for _, req := range soRequest.Items {
			itemData, err := service.ItemRepository.FindById(tx, req.ItemID.String(), false)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("item %s not found", req.ItemID.String())
			}

			if ex, ok := existingByItemID[req.ItemID]; ok {
				newQty := req.Quantity
				newPrice := req.UnitPrice
				newUoMID := itemData.UoMID
				newTotal := newQty * newPrice

				changed := (ex.Quantity != newQty) ||
					(ex.UnitPrice != newPrice) ||
					(ex.UoMID != newUoMID) ||
					(ex.TotalPrice != newTotal)

				ex.Quantity = newQty
				ex.UnitPrice = newPrice
				ex.UoMID = newUoMID
				ex.TotalPrice = newTotal

				if changed {
					if err := tx.Model(&models.SalesOrderItem{}).
						Where("id = ?", ex.ID).
						Updates(map[string]interface{}{
							"quantity":    ex.Quantity,
							"unit_price":  ex.UnitPrice,
							"uom_id":      ex.UoMID,
							"total_price": ex.TotalPrice,
						}).Error; err != nil {
						tx.Rollback()
						return nil, fmt.Errorf("error updating item %s: %w", ex.ID, err)
					}
				}

				finalItems = append(finalItems, *ex)
				seen[req.ItemID] = true
			} else {
				newRow := models.SalesOrderItem{
					ID:           uuid.New(),
					SalesOrderID: so.ID,
					ItemID:       req.ItemID,
					Quantity:     req.Quantity,
					UoMID:        itemData.UoMID,
					UnitPrice:    req.UnitPrice,
					TotalPrice:   req.Quantity * req.UnitPrice,
				}
				if err := tx.Create(&newRow).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("error creating new item: %w", err)
				}
				finalItems = append(finalItems, newRow)
				seen[req.ItemID] = true
			}
		}

		for _, ex := range existingItems {
			if !seen[ex.ItemID] {
				if err := tx.Unscoped().Delete(&models.SalesOrderItem{}, "id = ?", ex.ID).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("error deleting removed item: %w", err)
				}
			}
		}

		total := 0
		for _, it := range finalItems {
			total += it.TotalPrice
		}
		updates["total_amount"] = total
	}

	// apply updates ke SO
	if len(updates) > 0 {
		if err := tx.Model(&models.SalesOrder{}).
			Where("id = ?", so.ID).
			Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error updating sales order: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	updatedSO, err := service.SalesOrderRepository.FindById(nil, so.ID.String(), false)
	if err != nil {
		log.Printf("Warning: SO updated but failed to fetch updated data: %v", err)
		return so, nil
	}
	return updatedSO, nil
}

func (service *SalesOrderService) UpdateSalesOrderStatus(
	soId string,
	statusRequest *models.SalesOrderStatusUpdateRequest,
	userInfo *models.User,
) error {
	tx := configs.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	so, err := service.SalesOrderRepository.FindById(tx, soId, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := service.validateStatusTransition(so.SOStatus, statusRequest.SOStatus); err != nil {
		tx.Rollback()
		return err
	}

	if statusRequest.SOStatus == "Delivered" {
		for _, soItem := range so.SalesOrderItems {
			item, err := service.ItemRepository.FindById(tx, soItem.ItemID.String(), false)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("item not found: %w", err)
			}
			if item.Stock < soItem.Quantity {
				tx.Rollback()
				return fmt.Errorf("insufficient stock for item %s: available %d, required %d",
					item.ID.String(), item.Stock, soItem.Quantity)
			}

			oldStock := item.Stock
			item.Stock -= soItem.Quantity

			if _, err := service.ItemRepository.Update(tx, item); err != nil {
				tx.Rollback()
				return fmt.Errorf("error updating stock for item %s: %w", item.ID, err)
			}

			if item.Stock <= item.LowStock {
				go func(it models.Item) {
					metadata := map[string]interface{}{
						"item_id":   it.ID.String(),
						"item_name": it.Name,
						"stock":     it.Stock,
						"low_stock": it.LowStock,
					}
					title := fmt.Sprintf("Low Stock Alert: %s", it.Name)
					message := fmt.Sprintf("Stock for item %s is low. Current: %d, Threshold: %d", it.Name, it.Stock, it.LowStock)
					if err := helpers.SendNotificationAuto("low_stock", title, message, metadata); err != nil {
						fmt.Printf("failed to send low stock notification: %v\n", err)
					}
				}(*item)
			}

			// history stok via repo tx-aware
			lastHist, err := service.ItemHistoryRepository.FindLastByItem(tx, item.ID, "stock")
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error querying last stock history: %w", err)
			}

			changeType := "update_stock"
			if lastHist == nil {
				changeType = "create_stock"
				oldStock = 0
			}

			newHist := models.ItemHistory{
				ID:           uuid.New(),
				ItemID:       item.ID,
				ChangeType:   changeType,
				OldStock:     oldStock,
				NewStock:     item.Stock,
				CurrentStock: item.Stock,
				Description:  fmt.Sprintf("Delivered %d units (SO %s)", soItem.Quantity, so.SONumber),
				CreatedBy:    &userInfo.ID,
				UpdatedBy:    &userInfo.ID,
			}
			if _, err := service.ItemHistoryRepository.Insert(tx, &newHist); err != nil {
				tx.Rollback()
				return fmt.Errorf("error creating stock history for item %s: %w", item.ID, err)
			}
		}
	}

	if err := service.SalesOrderRepository.UpdateStatus(tx, so.ID.String(), statusRequest.SOStatus, statusRequest.PaymentStatus); err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating SO status: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (service *SalesOrderService) DeleteSalesOrders(req *models.SalesOrderIsHardDeleteRequest, userInfo *models.User) error {
	isHardDelete := req.IsHardDelete == "hardDelete"

	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to begin transaction for delete %s: %w", id.String(), tx.Error)
		}
		if _, err := service.SalesOrderRepository.FindById(tx, id.String(), true); err != nil {
			tx.Rollback()
			return fmt.Errorf("sales order %s not found: %w", id.String(), err)
		}

		if !isHardDelete {
			so, _ := service.SalesOrderRepository.FindById(tx, id.String(), true)
			if so != nil && so.SOStatus != "Draft" {
				tx.Rollback()
				return fmt.Errorf("can only soft delete sales orders in Draft status")
			}
		}

		if isHardDelete {
			// hard-delete anak langsung via tx
			if err := tx.Unscoped().Delete(&models.SalesOrderItem{}, "sales_order_id = ?", id).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("error hard deleting sales order items for %s: %w", id.String(), err)
			}
			if err := tx.Unscoped().Delete(&models.Payment{}, "sales_order_id = ?", id).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("error hard deleting payments for %s: %w", id.String(), err)
			}
			if err := service.SalesOrderRepository.Delete(tx, id.String(), true); err != nil {
				tx.Rollback()
				return fmt.Errorf("error hard deleting sales order %s: %w", id.String(), err)
			}
		} else {
			if err := service.SalesOrderRepository.Delete(tx, id.String(), false); err != nil {
				tx.Rollback()
				return fmt.Errorf("error soft deleting sales order %s: %w", id.String(), err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit delete %s: %w", id.String(), err)
		}
	}
	return nil
}

func (service *SalesOrderService) RestoreSalesOrders(req *models.SalesOrderRestoreRequest, userInfo *models.User) error {
	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to begin transaction for restore %s: %w", id.String(), tx.Error)
		}

		if _, err := service.SalesOrderRepository.Restore(tx, id.String()); err != nil {
			tx.Rollback()
			return fmt.Errorf("error restoring sales order %s: %w", id.String(), err)
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit restore %s: %w", id.String(), err)
		}
	}
	return nil
}

// Helper types & funcs

type stockViolation struct {
	ItemID    uuid.UUID
	ItemName  string
	Requested int
	Available int
}

func (service *SalesOrderService) mapSOToResponse(so models.SalesOrder) models.ResponseGetSalesOrder {
	return models.ResponseGetSalesOrder{
		ID:               so.ID,
		SONumber:         so.SONumber,
		SalesPersonID:    so.SalesPersonID,
		CustomerID:       so.CustomerID,
		SODate:           so.SODate,
		EstimatedArrival: so.EstimatedArrival,
		TermOfPayment:    so.TermOfPayment,
		SOStatus:         so.SOStatus,
		PaymentStatus:    so.PaymentStatus,
		TotalAmount:      so.TotalAmount,
		PaidAmount:       so.PaidAmount,
		DPAmount:         so.DPAmount,
		DueDate:          so.DueDate,
		Notes:            so.Notes,
		CreatedAt:        so.CreatedAt,
		UpdatedAt:        so.UpdatedAt,
		DeletedAt:        so.DeletedAt,
		SalesPerson:      so.SalesPerson,
		Customer:         so.Customer,
		Payments:         so.Payments,
		SalesOrderItems:  so.SalesOrderItems,
	}
}

func (service *SalesOrderService) validateStatusTransition(currentStatus, newStatus string) error {
	validTransitions := map[string][]string{
		"Draft":     {"Confirmed"},
		"Confirmed": {"Shipped", "Closed"},
		"Shipped":   {"Delivered", "Closed"},
		"Delivered": {"Closed"},
		"Closed":    {},
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("invalid current status: %s", currentStatus)
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
}

func (service *SalesOrderService) validateAndLockStock(tx *gorm.DB, items []models.SalesOrderItemRequest) error {
	violations := make([]stockViolation, 0)

	for _, it := range items {
		var item models.Item
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&item, "id = ?", it.ItemID).Error; err != nil {
			return fmt.Errorf("item %s not found", it.ItemID.String())
		}
		if it.Quantity > item.Stock {
			violations = append(violations, stockViolation{
				ItemID:    item.ID,
				ItemName:  item.Name,
				Requested: it.Quantity,
				Available: item.Stock,
			})
		}
	}

	return formatStockError(violations)
}

func formatStockError(violations []stockViolation) error {
	if len(violations) == 0 {
		return nil
	}
	if len(violations) == 1 {
		v := violations[0]
		return fmt.Errorf("item melebihi stok: %s (diminta %d, stok tersedia %d)", v.ItemName, v.Requested, v.Available)
	}
	msg := "beberapa item melebihi stok:\n"
	for _, v := range violations {
		msg += fmt.Sprintf("- %s (diminta %d, stok tersedia %d)\n", v.ItemName, v.Requested, v.Available)
	}
	return errors.New(msg)
}
