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
	SalesOrderRepository  repositories.SalesOrderRepository
	SalesPersonRepository repositories.SalesPersonRepository
	FacilityRepository    repositories.FacilityRepository
	ItemRepository        repositories.ItemRepository
	PaymentRepository     repositories.PaymentRepository
}

func NewSalesOrderService(
	soRepo repositories.SalesOrderRepository,
	spRepo repositories.SalesPersonRepository,
	facilityRepo repositories.FacilityRepository,
	itemRepo repositories.ItemRepository,
	paymentRepo repositories.PaymentRepository,
) *SalesOrderService {
	return &SalesOrderService{
		SalesOrderRepository:  soRepo,
		SalesPersonRepository: spRepo,
		FacilityRepository:    facilityRepo,
		ItemRepository:        itemRepo,
		PaymentRepository:     paymentRepo,
	}
}

func (service *SalesOrderService) GetAllSalesOrders() ([]models.ResponseGetSalesOrder, error) {
	sos, err := service.SalesOrderRepository.FindAll()
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

	sos, totalCount, err := service.SalesOrderRepository.FindAllPaginated(req)
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
	so, err := service.SalesOrderRepository.FindById(soId, false)
	if err != nil {
		return nil, err
	}

	response := service.mapSOToResponse(*so)
	return &response, nil
}

func (service *SalesOrderService) GenerateDocumentDeliveryOrder(soId string) (string, []byte, error) {
	so, err := service.SalesOrderRepository.FindById(soId, true)
	if err != nil {
		return "", nil, fmt.Errorf("Sales order not found: %w", err)
	}

	filename, data, err := documents.GenerateDeliveryOrderPDF(so)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate delivery order PDF: %w", err)
	}

	if so.SOStatus == "Confirmed" {
		if err := service.SalesOrderRepository.UpdateStatus(soId, "Shipped", ""); err != nil {
			log.Printf("Failed to update Sales order status: %v", err)
		}
	}

	return filename, data, nil
}

func (service *SalesOrderService) GenerateInvoice(soId string) (string, []byte, error) {
	so, err := service.SalesOrderRepository.FindById(soId, true)
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
	so, err := service.SalesOrderRepository.FindById(soId, true)
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

	if _, err := service.SalesPersonRepository.FindById(soRequest.SalesPersonID.String(), false); err != nil {
		return nil, errors.New("sales person not found")
	}
	if _, err := service.FacilityRepository.FindById(soRequest.FacilityID.String(), false); err != nil {
		return nil, errors.New("facility not found")
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := service.validateAndLockStock(tx, soRequest.Items); err != nil {
		tx.Rollback()
		return nil, err
	}

	var totalAmount int
	soItems := make([]models.SalesOrderItem, 0, len(soRequest.Items))
	for _, itemReq := range soRequest.Items {
		if _, err := service.ItemRepository.FindById(itemReq.ItemID.String(), false); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("item %s not found", itemReq.ItemID.String())
		}
		totalPrice := itemReq.Quantity * itemReq.UnitPrice
		totalAmount += totalPrice

		soItems = append(soItems, models.SalesOrderItem{
			ID:         uuid.New(),
			ItemID:     itemReq.ItemID,
			Quantity:   itemReq.Quantity,
			UnitPrice:  itemReq.UnitPrice,
			TotalPrice: totalPrice,
		})
	}

	soNumber, err := service.SalesOrderRepository.GenerateNextSONumber()
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
		FacilityID:       soRequest.FacilityID,
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

	if err := tx.Create(newSO).Error; err != nil {
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
		if err := tx.Create(dpPayment).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error creating DP payment: %w", err)
		}
		if err := tx.Model(newSO).Update("paid_amount", soRequest.DPAmount).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error updating SO paid amount: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	createdSO, err := service.SalesOrderRepository.FindById(newSO.ID.String(), false)
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
	so, err := service.SalesOrderRepository.FindById(soId, false)
	if err != nil {
		return nil, err
	}
	if so.SOStatus != "Draft" {
		return nil, errors.New("can only update sales orders in Draft status")
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if len(soRequest.Items) > 0 {
		if err := service.validateAndLockStock(tx, soRequest.Items); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	updates := map[string]interface{}{}

	if soRequest.SalesPersonID != uuid.Nil && soRequest.SalesPersonID != so.SalesPersonID {
		if _, err := service.SalesPersonRepository.FindById(soRequest.SalesPersonID.String(), false); err != nil {
			tx.Rollback()
			return nil, errors.New("sales person not found")
		}
		updates["sales_person_id"] = soRequest.SalesPersonID
	}

	if soRequest.FacilityID != uuid.Nil && soRequest.FacilityID != so.FacilityID {
		if _, err := service.FacilityRepository.FindById(soRequest.FacilityID.String(), false); err != nil {
			tx.Rollback()
			return nil, errors.New("facility not found")
		}
		updates["facility_id"] = soRequest.FacilityID
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
			if _, err := service.ItemRepository.FindById(req.ItemID.String(), false); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("item %s not found", req.ItemID.String())
			}

			if ex, ok := existingByItemID[req.ItemID]; ok {
				changed := (ex.Quantity != req.Quantity) || (ex.UnitPrice != req.UnitPrice)
				ex.Quantity = req.Quantity
				ex.UnitPrice = req.UnitPrice
				ex.TotalPrice = req.Quantity * req.UnitPrice
				if changed {
					if err := tx.Model(&models.SalesOrderItem{}).
						Where("id = ?", ex.ID).
						Updates(map[string]interface{}{
							"quantity":    ex.Quantity,
							"unit_price":  ex.UnitPrice,
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
				if err := tx.Delete(&models.SalesOrderItem{}, "id = ?", ex.ID).Error; err != nil {
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

	updatedSO, err := service.SalesOrderRepository.FindById(so.ID.String(), false)
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
	so, err := service.SalesOrderRepository.FindById(soId, false)
	if err != nil {
		return err
	}

	if err := service.validateStatusTransition(so.SOStatus, statusRequest.SOStatus); err != nil {
		return err
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if statusRequest.SOStatus == "Delivered" {
		for _, soItem := range so.SalesOrderItems {
			var item models.Item
			if err := tx.First(&item, "id = ?", soItem.ItemID).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("item not found: %w", err)
			}

			if item.Stock < soItem.Quantity {
				tx.Rollback()
				return fmt.Errorf(
					"insufficient stock for item %s: available %d, required %d",
					item.ID.String(), item.Stock, soItem.Quantity,
				)
			}

			oldStock := item.Stock
			item.Stock -= soItem.Quantity

			if err := tx.Save(&item).Error; err != nil {
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
					message := fmt.Sprintf(
						"Stock for item %s is low. Current: %d, Threshold: %d",
						it.Name, it.Stock, it.LowStock,
					)

					if err := helpers.SendNotificationAuto(
						"low_stock",
						title,
						message,
						metadata,
					); err != nil {
						fmt.Printf("failed to send low stock notification: %v\n", err)
					}
				}(item)
			}

			var lastStockHist models.ItemHistory
			err := tx.
				Where("item_id = ? AND change_type IN ?",
					item.ID,
					[]string{"create_stock", "update_stock"},
				).
				Order("created_at DESC").
				Limit(1).
				Find(&lastStockHist).Error
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error querying last stock history: %w", err)
			}

			changeType := "update_stock"
			if lastStockHist.ID == uuid.Nil {
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

			if err := tx.Create(&newHist).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("error creating stock history for item %s: %w", item.ID, err)
			}
		}
	}

	if err := tx.Model(&models.SalesOrder{}).
		Where("id = ?", so.ID).
		Updates(map[string]interface{}{
			"so_status":      statusRequest.SOStatus,
			"payment_status": statusRequest.PaymentStatus,
		}).Error; err != nil {
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
		so, err := service.SalesOrderRepository.FindById(id.String(), true)
		if err != nil {
			return fmt.Errorf("sales order %s not found: %w", id.String(), err)
		}

		if so.SOStatus != "Draft" && !isHardDelete {
			return fmt.Errorf("can only soft delete sales orders in Draft status")
		}

		if err := service.SalesOrderRepository.Delete(id.String(), isHardDelete); err != nil {
			return fmt.Errorf("error deleting sales order %s: %w", id.String(), err)
		}
	}
	return nil
}

func (service *SalesOrderService) RestoreSalesOrders(req *models.SalesOrderRestoreRequest, userInfo *models.User) error {
	for _, id := range req.IDs {
		so := &models.SalesOrder{}
		if _, err := service.SalesOrderRepository.Restore(so, id.String()); err != nil {
			return fmt.Errorf("error restoring sales order %s: %w", id.String(), err)
		}
	}
	return nil
}

// Helper functions
type stockViolation struct {
	ItemID    uuid.UUID
	ItemName  string
	Requested int
	Available int
}

func (service *SalesOrderService) mapSOToResponse(so models.SalesOrder) models.ResponseGetSalesOrder {
	return models.ResponseGetSalesOrder{
		ID:            so.ID,
		SONumber:      so.SONumber,
		SalesPersonID: so.SalesPersonID,
		FacilityID:    so.FacilityID,
		SODate:        so.SODate,
		EstimatedArrival: so.EstimatedArrival,
		TermOfPayment: so.TermOfPayment,
		SOStatus:      so.SOStatus,
		PaymentStatus: so.PaymentStatus,
		TotalAmount:   so.TotalAmount,
		PaidAmount:    so.PaidAmount,
		DPAmount:      so.DPAmount,
		DueDate:       so.DueDate,
		Notes:         so.Notes,
		CreatedAt:     so.CreatedAt,
		UpdatedAt:     so.UpdatedAt,
		DeletedAt:     so.DeletedAt,
		SalesPerson:   so.SalesPerson,
		Facility:      so.Facility,
		Payments:      so.Payments,
		SalesOrderItems: so.SalesOrderItems,
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
