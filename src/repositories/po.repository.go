package repositories

import (
	"fmt"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface
// ==============================

type PurchaseOrderRepository interface {
	FindAll(tx *gorm.DB) ([]models.PurchaseOrder, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.PurchaseOrder, int64, error)
	FindById(tx *gorm.DB, poId string, includeTrash bool) (*models.PurchaseOrder, error)
	FindByPONumber(tx *gorm.DB, poNumber string) (*models.PurchaseOrder, error)

	CountAllThisMonth(tx *gorm.DB) (int64, error)
	CountAllLastMonth(tx *gorm.DB) (int64, error)
	CountByStatus(tx *gorm.DB, status string) (int64, error)
	CountByPaymentStatus(tx *gorm.DB, paymentStatus string) (int64, error)

	Insert(tx *gorm.DB, po *models.PurchaseOrder) (*models.PurchaseOrder, error)
	Update(tx *gorm.DB, po *models.PurchaseOrder) (*models.PurchaseOrder, error)
	UpdateStatus(tx *gorm.DB, poId string, poStatus, paymentStatus string) error

	Delete(tx *gorm.DB, poId string, isHardDelete bool) error
	Restore(tx *gorm.DB, poId string) (*models.PurchaseOrder, error)

	GenerateNextPONumber(tx *gorm.DB) (string, error)
}

// ==============================
// Implementation
// ==============================

type PurchaseOrderRepositoryImpl struct {
	DB *gorm.DB
}

func NewPurchaseOrderRepository(db *gorm.DB) *PurchaseOrderRepositoryImpl {
	return &PurchaseOrderRepositoryImpl{DB: db}
}

func (r *PurchaseOrderRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *PurchaseOrderRepositoryImpl) FindAll(tx *gorm.DB) ([]models.PurchaseOrder, error) {
	var pos []models.PurchaseOrder
	db := r.useDB(tx).Unscoped().
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

func (r *PurchaseOrderRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.PurchaseOrder, int64, error) {
	var (
		pos        []models.PurchaseOrder
		totalCount int64
	)

	query := r.useDB(tx).Unscoped().
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
		// no filter
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

func (r *PurchaseOrderRepositoryImpl) FindById(tx *gorm.DB, poId string, includeTrash bool) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	db := r.useDB(tx)
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

	return &po, nil
}

func (r *PurchaseOrderRepositoryImpl) FindByPONumber(tx *gorm.DB, poNumber string) (*models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	if err := r.useDB(tx).Where("po_number = ?", poNumber).First(&po).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}
	return &po, nil
}

// ---------- Counts ----------

func (r *PurchaseOrderRepositoryImpl) CountAllThisMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.PurchaseOrder{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())").
		Count(&count).Error
	return count, err
}

func (r *PurchaseOrderRepositoryImpl) CountAllLastMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.PurchaseOrder{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Count(&count).Error
	return count, err
}

func (r *PurchaseOrderRepositoryImpl) CountByStatus(tx *gorm.DB, status string) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.PurchaseOrder{}).
		Where("po_status = ?", status).
		Count(&count).Error
	return count, err
}

func (r *PurchaseOrderRepositoryImpl) CountByPaymentStatus(tx *gorm.DB, paymentStatus string) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.PurchaseOrder{}).
		Where("payment_status = ?", paymentStatus).
		Count(&count).Error
	return count, err
}

// ---------- Mutations ----------

func (r *PurchaseOrderRepositoryImpl) Insert(tx *gorm.DB, po *models.PurchaseOrder) (*models.PurchaseOrder, error) {
	if po.ID == uuid.Nil {
		return nil, fmt.Errorf("purchase order ID cannot be empty")
	}
	if err := r.useDB(tx).Create(po).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}
	return po, nil
}

func (r *PurchaseOrderRepositoryImpl) Update(tx *gorm.DB, po *models.PurchaseOrder) (*models.PurchaseOrder, error) {
	if po.ID == uuid.Nil {
		return nil, fmt.Errorf("purchase order ID cannot be empty")
	}
	if err := r.useDB(tx).Save(po).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}
	return po, nil
}

func (r *PurchaseOrderRepositoryImpl) UpdateStatus(tx *gorm.DB, poId string, poStatus, paymentStatus string) error {
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

	if err := r.useDB(tx).Model(&models.PurchaseOrder{}).
		Where("id = ?", poId).
		Updates(updates).Error; err != nil {
		return HandleDatabaseError(err, "purchase_order")
	}
	return nil
}

func (r *PurchaseOrderRepositoryImpl) Delete(tx *gorm.DB, poId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var po models.PurchaseOrder
	if err := db.Unscoped().First(&po, "id = ?", poId).Error; err != nil {
		return HandleDatabaseError(err, "purchase_order")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&po).Error; err != nil {
			return HandleDatabaseError(err, "purchase_order")
		}
	} else {
		if err := db.Delete(&po).Error; err != nil {
			return HandleDatabaseError(err, "purchase_order")
		}
	}
	return nil
}

func (r *PurchaseOrderRepositoryImpl) Restore(tx *gorm.DB, poId string) (*models.PurchaseOrder, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.PurchaseOrder{}).
		Where("id = ?", poId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}

	var restored models.PurchaseOrder
	if err := db.
		Preload("Supplier").
		Preload("PurchaseOrderItems").
		Preload("PurchaseOrderItems.Item").
		Preload("PurchaseOrderItems.Item.Category").
		Preload("Payments").
		Preload("Payments.Invoice").
		First(&restored, "id = ?", poId).Error; err != nil {
		return nil, HandleDatabaseError(err, "purchase_order")
	}
	return &restored, nil
}

// ---------- Utilities ----------

func (r *PurchaseOrderRepositoryImpl) GenerateNextPONumber(tx *gorm.DB) (string, error) {
	var lastPO models.PurchaseOrder
	currentYear := time.Now().Year()
	prefix := fmt.Sprintf("PO-%d-", currentYear)

	db := r.useDB(tx)

	err := db.Where("po_number LIKE ?", prefix+"%").
		Order("po_number DESC").
		First(&lastPO).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}

	nextNumber := 1
	if err != gorm.ErrRecordNotFound {
		parts := strings.Split(lastPO.PONumber, "-")
		if len(parts) >= 3 {
			var parsed int
			if n, scanErr := fmt.Sscanf(parts[2], "%d", &parsed); scanErr == nil && n == 1 {
				nextNumber = parsed + 1
			}
		}
	}

	return fmt.Sprintf("%s%04d", prefix, nextNumber), nil
}
