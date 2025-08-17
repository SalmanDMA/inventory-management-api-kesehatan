package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	FindAll() ([]models.Payment, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.Payment, int64, error)
	FindById(paymentId string, isSoftDelete bool) (*models.Payment, error)
	FindByPurchaseOrderId(poId string) ([]models.Payment, error)
	FindBySalesOrderId(soId string) ([]models.Payment, error)
	Insert(payment *models.Payment) (*models.Payment, error)
	Update(payment *models.Payment) (*models.Payment, error)
	Delete(paymentId string, isHardDelete bool) error
}

type PaymentRepositoryImpl struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepositoryImpl {
	return &PaymentRepositoryImpl{DB: db}
}

func (r *PaymentRepositoryImpl) FindAll() ([]models.Payment, error) {
	var payments []models.Payment
	db := r.DB.Unscoped().
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Supplier").
		Preload("Invoice")

	if err := db.Find(&payments).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payments, nil
}

func (r *PaymentRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.Payment, int64, error) {
	var payments []models.Payment
	var totalCount int64

	query := r.DB.Unscoped().
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Supplier").
		Preload("Invoice")

	switch req.Status {
	case "active":
		query = query.Where("payments.deleted_at IS NULL")
	case "deleted":
		query = query.Where("payments.deleted_at IS NOT NULL")
	case "all":
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

	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.Joins("LEFT JOIN purchase_orders ON purchase_orders.id = payments.purchase_order_id").
			Joins("LEFT JOIN suppliers ON suppliers.id = purchase_orders.supplier_id").
			Where(`
				LOWER(payments.reference_number) LIKE ? OR
				LOWER(payments.payment_method) LIKE ? OR
				LOWER(payments.notes) LIKE ? OR
				LOWER(purchase_orders.po_number) LIKE ? OR
				LOWER(suppliers.name) LIKE ?
			`,
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
		)
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

func (r *PaymentRepositoryImpl) FindById(paymentId string, isSoftDelete bool) (*models.Payment, error) {
	var payment *models.Payment
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.
		Preload("PurchaseOrder").
		Preload("PurchaseOrder.Supplier").
		Preload("Invoice").
		First(&payment, "id = ?", paymentId).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	
	return payment, nil
}

func (r *PaymentRepositoryImpl) FindByPurchaseOrderId(poId string) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.DB.
		Preload("Invoice").
		Where("purchase_order_id = ?", poId).
		Find(&payments).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payments, nil
}

func (r *PaymentRepositoryImpl) FindBySalesOrderId(soId string) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.DB.
		Preload("Invoice").
		Where("sales_order_id = ?", soId).
		Find(&payments).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payments, nil
}

func (r *PaymentRepositoryImpl) Insert(payment *models.Payment) (*models.Payment, error) {
	if payment.ID == uuid.Nil {
		return nil, fmt.Errorf("payment ID cannot be empty")
	}

	if err := r.DB.Create(&payment).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payment, nil
}

func (r *PaymentRepositoryImpl) Update(payment *models.Payment) (*models.Payment, error) {
	if payment.ID == uuid.Nil {
		return nil, fmt.Errorf("payment ID cannot be empty")
	}

	if err := r.DB.Save(&payment).Error; err != nil {
		return nil, HandleDatabaseError(err, "payment")
	}
	return payment, nil
}

func (r *PaymentRepositoryImpl) Delete(paymentId string, isHardDelete bool) error {
	var payment *models.Payment

	if err := r.DB.Unscoped().First(&payment, "id = ?", paymentId).Error; err != nil {
		return HandleDatabaseError(err, "payment")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&payment).Error; err != nil {
			return HandleDatabaseError(err, "payment")
		}
	} else {
		if err := r.DB.Delete(&payment).Error; err != nil {
			return HandleDatabaseError(err, "payment")
		}
	}
	return nil
}

