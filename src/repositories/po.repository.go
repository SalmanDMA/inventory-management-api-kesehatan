package repositories

import (
	"fmt"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseOrderRepository interface {
	FindAll() ([]models.PurchaseOrder, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.PurchaseOrder, int64, error)
	FindById(poId string, includeTrash bool) (*models.PurchaseOrder, error)
	FindByPONumber(poNumber string) (*models.PurchaseOrder, error)
	CountAllThisMonth() (int64, error)
	CountAllLastMonth() (int64, error)
	CountByStatus(status string) (int64, error)
	CountByPaymentStatus(paymentStatus string) (int64, error)
	Insert(po *models.PurchaseOrder) (*models.PurchaseOrder, error)
	Update(po *models.PurchaseOrder) (*models.PurchaseOrder, error)
	UpdateStatus(poId string, poStatus, paymentStatus string) error
	Delete(poId string, isHardDelete bool) error
	Restore(po *models.PurchaseOrder, poId string) (*models.PurchaseOrder, error)
	GenerateNextPONumber() (string, error)
}

// PurchaseOrderRepositoryImpl implementation
type PurchaseOrderRepositoryImpl struct {
	DB *gorm.DB
}

func NewPurchaseOrderRepository(db *gorm.DB) *PurchaseOrderRepositoryImpl {
	return &PurchaseOrderRepositoryImpl{DB: db}
}

func (r *PurchaseOrderRepositoryImpl) FindAll() ([]models.PurchaseOrder, error) {
	var pos []models.PurchaseOrder
	db := r.DB.Unscoped().
		Preload("Supplier").
		Preload("PurchaseOrderItems").
		Preload("PurchaseOrderItems.Item").
		Preload("PurchaseOrderItems.Item.Category").
		Preload("Payments").
		Preload("Payments.Invoice")

	if err := db.Find(&pos).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}
	return pos, nil
}

func (r *PurchaseOrderRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.PurchaseOrder, int64, error) {
	var pos []models.PurchaseOrder
	var totalCount int64

	query := r.DB.Unscoped().
		Preload("Supplier").
		Preload("PurchaseOrderItems").
		Preload("PurchaseOrderItems.Item").
		Preload("PurchaseOrderItems.Item.Category").
		Preload("Payments").
		Preload("Payments.Invoice")

	switch req.Status {
	case "active":
		query = query.Where("purchase_orders.deleted_at IS NULL")
	case "deleted":
		query = query.Where("purchase_orders.deleted_at IS NOT NULL")
	case "all":
	default:
		query = query.Where("purchase_orders.deleted_at IS NULL")
	}

	if req.SupplierID != "" {
		if supplierUUID, err := uuid.Parse(req.SupplierID); err == nil {
			query = query.Where("supplier_id = ?", supplierUUID)
		}
	}

	if req.POStatus != "" {
		query = query.Where("po_status = ?", req.POStatus)
	}

	if req.PaymentStatus != "" {
		query = query.Where("payment_status = ?", req.PaymentStatus)
	}

	if req.TermOfPayment != "" {
		query = query.Where("term_of_payment = ?", req.TermOfPayment)
	}

	if req.SupplierID != "" {
		if supplierUUID, err := uuid.Parse(req.SupplierID); err == nil {
			query = query.Where("supplier_id = ?", supplierUUID)
		}
	}

	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.Joins("LEFT JOIN suppliers ON suppliers.id = purchase_orders.supplier_id").
			Where(`
				LOWER(purchase_orders.po_number) LIKE ? OR
				LOWER(purchase_orders.notes) LIKE ? OR
				LOWER(suppliers.name) LIKE ? OR
				LOWER(suppliers.code) LIKE ?
			`,
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Model(&models.PurchaseOrder{}).Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "purchase_order")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&pos).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "purchase_order")
	}

	return pos, totalCount, nil
}

func (r *PurchaseOrderRepositoryImpl) FindById(poId string, includeTrash bool) (*models.PurchaseOrder, error) {
	var po *models.PurchaseOrder
	db := r.DB

	if includeTrash {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Supplier").
		Preload("PurchaseOrderItems").
		Preload("PurchaseOrderItems.Item").
		Preload("PurchaseOrderItems.Item.Category").
		Preload("Payments").
		Preload("Payments.Invoice").
		First(&po, "id = ?", poId).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}
	
	return po, nil
}

func (r *PurchaseOrderRepositoryImpl) FindByPONumber(poNumber string) (*models.PurchaseOrder, error) {
	var po *models.PurchaseOrder
	if err := r.DB.Where("po_number = ?", poNumber).First(&po).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}
	return po, nil
}

func (r *PurchaseOrderRepositoryImpl) CountAllThisMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.PurchaseOrder{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())").
		Count(&count).Error
	return count, err
}

func (r *PurchaseOrderRepositoryImpl) CountAllLastMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.PurchaseOrder{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Count(&count).Error
	return count, err
}

func (r *PurchaseOrderRepositoryImpl) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.DB.Model(&models.PurchaseOrder{}).
		Where("po_status = ?", status).
		Count(&count).Error
	return count, err
}

func (r *PurchaseOrderRepositoryImpl) CountByPaymentStatus(paymentStatus string) (int64, error) {
	var count int64
	err := r.DB.Model(&models.PurchaseOrder{}).
		Where("payment_status = ?", paymentStatus).
		Count(&count).Error
	return count, err
}

func (r *PurchaseOrderRepositoryImpl) Insert(po *models.PurchaseOrder) (*models.PurchaseOrder, error) {
	if po.ID == uuid.Nil {
		return nil, fmt.Errorf("purchase order ID cannot be empty")
	}

	if err := r.DB.Create(&po).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}
	return po, nil
}

func (r *PurchaseOrderRepositoryImpl) Update(po *models.PurchaseOrder) (*models.PurchaseOrder, error) {
	if po.ID == uuid.Nil {
		return nil, fmt.Errorf("purchase order ID cannot be empty")
	}

	if err := r.DB.Save(&po).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}
	return po, nil
}

func (r *PurchaseOrderRepositoryImpl) UpdateStatus(poId string, poStatus, paymentStatus string) error {
	updates := make(map[string]interface{})
	
	if poStatus != "" {
		updates["po_status"] = poStatus
	}
	if paymentStatus != "" {
		updates["payment_status"] = paymentStatus
	}
	
	if len(updates) == 0 {
		return fmt.Errorf("no status updates provided")
	}

	if err := r.DB.Model(&models.PurchaseOrder{}).Where("id = ?", poId).Updates(updates).Error; err != nil {
		return HandleDatabaseError(err, "purchase_order")
	}
	return nil
}

func (r *PurchaseOrderRepositoryImpl) Delete(poId string, isHardDelete bool) error {
	var po *models.PurchaseOrder

	if err := r.DB.Unscoped().First(&po, "id = ?", poId).Error; err != nil {
		return HandleDatabaseError(err, "purchase_order")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&po).Error; err != nil {
			return HandleDatabaseError(err, "purchase_order")
		}
	} else {
		if err := r.DB.Delete(&po).Error; err != nil {
			return HandleDatabaseError(err, "purchase_order")
		}
	}
	return nil
}

func (r *PurchaseOrderRepositoryImpl) Restore(po *models.PurchaseOrder, poId string) (*models.PurchaseOrder, error) {
	if err := r.DB.Unscoped().Model(po).Where("id = ?", poId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredPO *models.PurchaseOrder
	if err := r.DB.Unscoped().First(&restoredPO, "id = ?", poId).Error; err != nil {
		return nil, err
	}
	
	return restoredPO, nil
}

func (r *PurchaseOrderRepositoryImpl) GenerateNextPONumber() (string, error) {
	var lastPO models.PurchaseOrder
	currentYear := time.Now().Year()
	prefix := fmt.Sprintf("PO-%d-", currentYear)
	
	err := r.DB.Where("po_number LIKE ?", prefix+"%").
		Order("po_number DESC").
		First(&lastPO).Error
	
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}
	
	var nextNumber int = 1
	if err != gorm.ErrRecordNotFound {
		// Extract number from last PO number
		parts := strings.Split(lastPO.PONumber, "-")
		if len(parts) >= 3 {
			if num, parseErr := fmt.Sscanf(parts[2], "%d", &nextNumber); parseErr == nil && num == 1 {
				nextNumber++
			}
		}
	}
	
	return fmt.Sprintf("%s%04d", prefix, nextNumber), nil
}