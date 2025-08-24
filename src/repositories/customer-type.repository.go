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

type CustomerTypeRepository interface {
	FindAll(tx *gorm.DB) ([]models.CustomerType, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.CustomerType, int64, error)
	FindById(tx *gorm.DB, customerTypeId string, includeTrashed bool) (*models.CustomerType, error)
	FindByName(tx *gorm.DB, customerTypeName string) (*models.CustomerType, error)
	Insert(tx *gorm.DB, customerType *models.CustomerType) (*models.CustomerType, error)
	Update(tx *gorm.DB, customerType *models.CustomerType) (*models.CustomerType, error)
	Delete(tx *gorm.DB, customerTypeId string, isHardDelete bool) error
	Restore(tx *gorm.DB, customerTypeId string) (*models.CustomerType, error)
}

// ==============================
// Implementation
// ==============================

type CustomerTypeRepositoryImpl struct {
	DB *gorm.DB
}

func NewCustomerTypeRepository(db *gorm.DB) *CustomerTypeRepositoryImpl {
	return &CustomerTypeRepositoryImpl{DB: db}
}

func (r *CustomerTypeRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *CustomerTypeRepositoryImpl) FindAll(tx *gorm.DB) ([]models.CustomerType, error) {
	var customerTypes []models.CustomerType
	if err := r.useDB(tx).
		Unscoped().
		Find(&customerTypes).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer_type")
	}
	return customerTypes, nil
}

func (r *CustomerTypeRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.CustomerType, int64, error) {
	var (
		customerTypes []models.CustomerType
		totalCount    int64
	)

	query := r.useDB(tx).
		Unscoped().
		Model(&models.CustomerType{})

	switch req.Status {
	case "active":
		query = query.Where("customer_types.deleted_at IS NULL")
	case "deleted":
		query = query.Where("customer_types.deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		query = query.Where("customer_types.deleted_at IS NULL")
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		query = query.Where(`
			LOWER(customer_types.name) LIKE ? OR
			LOWER(COALESCE(customer_types.color, '')) LIKE ? OR
			LOWER(COALESCE(customer_types.description, '')) LIKE ?
		`, p, p, p)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "customer_type")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.
		Order("customer_types.created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&customerTypes).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "customer_type")
	}

	return customerTypes, totalCount, nil
}

func (r *CustomerTypeRepositoryImpl) FindById(tx *gorm.DB, customerTypeId string, includeTrashed bool) (*models.CustomerType, error) {
	var ft models.CustomerType
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.First(&ft, "id = ?", customerTypeId).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer_type")
	}
	return &ft, nil
}

func (r *CustomerTypeRepositoryImpl) FindByName(tx *gorm.DB, customerTypeName string) (*models.CustomerType, error) {
	var ft models.CustomerType
	if err := r.useDB(tx).
		Where("name = ?", customerTypeName).
		First(&ft).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer_type")
	}
	return &ft, nil
}

// ---------- Mutations ----------

func (r *CustomerTypeRepositoryImpl) Insert(tx *gorm.DB, customerType *models.CustomerType) (*models.CustomerType, error) {
	if customerType.ID == uuid.Nil {
		return nil, fmt.Errorf("customerType ID cannot be empty")
	}
	if err := r.useDB(tx).Create(customerType).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer_type")
	}
	return customerType, nil
}

func (r *CustomerTypeRepositoryImpl) Update(tx *gorm.DB, customerType *models.CustomerType) (*models.CustomerType, error) {
	if customerType.ID == uuid.Nil {
		return nil, fmt.Errorf("customerType ID cannot be empty")
	}
	if err := r.useDB(tx).Save(customerType).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer_type")
	}
	return customerType, nil
}

func (r *CustomerTypeRepositoryImpl) Delete(tx *gorm.DB, customerTypeId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var ft models.CustomerType
	if err := db.Unscoped().First(&ft, "id = ?", customerTypeId).Error; err != nil {
		return HandleDatabaseError(err, "customer_type")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&ft).Error; err != nil {
			return HandleDatabaseError(err, "customer_type")
		}
	} else {
		if err := db.Delete(&ft).Error; err != nil {
			return HandleDatabaseError(err, "customer_type")
		}
	}
	return nil
}

func (r *CustomerTypeRepositoryImpl) Restore(tx *gorm.DB, customerTypeId string) (*models.CustomerType, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.CustomerType{}).
		Where("id = ?", customerTypeId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer_type")
	}

	var restored models.CustomerType
	if err := db.First(&restored, "id = ?", customerTypeId).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer_type")
	}
	return &restored, nil
}
