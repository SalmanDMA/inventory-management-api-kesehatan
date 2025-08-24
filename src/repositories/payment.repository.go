package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type PaymentRepository interface {
	FindAll(tx *gorm.DB) ([]models.Payment, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Payment, int64, error)
	FindById(tx *gorm.DB, paymentId string, includeTrashed bool) (*models.Payment, error)
	FindByPurchaseOrderId(tx *gorm.DB, poId string) ([]models.Payment, error)
	FindBySalesOrderId(tx *gorm.DB, soId string) ([]models.Payment, error)
	Insert(tx *gorm.DB, payment *models.Payment) (*models.Payment, error)
	Update(tx *gorm.DB, payment *models.Payment) (*models.Payment, error)
	Delete(tx *gorm.DB, paymentId string, isHardDelete bool) error
}

// ==============================
// Implementation
// ==============================

type PaymentRepositoryImpl struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepositoryImpl {
	return &PaymentRepositoryImpl{DB: db}
}

func (r *PaymentRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *PaymentRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Payment, error) {
	var payments []models.Payment
	db := r.useDB(tx).Unscoped().
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Supplier").
		Preload("Invoice")

	if err := db.Find(&payments).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payments, nil
}

func (r *PaymentRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Payment, int64, error) {
	var (
		payments   []models.Payment
		totalCount int64
	)

	query := r.useDB(tx).Unscoped().
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Supplier").
		Preload("Invoice")

	switch req.Status {
	case "active":
		query = query.Where("payments.deleted_at IS NULL")
	case "deleted":
		query = query.Where("payments.deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		query = query.Where("payments.deleted_at IS NULL")
	}

	if req.PurchaseOrderID != "" {
		if poUUID, err := uuid.Parse(req.PurchaseOrderID); err == nil {
			query = query.Where("purchase_order_id = ?", poUUID)
		}
	}
	if req.PaymentType != "" {
		query = query.Where("payment_type = ?", req.PaymentType)
	}
	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		query = query.
			Joins("LEFT JOIN purchase_orders ON purchase_orders.id = payments.purchase_order_id").
			Joins("LEFT JOIN suppliers ON suppliers.id = purchase_orders.supplier_id").
			Where(`
				LOWER(payments.reference_number) LIKE ? OR
				LOWER(payments.payment_method) LIKE ? OR
				LOWER(COALESCE(payments.notes, '')) LIKE ? OR
				LOWER(purchase_orders.po_number) LIKE ? OR
				LOWER(suppliers.name) LIKE ?
			`, p, p, p, p, p)
	}

	if err := query.Model(&models.Payment{}).Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "payment")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&payments).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "payment")
	}

	return payments, totalCount, nil
}

func (r *PaymentRepositoryImpl) FindById(tx *gorm.DB, paymentId string, includeTrashed bool) (*models.Payment, error) {
	var payment models.Payment
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Supplier").
		Preload("Invoice").
		First(&payment, "id = ?", paymentId).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return &payment, nil
}

func (r *PaymentRepositoryImpl) FindByPurchaseOrderId(tx *gorm.DB, poId string) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.useDB(tx).
		Preload("Invoice").
		Where("purchase_order_id = ?", poId).
		Find(&payments).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payments, nil
}

func (r *PaymentRepositoryImpl) FindBySalesOrderId(tx *gorm.DB, soId string) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.useDB(tx).
		Preload("Invoice").
		Where("sales_order_id = ?", soId).
		Find(&payments).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payments, nil
}

// ---------- Mutations ----------

func (r *PaymentRepositoryImpl) Insert(tx *gorm.DB, payment *models.Payment) (*models.Payment, error) {
	if payment.ID == uuid.Nil {
		return nil, fmt.Errorf("payment ID cannot be empty")
	}
	if err := r.useDB(tx).Create(payment).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payment, nil
}

func (r *PaymentRepositoryImpl) Update(tx *gorm.DB, payment *models.Payment) (*models.Payment, error) {
	if payment.ID == uuid.Nil {
		return nil, fmt.Errorf("payment ID cannot be empty")
	}
	if err := r.useDB(tx).Save(payment).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payment, nil
}

func (r *PaymentRepositoryImpl) Delete(tx *gorm.DB, paymentId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var payment models.Payment
	if err := db.Unscoped().First(&payment, "id = ?", paymentId).Error; err != nil {
		return HandleDatabaseError(err, "payment")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&payment).Error; err != nil {
			return HandleDatabaseError(err, "payment")
		}
	} else {
		if err := db.Delete(&payment).Error; err != nil {
			return HandleDatabaseError(err, "payment")
		}
	}
	return nil
}
