package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers/documents"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/google/uuid"
)

type PurchaseOrderService struct {
	PurchaseOrderRepository repositories.PurchaseOrderRepository
	SupplierRepository      repositories.SupplierRepository
	ItemRepository          repositories.ItemRepository
	PaymentRepository       repositories.PaymentRepository
}

func NewPurchaseOrderService(
	poRepo repositories.PurchaseOrderRepository,
	supplierRepo repositories.SupplierRepository,
	itemRepo repositories.ItemRepository,
	paymentRepo repositories.PaymentRepository,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		PurchaseOrderRepository: poRepo,
		SupplierRepository:      supplierRepo,
		ItemRepository:          itemRepo,
		PaymentRepository:       paymentRepo,
	}
}

func (service *PurchaseOrderService) GetAllPurchaseOrders() ([]models.ResponseGetPurchaseOrder, error) {
	pos, err := service.PurchaseOrderRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var posResponse []models.ResponseGetPurchaseOrder
	for _, po := range pos {
		posResponse = append(posResponse, service.mapPOToResponse(po))
	}

	return posResponse, nil
}

func (service *PurchaseOrderService) GetAllPurchaseOrdersPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.PurchaseOrderPaginatedResponse, error) {
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

	pos, totalCount, err := service.PurchaseOrderRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	posResponse := []models.ResponseGetPurchaseOrder{}
	for _, po := range pos {
		posResponse = append(posResponse, service.mapPOToResponse(po))
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

	return &models.PurchaseOrderPaginatedResponse{
		Data:       posResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *PurchaseOrderService) GetPurchaseOrderByID(poId string) (*models.ResponseGetPurchaseOrder, error) {
	po, err := service.PurchaseOrderRepository.FindById(poId, false)
	if err != nil {
		return nil, err
	}

	response := service.mapPOToResponse(*po)
	return &response, nil
}

func (service *PurchaseOrderService) GenerateDocumentPurchaseOrder(poId string) (string, []byte, error) {
	po, err := service.PurchaseOrderRepository.FindById(poId, true)
	if err != nil {
		return "", nil, fmt.Errorf("purchase order not found: %w", err)
	}

	filename, data, err := documents.GeneratePurchaseOrderPDF(po)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate purchase order PDF: %w", err)
	}

	return filename, data, nil
}

func (service *PurchaseOrderService) CreatePurchaseOrder(
	poRequest *models.PurchaseOrderCreateRequest,
	userInfo *models.User,
) (*models.PurchaseOrder, error) {
	if _, err := service.SupplierRepository.FindById(poRequest.SupplierID.String(), false); err != nil {
		return nil, errors.New("supplier not found")
	}

	var totalAmount int
	poItems := make([]models.PurchaseOrderItem, 0, len(poRequest.Items))

	for _, itemReq := range poRequest.Items {
		if _, err := service.ItemRepository.FindById(itemReq.ItemID.String(), false); err != nil {
			return nil, fmt.Errorf("item %s not found", itemReq.ItemID.String())
		}
		totalPrice := itemReq.Quantity * itemReq.UnitPrice
		totalAmount += totalPrice

		poItems = append(poItems, models.PurchaseOrderItem{
			ID:         uuid.New(),
			ItemID:     itemReq.ItemID,
			Quantity:   itemReq.Quantity,
			UnitPrice:  itemReq.UnitPrice,
			TotalPrice: totalPrice,
			Status:     "Ordered",
		})
	}

	poNumber, err := service.PurchaseOrderRepository.GenerateNextPONumber()
	if err != nil {
		return nil, fmt.Errorf("error generating PO number: %w", err)
	}

	paymentStatus := "Unpaid"
	if poRequest.TermOfPayment == "DP" && poRequest.DPAmount > 0 {
		paymentStatus = "Partial"
	}

	newPO := &models.PurchaseOrder{
		ID:                 uuid.New(),
		PONumber:           poNumber,
		SupplierID:         poRequest.SupplierID,
		PODate:             poRequest.PODate,
		EstimatedArrival:   poRequest.EstimatedArrival,
		TermOfPayment:      poRequest.TermOfPayment,
		POStatus:           poRequest.POStatus,
		PaymentStatus:      paymentStatus,
		TotalAmount:        totalAmount,
		DPAmount:           poRequest.DPAmount,
		DueDate:            poRequest.DueDate,
		Notes:              poRequest.Notes,
		PurchaseOrderItems: poItems,
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

	if err := tx.Create(newPO).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error creating purchase order: %w", err)
	}

	if poRequest.TermOfPayment == "DP" && poRequest.DPAmount > 0 {
		dpPayment := &models.Payment{
			ID:              uuid.New(),
			PurchaseOrderID: &newPO.ID,
			PaymentType:     "DP",
			Amount:          poRequest.DPAmount,
			PaymentDate:     time.Now(),
			PaymentMethod:   "Pending",
			Notes:           "Initial DP payment",
		}
		if err := tx.Create(dpPayment).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error creating DP payment: %w", err)
		}

	if err := tx.Model(newPO).Update("paid_amount", poRequest.DPAmount).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error updating PO paid amount: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	createdPO, err := service.PurchaseOrderRepository.FindById(newPO.ID.String(), false)
	if err != nil {
		log.Printf("Warning: PO created but failed to fetch created data: %v", err)
		return newPO, nil
	}
	return createdPO, nil
}

func (service *PurchaseOrderService) UpdatePurchaseOrder(
	poId string,
	poRequest *models.PurchaseOrderUpdateRequest,
	userInfo *models.User,
) (*models.PurchaseOrder, error) {
	po, err := service.PurchaseOrderRepository.FindById(poId, false)
	if err != nil {
		return nil, err
	}
	if po.POStatus != "Draft" {
		return nil, errors.New("can only update purchase orders in Draft status")
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

	updates := map[string]interface{}{}

	if poRequest.SupplierID != uuid.Nil && poRequest.SupplierID != po.SupplierID {
		if _, err := service.SupplierRepository.FindById(poRequest.SupplierID.String(), false); err != nil {
			tx.Rollback()
			return nil, errors.New("supplier not found")
		}
		updates["supplier_id"] = poRequest.SupplierID
	}

	if !poRequest.PODate.IsZero() {
		updates["po_date"] = poRequest.PODate
	}
	if poRequest.EstimatedArrival != nil {
		updates["estimated_arrival"] = poRequest.EstimatedArrival
	}
	if poRequest.TermOfPayment != "" {
		updates["term_of_payment"] = poRequest.TermOfPayment
	}

	if poRequest.DPAmount >= 0 {
		updates["dp_amount"] = poRequest.DPAmount
	}
	if poRequest.DueDate != nil {
		updates["due_date"] = poRequest.DueDate
	}
	if poRequest.Notes != "" {
		updates["notes"] = poRequest.Notes
	}

	if len(poRequest.Items) > 0 {
		var existingItems []models.PurchaseOrderItem
		if err := tx.Where("purchase_order_id = ?", po.ID).Find(&existingItems).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error fetching existing items: %w", err)
		}

		existingByItemID := make(map[uuid.UUID]*models.PurchaseOrderItem, len(existingItems))
		for i := range existingItems {
			existingByItemID[existingItems[i].ItemID] = &existingItems[i]
		}

		finalItems := make([]models.PurchaseOrderItem, 0, len(poRequest.Items))
		seen := make(map[uuid.UUID]bool)
		for _, req := range poRequest.Items {
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
					if err := tx.Model(&models.PurchaseOrderItem{}).
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
				newRow := models.PurchaseOrderItem{
					ID:              uuid.New(),
					PurchaseOrderID: po.ID,
					ItemID:          req.ItemID,
					Quantity:        req.Quantity,
					UnitPrice:       req.UnitPrice,
					TotalPrice:      req.Quantity * req.UnitPrice,
					Status:          "Ordered",
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
				if err := tx.Delete(&models.PurchaseOrderItem{}, "id = ?", ex.ID).Error; err != nil {
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
		if err := tx.Model(&models.PurchaseOrder{}).
			Where("id = ?", po.ID).
			Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error updating purchase order: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	updatedPO, err := service.PurchaseOrderRepository.FindById(po.ID.String(), false)
	if err != nil {
		log.Printf("Warning: PO updated but failed to fetch updated data: %v", err)
		return po, nil
	}
	return updatedPO, nil
}

func (service *PurchaseOrderService) UpdatePurchaseOrderStatus(poId string, statusRequest *models.PurchaseOrderStatusUpdateRequest, userInfo *models.User) error {
	po, err := service.PurchaseOrderRepository.FindById(poId, false)
	if err != nil {
		return err
	}

	// Validate status transitions
	if err := service.validateStatusTransition(po.POStatus, statusRequest.POStatus); err != nil {
		return err
	}

	return service.PurchaseOrderRepository.UpdateStatus(poId, statusRequest.POStatus, statusRequest.PaymentStatus)
}

func (service *PurchaseOrderService) ReceiveItems(poId string, receiveRequest *models.ReceiveItemsRequest, userInfo *models.User) error {
	po, err := service.PurchaseOrderRepository.FindById(poId, false)
	if err != nil {
		return err
	}

	if po.POStatus != "Ordered" {
		return errors.New("can only receive items for purchase orders in 'Ordered' status")
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

	allItemsFullyProcessed := true 
	allItemsFullyReturned := true

	for _, req := range receiveRequest.Items {
		var poItem models.PurchaseOrderItem
		if err := tx.First(&poItem, "id = ? AND purchase_order_id = ?", req.PurchaseOrderItemID, poId).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("purchase order item not found: %w", err)
		}

		deltaRecv := req.ReceivedQuantity
		deltaRet := req.ReturnedQuantity
		if deltaRecv < 0 || deltaRet < 0 {
			tx.Rollback()
			return errors.New("quantities must be non-negative")
		}

		newRecv := poItem.ReceivedQuantity + deltaRecv
		newRet := poItem.ReturnedQuantity + deltaRet
		totalProcessed := newRecv + newRet

		if totalProcessed > poItem.Quantity {
			tx.Rollback()
			return errors.New("received + returned quantity cannot exceed ordered quantity")
		}

		poItem.ReceivedQuantity = newRecv
		poItem.ReturnedQuantity = newRet

		switch {
		case totalProcessed == 0:
			poItem.Status = "Ordered"
		case totalProcessed < poItem.Quantity:
			poItem.Status = "Partial"
		default:
			if newRet == poItem.Quantity {
				poItem.Status = "Returned"
			} else if newRecv == poItem.Quantity {
				poItem.Status = "Received"
			} else {
				poItem.Status = "Completed"
			}
		}

		if totalProcessed < poItem.Quantity {
			allItemsFullyProcessed = false
		}
		if !(totalProcessed == poItem.Quantity && newRet == poItem.Quantity) {
			allItemsFullyReturned = false
		}

		if err := tx.Save(&poItem).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error updating purchase order item: %w", err)
		}

		if deltaRecv > 0 {
			var item models.Item
			if err := tx.First(&item, "id = ?", poItem.ItemID).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("item not found: %w", err)
			}

			item.Stock += deltaRecv
			if err := tx.Save(&item).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("error updating item stock: %w", err)
			}

			var lastStockHist models.ItemHistory
			err := tx.
							Where("item_id = ? AND change_type IN ?",
											poItem.ItemID,
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
			oldStock := lastStockHist.CurrentStock
			if lastStockHist.ID == uuid.Nil {
							changeType = "create_stock"
							oldStock = 0
			}

			newHist := models.ItemHistory{
							ID:           uuid.New(),
							ItemID:       poItem.ItemID,
							ChangeType:   changeType,               
							OldStock:     oldStock,
							NewStock:     item.Stock,
							CurrentStock: item.Stock,
							Description:  fmt.Sprintf("Received %d units (%s)", deltaRecv, po.PONumber),
							CreatedBy:    &userInfo.ID,
							UpdatedBy:    &userInfo.ID,
			}

			if err := tx.Create(&newHist).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("error creating item history: %w", err)
			}
		}
	}

	newPOStatus := po.POStatus
	switch {
	case allItemsFullyProcessed && allItemsFullyReturned:
		newPOStatus = "Returned"
	case allItemsFullyProcessed && !allItemsFullyReturned:
		newPOStatus = "Received"
	default:
		newPOStatus = "Ordered"
	}

	if err := tx.Model(&models.PurchaseOrder{}).
		Where("id = ?", poId).
		Update("po_status", newPOStatus).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating purchase order status: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (service *PurchaseOrderService) DeletePurchaseOrders(req *models.PurchaseOrderIsHardDeleteRequest, userInfo *models.User) error {
	isHardDelete := req.IsHardDelete == "hardDelete"

	for _, id := range req.IDs {
		po, err := service.PurchaseOrderRepository.FindById(id.String(), true)
		if err != nil {
			return fmt.Errorf("purchase order %s not found: %w", id.String(), err)
		}

		if po.POStatus != "Draft" && !isHardDelete {
			return fmt.Errorf("can only soft delete purchase orders in Draft status")
		}

		if err := service.PurchaseOrderRepository.Delete(id.String(), isHardDelete); err != nil {
			return fmt.Errorf("error deleting purchase order %s: %w", id.String(), err)
		}
	}

	return nil
}

func (service *PurchaseOrderService) RestorePurchaseOrders(req *models.PurchaseOrderRestoreRequest, userInfo *models.User) error {
	for _, id := range req.IDs {
		po := &models.PurchaseOrder{}
		if _, err := service.PurchaseOrderRepository.Restore(po, id.String()); err != nil {
			return fmt.Errorf("error restoring purchase order %s: %w", id.String(), err)
		}
	}

	return nil
}

// Helper functions
func (service *PurchaseOrderService) mapPOToResponse(po models.PurchaseOrder) models.ResponseGetPurchaseOrder {
	return models.ResponseGetPurchaseOrder{
		ID:                 po.ID,
		PONumber:           po.PONumber,
		SupplierID:         po.SupplierID,
		PODate:             po.PODate,
		EstimatedArrival:   po.EstimatedArrival,
		TermOfPayment:      po.TermOfPayment,
		POStatus:           po.POStatus,
		PaymentStatus:      po.PaymentStatus,
		TotalAmount:        po.TotalAmount,
		PaidAmount:         po.PaidAmount,
		DPAmount:           po.DPAmount,
		DueDate:            po.DueDate,
		Notes:              po.Notes,
		CreatedAt:          po.CreatedAt,
		UpdatedAt:          po.UpdatedAt,
		DeletedAt:          po.DeletedAt,
		Supplier:           po.Supplier,
		PurchaseOrderItems: po.PurchaseOrderItems,
		Payments:           po.Payments,
	}
}

func (service *PurchaseOrderService) validateStatusTransition(currentStatus, newStatus string) error {
	validTransitions := map[string][]string{
		"Draft":    {"Ordered"},
		"Ordered":  {"Received", "Returned", "Closed"},
		"Received": {"Closed", "Returned"},
		"Returned": {"Closed"},
		"Closed":   {},
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

