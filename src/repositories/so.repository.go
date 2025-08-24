package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type SalesOrderRepository interface {
	FindAll(tx *gorm.DB) ([]models.SalesOrder, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.SalesOrder, int64, error)
	FindById(tx *gorm.DB, soId string, includeTrash bool) (*models.SalesOrder, error)
	FindBySONumber(tx *gorm.DB, soNumber string) (*models.SalesOrder, error)

	CountAllThisMonth(tx *gorm.DB) (int64, error)
	CountAllLastMonth(tx *gorm.DB) (int64, error)
	CountActiveThisMonth(tx *gorm.DB) (int64, error)
	CountActiveLastMonth(tx *gorm.DB) (int64, error)
	SumClosedValueThisMonth(tx *gorm.DB) (float64, error)
	SumClosedValueLastMonth(tx *gorm.DB) (float64, error)
	CountByStatus(tx *gorm.DB, status string) (int64, error)
	CountByPaymentStatus(tx *gorm.DB, paymentStatus string) (int64, error)

	Insert(tx *gorm.DB, so *models.SalesOrder) (*models.SalesOrder, error)
	Update(tx *gorm.DB, so *models.SalesOrder) (*models.SalesOrder, error)
	UpdateStatus(tx *gorm.DB, soId string, soStatus, paymentStatus string) error
	Delete(tx *gorm.DB, soId string, isHardDelete bool) error
	Restore(tx *gorm.DB, soId string) (*models.SalesOrder, error)

	GenerateNextSONumber(tx *gorm.DB) (string, error)
}

// ==============================
// Implementation
// ==============================

type SalesOrderRepositoryImpl struct {
	DB *gorm.DB
}

func NewSalesOrderRepository(db *gorm.DB) *SalesOrderRepositoryImpl {
	return &SalesOrderRepositoryImpl{DB: db}
}

func (r *SalesOrderRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *SalesOrderRepositoryImpl) FindAll(tx *gorm.DB) ([]models.SalesOrder, error) {
	var sos []models.SalesOrder
	db := r.useDB(tx).Unscoped().
		Preload("SalesPerson").
		Preload("Customer").
		Preload("SalesOrderItems").
		Preload("SalesOrderItems.Item").
		Preload("SalesOrderItems.Item.Category").
		Preload("Payments").
		Preload("Payments.Invoice")

	if err := db.Find(&sos).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}
	return sos, nil
}

func (r *SalesOrderRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.SalesOrder, int64, error) {
	var (
		sos        []models.SalesOrder
		totalCount int64
	)

	query := r.useDB(tx).Unscoped().
		Preload("SalesPerson").
		Preload("Customer").
		Preload("SalesOrderItems").
		Preload("SalesOrderItems.Item").
		Preload("SalesOrderItems.Item.Category").
		Preload("Payments").
		Preload("Payments.Invoice")

	switch req.Status {
	case "active":
		query = query.Where("sales_orders.deleted_at IS NULL")
	case "deleted":
		query = query.Where("sales_orders.deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		query = query.Where("sales_orders.deleted_at IS NULL")
	}

	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.
			Joins("LEFT JOIN customers ON customers.id = sales_orders.customer_id").
			Joins("LEFT JOIN sales_people ON sales_people.id = sales_orders.sales_person_id").
			Where(`
				LOWER(sales_orders.so_number) LIKE ? OR
				LOWER(sales_orders.notes) LIKE ? OR
				LOWER(customers.name) LIKE ? OR
				LOWER(sales_people.name) LIKE ?
			`,
				searchPattern, searchPattern, searchPattern, searchPattern,
			)
	}

	if req.SOStatus != "" {
		query = query.Where("so_status = ?", req.SOStatus)
	}
	if req.PaymentStatus != "" {
		query = query.Where("payment_status = ?", req.PaymentStatus)
	}
	if req.TermOfPayment != "" {
		query = query.Where("term_of_payment = ?", req.TermOfPayment)
	}
	if req.CustomerID != "" {
		if customerUUID, err := uuid.Parse(req.CustomerID); err == nil {
			query = query.Where("customer_id = ?", customerUUID)
		}
	}
	if req.SalesPersonID != "" {
		if salesPersonUUID, err := uuid.Parse(req.SalesPersonID); err == nil {
			query = query.Where("sales_person_id = ?", salesPersonUUID)
		}
	}

	if err := query.Model(&models.SalesOrder{}).Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "sales_order")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&sos).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "sales_order")
	}

	return sos, totalCount, nil
}

func (r *SalesOrderRepositoryImpl) FindById(tx *gorm.DB, soId string, includeTrash bool) (*models.SalesOrder, error) {
	var so models.SalesOrder
	db := r.useDB(tx)
	if includeTrash {
		db = db.Unscoped()
	}

	if err := db.
		Preload("SalesPerson").
		Preload("Customer").
		Preload("SalesOrderItems").
		Preload("SalesOrderItems.Item").
		Preload("SalesOrderItems.Item.Category").
		Preload("Payments").
		Preload("Payments.Invoice").
		First(&so, "id = ?", soId).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}

	return &so, nil
}

func (r *SalesOrderRepositoryImpl) FindBySONumber(tx *gorm.DB, soNumber string) (*models.SalesOrder, error) {
	var so models.SalesOrder
	if err := r.useDB(tx).Where("so_number = ?", soNumber).First(&so).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}
	return &so, nil
}

// ---------- Counts & Sums ----------

func (r *SalesOrderRepositoryImpl) CountAllThisMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())").
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) CountAllLastMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) CountActiveThisMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', so_date) = DATE_TRUNC('month', NOW())").
		Where("so_status IN ?", []string{"Confirmed", "Shipped", "Delivered"}).
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) CountActiveLastMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', so_date) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Where("so_status IN ?", []string{"Confirmed", "Shipped", "Delivered"}).
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) SumClosedValueThisMonth(tx *gorm.DB) (float64, error) {
	var total sql.NullFloat64
	err := r.useDB(tx).Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', so_date) = DATE_TRUNC('month', NOW())").
		Where("so_status IN ?", []string{"Delivered", "Closed"}).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

func (r *SalesOrderRepositoryImpl) SumClosedValueLastMonth(tx *gorm.DB) (float64, error) {
	var total sql.NullFloat64
	err := r.useDB(tx).Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', so_date) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Where("so_status IN ?", []string{"Delivered", "Closed"}).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

func (r *SalesOrderRepositoryImpl) CountByStatus(tx *gorm.DB, status string) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.SalesOrder{}).
		Where("so_status = ?", status).
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) CountByPaymentStatus(tx *gorm.DB, paymentStatus string) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.SalesOrder{}).
		Where("payment_status = ?", paymentStatus).
		Count(&count).Error
	return count, err
}

// ---------- Mutations ----------

func (r *SalesOrderRepositoryImpl) Insert(tx *gorm.DB, so *models.SalesOrder) (*models.SalesOrder, error) {
	if so.ID == uuid.Nil {
		return nil, fmt.Errorf("sales order ID cannot be empty")
	}
	if err := r.useDB(tx).Create(so).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}
	return so, nil
}

func (r *SalesOrderRepositoryImpl) Update(tx *gorm.DB, so *models.SalesOrder) (*models.SalesOrder, error) {
	if so.ID == uuid.Nil {
		return nil, fmt.Errorf("sales order ID cannot be empty")
	}
	if err := r.useDB(tx).Save(so).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}
	return so, nil
}

func (r *SalesOrderRepositoryImpl) UpdateStatus(tx *gorm.DB, soId string, soStatus, paymentStatus string) error {
	updates := make(map[string]interface{})
	if soStatus != "" {
		updates["so_status"] = soStatus
	}
	if paymentStatus != "" {
		updates["payment_status"] = paymentStatus
	}
	if len(updates) == 0 {
		return fmt.Errorf("no status updates provided")
	}

	if err := r.useDB(tx).Model(&models.SalesOrder{}).
		Where("id = ?", soId).
		Updates(updates).Error; err != nil {
		return HandleDatabaseError(err, "sales_order")
	}
	return nil
}

func (r *SalesOrderRepositoryImpl) Delete(tx *gorm.DB, soId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var so models.SalesOrder
	if err := db.Unscoped().First(&so, "id = ?", soId).Error; err != nil {
		return HandleDatabaseError(err, "sales_order")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&so).Error; err != nil {
			return HandleDatabaseError(err, "sales_order")
		}
	} else {
		if err := db.Delete(&so).Error; err != nil {
			return HandleDatabaseError(err, "sales_order")
		}
	}
	return nil
}

func (r *SalesOrderRepositoryImpl) Restore(tx *gorm.DB, soId string) (*models.SalesOrder, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.SalesOrder{}).
		Where("id = ?", soId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}

	var restored models.SalesOrder
	if err := db.
		Preload("SalesPerson").
		Preload("Customer").
		Preload("SalesOrderItems").
		Preload("SalesOrderItems.Item").
		Preload("SalesOrderItems.Item.Category").
		Preload("Payments").
		Preload("Payments.Invoice").
		First(&restored, "id = ?", soId).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}

	return &restored, nil
}

// ---------- Utilities ----------

func (r *SalesOrderRepositoryImpl) GenerateNextSONumber(tx *gorm.DB) (string, error) {
	var lastSO models.SalesOrder
	currentYear := time.Now().Year()
	prefix := fmt.Sprintf("SO-%d-", currentYear)

	db := r.useDB(tx)

	err := db.Where("so_number LIKE ?", prefix+"%").
		Order("so_number DESC").
		First(&lastSO).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}

	nextNumber := 1
	if err != gorm.ErrRecordNotFound {
		parts := strings.Split(lastSO.SONumber, "-")
		if len(parts) >= 3 {
			var parsed int
			if n, scanErr := fmt.Sscanf(parts[2], "%d", &parsed); scanErr == nil && n == 1 {
				nextNumber = parsed + 1
			}
		}
	}

	return fmt.Sprintf("%s%04d", prefix, nextNumber), nil
}
