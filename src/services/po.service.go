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
	"gorm.io/gorm"
)

type PurchaseOrderService struct {
	PurchaseOrderRepository repositories.PurchaseOrderRepository
	SupplierRepository      repositories.SupplierRepository
	ItemRepository          repositories.ItemRepository
	PaymentRepository       repositories.PaymentRepository
	ItemHistoryRepository   repositories.ItemHistoryRepository
}

func NewPurchaseOrderService(
	poRepo repositories.PurchaseOrderRepository,
	supplierRepo repositories.SupplierRepository,
	itemRepo repositories.ItemRepository,
	paymentRepo repositories.PaymentRepository,
	itemHistoryRepo repositories.ItemHistoryRepository,
) *PurchaseOrderService {
	return &PurchaseOrderService{
		PurchaseOrderRepository: poRepo,
		SupplierRepository:      supplierRepo,
		ItemRepository:          itemRepo,
		PaymentRepository:       paymentRepo,
		ItemHistoryRepository:   itemHistoryRepo,
	}
}

func (service *PurchaseOrderService) GetAllPurchaseOrders() ([]models.ResponseGetPurchaseOrder, error) {
	pos, err := service.PurchaseOrderRepository.FindAll(nil)
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

	pos, totalCount, err := service.PurchaseOrderRepository.FindAllPaginated(nil, req)
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
	po, err := service.PurchaseOrderRepository.FindById(nil, poId, false)
	if err != nil {
		return nil, err
	}

	response := service.mapPOToResponse(*po)
	return &response, nil
}

func (service *PurchaseOrderService) GenerateDocumentPurchaseOrder(poId string) (string, []byte, error) {
	po, err := service.PurchaseOrderRepository.FindById(nil, poId, true)
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

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// validate supplier (dalam tx)
	if _, err := service.SupplierRepository.FindById(tx, poRequest.SupplierID.String(), false); err != nil {
		tx.Rollback()
		return nil, errors.New("supplier not found")
	}

	// build items + hitung total dalam tx (ambil UoM dari item repo)
	var totalAmount int
	poItems := make([]models.PurchaseOrderItem, 0, len(poRequest.Items))

	for _, itemReq := range poRequest.Items {
		itemData, err := service.ItemRepository.FindById(tx, itemReq.ItemID.String(), false)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("item %s not found", itemReq.ItemID.String())
		}

		totalPrice := itemReq.Quantity * itemReq.UnitPrice
		totalAmount += totalPrice

		poItems = append(poItems, models.PurchaseOrderItem{
			ID:         uuid.New(),
			ItemID:     itemReq.ItemID,
			Quantity:   itemReq.Quantity,
			UoMID:      itemData.UoMID,
			UnitPrice:  itemReq.UnitPrice,
			TotalPrice: totalPrice,
			Status:     "Ordered",
		})
	}

	poNumber, err := service.PurchaseOrderRepository.GenerateNextPONumber(tx)
	if err != nil {
		tx.Rollback()
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
		Tax:                poRequest.Tax,
		PurchaseOrderItems: poItems,
	}

	if _, err := service.PurchaseOrderRepository.Insert(tx, newPO); err != nil {
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
		if _, err := service.PaymentRepository.Insert(tx, dpPayment); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error creating DP payment: %w", err)
		}

		// update paid_amount via repo update
		newPO.PaidAmount = poRequest.DPAmount
		if _, err := service.PurchaseOrderRepository.Update(tx, newPO); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error updating PO paid amount: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	createdPO, err := service.PurchaseOrderRepository.FindById(nil, newPO.ID.String(), false)
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

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	po, err := service.PurchaseOrderRepository.FindById(tx, poId, false)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if po.POStatus != "Draft" {
		tx.Rollback()
		return nil, errors.New("can only update purchase orders in Draft status")
	}

	updates := map[string]interface{}{}

	if poRequest.SupplierID != uuid.Nil && poRequest.SupplierID != po.SupplierID {
		if _, err := service.SupplierRepository.FindById(tx, poRequest.SupplierID.String(), false); err != nil {
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
	updates["tax"] = poRequest.Tax

	if len(poRequest.Items) > 0 {
		// ambil existing items dulu
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
					if err := tx.Model(&models.PurchaseOrderItem{}).
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
				newRow := models.PurchaseOrderItem{
					ID:              uuid.New(),
					PurchaseOrderID: po.ID,
					ItemID:          req.ItemID,
					Quantity:        req.Quantity,
					UoMID:           itemData.UoMID,
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

		// hapus yang tidak ada lagi
		for _, ex := range existingItems {
			if !seen[ex.ItemID] {
				if err := tx.Unscoped().Delete(&models.PurchaseOrderItem{}, "id = ?", ex.ID).Error; err != nil {
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

	// lakukan update ke PO kalau ada perubahan
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

	updatedPO, err := service.PurchaseOrderRepository.FindById(nil, po.ID.String(), false)
	if err != nil {
		log.Printf("Warning: PO updated but failed to fetch updated data: %v", err)
		return po, nil
	}
	return updatedPO, nil
}

func (service *PurchaseOrderService) UpdatePurchaseOrderStatus(poId string, statusRequest *models.PurchaseOrderStatusUpdateRequest, userInfo *models.User) error {
	po, err := service.PurchaseOrderRepository.FindById(nil, poId, false)
	if err != nil {
		return err
	}

	// Validate status transitions
	if err := service.validateStatusTransition(po.POStatus, statusRequest.POStatus); err != nil {
		return err
	}

	return service.PurchaseOrderRepository.UpdateStatus(nil, poId, statusRequest.POStatus, statusRequest.PaymentStatus)
}

func (service *PurchaseOrderService) ReceiveItems(poId string, receiveRequest *models.ReceiveItemsRequest, userInfo *models.User) error {
	tx := configs.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	po, err := service.PurchaseOrderRepository.FindById(tx, poId, false)
	if err != nil {
		tx.Rollback()
		return err
	}
	if po.POStatus != "Ordered" {
		tx.Rollback()
		return errors.New("can only receive items for purchase orders in 'Ordered' status")
	}

	allReceived := true
	allReturned := true
	allOrdered := true

	for _, req := range receiveRequest.Items {
		var poItem models.PurchaseOrderItem
		if err := tx.First(&poItem, "id = ? AND purchase_order_id = ?", req.PurchaseOrderItemID, poId).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("purchase order item not found: %w", err)
		}

		oldRecv := poItem.ReceivedQuantity
		oldRet := poItem.ReturnedQuantity
		qty := poItem.Quantity

		deltaRecv := req.ReceivedQuantity
		deltaRet := req.ReturnedQuantity
		if deltaRecv < 0 || deltaRet < 0 {
			tx.Rollback()
			return errors.New("quantities must be non-negative")
		}

		newRecv := oldRecv + deltaRecv
		newRet := oldRet + deltaRet
		totalProcessed := newRecv + newRet

		if totalProcessed > qty {
			tx.Rollback()
			return errors.New("received + returned quantity cannot exceed ordered quantity")
		}

		oldStatus := "Partial"
		switch {
		case oldRecv+oldRet == 0:
			oldStatus = "Ordered"
		case oldRet == qty:
			oldStatus = "Returned"
		case oldRecv == qty && oldRet == 0:
			oldStatus = "Received"
		}

		newStatus := "Partial"
		switch {
		case totalProcessed == 0:
			newStatus = "Ordered"
		case newRet == qty:
			newStatus = "Returned"
		case newRecv == qty && newRet == 0:
			newStatus = "Received"
		default:
			newStatus = "Partial"
		}

		poItem.ReceivedQuantity = newRecv
		poItem.ReturnedQuantity = newRet
		poItem.Status = newStatus

		if newStatus != "Received" {
			allReceived = false
		}
		if newStatus != "Returned" {
			allReturned = false
		}
		if newStatus != "Ordered" {
			allOrdered = false
		}

		if err := tx.Save(&poItem).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("error updating purchase order item: %w", err)
		}

		// Bila item BARU menjadi "Received", baru tambahkan stock + catat history
		if oldStatus != "Received" && newStatus == "Received" {
			item, err := service.ItemRepository.FindById(tx, poItem.ItemID.String(), false)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("item not found: %w", err)
			}

			toAdd := newRecv
			item.Stock += toAdd
			if _, err := service.ItemRepository.Update(tx, item); err != nil {
				tx.Rollback()
				return fmt.Errorf("error updating item stock: %w", err)
			}

			// cek last stock history via repo (tx-aware)
			lastHist, err := service.ItemHistoryRepository.FindLastByItem(tx, poItem.ItemID, "stock")
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error querying last stock history: %w", err)
			}

			changeType := "update_stock"
			oldStock := 0
			if lastHist != nil {
				oldStock = lastHist.CurrentStock
			} else {
				changeType = "create_stock"
			}

			newHist := models.ItemHistory{
				ID:           uuid.New(),
				ItemID:       poItem.ItemID,
				ChangeType:   changeType,
				OldStock:     oldStock,
				NewStock:     item.Stock,
				CurrentStock: item.Stock,
				Description:  fmt.Sprintf("PO fully received: +%d units (%s)", toAdd, po.PONumber),
				CreatedBy:    &userInfo.ID,
				UpdatedBy:    &userInfo.ID,
			}
			if _, err := service.ItemHistoryRepository.Insert(tx, &newHist); err != nil {
				tx.Rollback()
				return fmt.Errorf("error creating item history: %w", err)
			}
		}
	}

	newPOStatus := "Partial"
	switch {
	case allReceived:
		newPOStatus = "Received"
	case allReturned:
		newPOStatus = "Returned"
	case allOrdered:
		newPOStatus = "Ordered"
	default:
		newPOStatus = "Partial"
	}

	if err := service.PurchaseOrderRepository.UpdateStatus(tx, poId, newPOStatus, ""); err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating purchase order status: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (service *PurchaseOrderService) DeletePurchaseOrders(
	req *models.PurchaseOrderIsHardDeleteRequest,
	userInfo *models.User,
) error {
	for _, poId := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for PO delete %v: %v\n", poId, tx.Error)
			return errors.New("error beginning transaction")
		}

		if _, err := service.PurchaseOrderRepository.FindById(tx, poId.String(), true); err != nil {
			tx.Rollback()
			if err == repositories.ErrPurchaseOrderNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Purchase order not found: %v\n", poId)
				continue
			}
			log.Printf("Error finding purchase order %v: %v\n", poId, err)
			return errors.New("error finding purchase order")
		}

		if req.IsHardDelete == "hardDelete" {
			if err := tx.Unscoped().Delete(&models.PurchaseOrderItem{}, "purchase_order_id = ?", poId).Error; err != nil {
				tx.Rollback()
				log.Printf("Error hard deleting purchase order items %v: %v\n", poId, err)
				return errors.New("error hard deleting purchase order items")
			}
			if err := tx.Unscoped().Delete(&models.Payment{}, "purchase_order_id = ?", poId).Error; err != nil {
				tx.Rollback()
				log.Printf("Error hard deleting payments for purchase order %v: %v\n", poId, err)
				return errors.New("error hard deleting payments for purchase order")
			}
			if err := service.PurchaseOrderRepository.Delete(tx, poId.String(), true); err != nil {
				tx.Rollback()
				log.Printf("Error hard deleting purchase order %v: %v\n", poId, err)
				return errors.New("error hard deleting purchase order")
			}
		} else {
			if err := service.PurchaseOrderRepository.Delete(tx, poId.String(), false); err != nil {
				tx.Rollback()
				log.Printf("Error soft deleting purchase order %v: %v\n", poId, err)
				return errors.New("error soft deleting purchase order")
			}
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Failed to commit transaction for purchase order delete %v: %v\n", poId, err)
			return errors.New("error committing transaction")
		}
	}

	return nil
}

func (service *PurchaseOrderService) RestorePurchaseOrders(
	req *models.PurchaseOrderRestoreRequest,
	userInfo *models.User,
) ([]models.PurchaseOrder, error) {
	var restoredPOs []models.PurchaseOrder

	for _, poId := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for user restore %v: %v\n", poId, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		if _, err := service.PurchaseOrderRepository.FindById(tx, poId.String(), true); err != nil {
			tx.Rollback()
			if err == repositories.ErrPurchaseOrderNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Purchase order not found for restore: %v\n", poId)
				continue
			}
			log.Printf("Error finding purchase order %v: %v\n", poId, err)
			return nil, errors.New("error finding purchase order")
		}

		_, err := service.PurchaseOrderRepository.Restore(tx, poId.String())
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrPurchaseOrderNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Purchase order not found for restore: %v\n", poId)
				continue
			}
			log.Printf("Error restoring purchase order %v: %v\n", poId, err)
			return nil, errors.New("error restoring purchase order")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Failed to commit transaction for purchase order restore %v: %v\n", poId, err)
			return nil, errors.New("error committing transaction")
		}

		restoredPO, err := service.PurchaseOrderRepository.FindById(nil, poId.String(), false)
		if err != nil {
			log.Printf("Warning: PO restored but failed to fetch restored data: %v", err)
			continue
		}

		restoredPOs = append(restoredPOs, *restoredPO)
	}

	return restoredPOs, nil
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
		Tax: 															po.Tax,
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
		"Partial":  {"Received", "Returned"},
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
