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

type SalesOrderRepository interface {
	FindAll() ([]models.SalesOrder, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.SalesOrder, int64, error)
	FindById(soId string, includeTrash bool) (*models.SalesOrder, error)
	FindBySONumber(soNumber string) (*models.SalesOrder, error)
	CountAllThisMonth() (int64, error)
	CountAllLastMonth() (int64, error)
	CountActiveThisMonth() (int64, error)
	CountActiveLastMonth() (int64, error)
	SumClosedValueThisMonth() (float64, error)
	SumClosedValueLastMonth() (float64, error)
	CountByStatus(status string) (int64, error)
	CountByPaymentStatus(paymentStatus string) (int64, error)
	Insert(so *models.SalesOrder) (*models.SalesOrder, error)
	Update(so *models.SalesOrder) (*models.SalesOrder, error)
	UpdateStatus(soId string, soStatus, paymentStatus string) error
	Delete(soId string, isHardDelete bool) error
	Restore(so *models.SalesOrder, soId string) (*models.SalesOrder, error)
	GenerateNextSONumber() (string, error)
}

// SalesOrderRepositoryImpl implementation
type SalesOrderRepositoryImpl struct {
	DB *gorm.DB
}

func NewSalesOrderRepository(db *gorm.DB) *SalesOrderRepositoryImpl {
	return &SalesOrderRepositoryImpl{DB: db}
}

func (r *SalesOrderRepositoryImpl) FindAll() ([]models.SalesOrder, error) {
	var sos []models.SalesOrder
	db := r.DB.Unscoped().
		Preload("SalesPerson").
		Preload("Facility").
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

func (r *SalesOrderRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.SalesOrder, int64, error) {
	var sos []models.SalesOrder
	var totalCount int64

	query := r.DB.Unscoped().
		Preload("SalesPerson").
		Preload("Facility").
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
	default:
		query = query.Where("sales_orders.deleted_at IS NULL")
	}

	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.
			Joins("LEFT JOIN facilities ON facilities.id = sales_orders.facility_id").
			Joins("LEFT JOIN sales_people ON sales_people.id = sales_orders.sales_person_id").
			Where(`
				LOWER(sales_orders.so_number) LIKE ? OR
				LOWER(sales_orders.notes) LIKE ? OR
				LOWER(facilities.name) LIKE ? OR
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

	if req.FacilityID != "" {
		if facilityUUID, err := uuid.Parse(req.FacilityID); err == nil {
			query = query.Where("facility_id = ?", facilityUUID)
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

func (r *SalesOrderRepositoryImpl) FindById(soId string, includeTrash bool) (*models.SalesOrder, error) {
	var so *models.SalesOrder
	db := r.DB

	if includeTrash {
		db = db.Unscoped()
	}

	if err := db.
		Preload("SalesPerson").
		Preload("Facility").
		Preload("SalesOrderItems").
		Preload("SalesOrderItems.Item").
		Preload("SalesOrderItems.Item.Category").
		Preload("Payments").
		Preload("Payments.Invoice").
		First(&so, "id = ?", soId).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}

	return so, nil
}

func (r *SalesOrderRepositoryImpl) FindBySONumber(soNumber string) (*models.SalesOrder, error) {
	var so *models.SalesOrder
	if err := r.DB.Where("so_number = ?", soNumber).First(&so).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}
	return so, nil
}

func (r *SalesOrderRepositoryImpl) CountAllThisMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())").
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) CountAllLastMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Count(&count).Error
	return count, err
}


func (r *SalesOrderRepositoryImpl) CountActiveThisMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', so_date) = DATE_TRUNC('month', NOW())").
		Where("so_status IN ?", []string{"Confirmed", "Shipped", "Delivered"}).
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) CountActiveLastMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', so_date) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Where("so_status IN ?", []string{"Confirmed", "Shipped", "Delivered"}).
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) SumClosedValueThisMonth() (float64, error) {
	var total sql.NullFloat64
	err := r.DB.Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', so_date) = DATE_TRUNC('month', NOW())").
		Where("so_status = ?", "Closed").
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

func (r *SalesOrderRepositoryImpl) SumClosedValueLastMonth() (float64, error) {
	var total sql.NullFloat64
	err := r.DB.Model(&models.SalesOrder{}).
		Where("DATE_TRUNC('month', so_date) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Where("so_status = ?", "Closed").
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

func (r *SalesOrderRepositoryImpl) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.DB.Model(&models.SalesOrder{}).
		Where("so_status = ?", status).
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) CountByPaymentStatus(paymentStatus string) (int64, error) {
	var count int64
	err := r.DB.Model(&models.SalesOrder{}).
		Where("payment_status = ?", paymentStatus).
		Count(&count).Error
	return count, err
}

func (r *SalesOrderRepositoryImpl) Insert(so *models.SalesOrder) (*models.SalesOrder, error) {
	if so.ID == uuid.Nil {
		return nil, fmt.Errorf("sales order ID cannot be empty")
	}

	if err := r.DB.Create(&so).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}
	return so, nil
}

func (r *SalesOrderRepositoryImpl) Update(so *models.SalesOrder) (*models.SalesOrder, error) {
	if so.ID == uuid.Nil {
		return nil, fmt.Errorf("sales order ID cannot be empty")
	}

	if err := r.DB.Save(&so).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_order")
	}
	return so, nil
}

func (r *SalesOrderRepositoryImpl) UpdateStatus(soId string, soStatus, paymentStatus string) error {
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

	if err := r.DB.Model(&models.SalesOrder{}).Where("id = ?", soId).Updates(updates).Error; err != nil {
		return HandleDatabaseError(err, "sales_order")
	}
	return nil
}

func (r *SalesOrderRepositoryImpl) Delete(soId string, isHardDelete bool) error {
	var so *models.SalesOrder

	if err := r.DB.Unscoped().First(&so, "id = ?", soId).Error; err != nil {
		return HandleDatabaseError(err, "sales_order")
	}

	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&so).Error; err != nil {
			return HandleDatabaseError(err, "sales_order")
		}
	} else {
		if err := r.DB.Delete(&so).Error; err != nil {
			return HandleDatabaseError(err, "sales_order")
		}
	}
	return nil
}

func (r *SalesOrderRepositoryImpl) Restore(so *models.SalesOrder, soId string) (*models.SalesOrder, error) {
	if err := r.DB.Unscoped().Model(so).Where("id = ?", soId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredSO *models.SalesOrder
	if err := r.DB.Unscoped().First(&restoredSO, "id = ?", soId).Error; err != nil {
		return nil, err
	}

	return restoredSO, nil
}

func (r *SalesOrderRepositoryImpl) GenerateNextSONumber() (string, error) {
	var lastSO models.SalesOrder
	currentYear := time.Now().Year()
	prefix := fmt.Sprintf("SO-%d-", currentYear)

	err := r.DB.Where("so_number LIKE ?", prefix+"%").
		Order("so_number DESC").
		First(&lastSO).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}

	var nextNumber int = 1
	if err != gorm.ErrRecordNotFound {
		parts := strings.Split(lastSO.SONumber, "-")
		if len(parts) >= 3 {
			if num, parseErr := fmt.Sscanf(parts[2], "%d", &nextNumber); parseErr == nil && num == 1 {
				nextNumber++
			}
		}
	}

	return fmt.Sprintf("%s%04d", prefix, nextNumber), nil
}
