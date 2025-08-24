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

type CustomerRepository interface {
	FindAll(tx *gorm.DB) ([]models.Customer, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Customer, int64, error)
	FindById(tx *gorm.DB, customerId string, includeTrashed bool) (*models.Customer, error)
	FindByName(tx *gorm.DB, customerName string) (*models.Customer, error)
	FindByNomor(tx *gorm.DB, customerNomor string) (*models.Customer, error)
	Insert(tx *gorm.DB, customer *models.Customer) (*models.Customer, error)
	Update(tx *gorm.DB, customer *models.Customer) (*models.Customer, error)
	Delete(tx *gorm.DB, customerId string, isHardDelete bool) error
	Restore(tx *gorm.DB, customerId string) (*models.Customer, error)
}

// ==============================
// Implementation
// ==============================

type CustomerRepositoryImpl struct {
	DB *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepositoryImpl {
	return &CustomerRepositoryImpl{DB: db}
}

func (r *CustomerRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *CustomerRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Customer, error) {
	var customers []models.Customer
	if err := r.useDB(tx).
		Unscoped().
		Preload("Area").
		Preload("CustomerType").
		Find(&customers).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer")
	}
	return customers, nil
}

func (r *CustomerRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Customer, int64, error) {
	var (
		customers  []models.Customer
		totalCount  int64
	)

	query := r.useDB(tx).
		Unscoped().
		Model(&models.Customer{}).
		Preload("Area").
		Preload("CustomerType")

	switch req.Status {
	case "active":
		query = query.Where("customers.deleted_at IS NULL")
	case "deleted":
		query = query.Where("customers.deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		query = query.Where("customers.deleted_at IS NULL")
	}

	if req.AreaID != "" {
		if areaUUID, err := uuid.Parse(req.AreaID); err == nil {
			query = query.Where("customers.area_id = ?", areaUUID)
		}
	}
	if req.CustomerTypeID != "" {
		if ftUUID, err := uuid.Parse(req.CustomerTypeID); err == nil {
			query = query.Where("customers.customer_type_id = ?", ftUUID)
		}
	}

	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.
			Joins("LEFT JOIN areas a ON a.id = customers.area_id").
			Joins("LEFT JOIN customer_types ft ON ft.id = customers.customer_type_id").
			Where(`
				LOWER(customers.name) LIKE ? OR
				LOWER(customers.nomor) LIKE ? OR
				LOWER(a.name)          LIKE ? OR
				LOWER(ft.name)         LIKE ?
			`, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if err := query.
		Model(&models.Customer{}).
		Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "customer")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.
		Offset(offset).
		Limit(req.Limit).
		Order("customers.created_at DESC").
		Find(&customers).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "customer")
	}

	return customers, totalCount, nil
}

func (r *CustomerRepositoryImpl) FindById(tx *gorm.DB, customerId string, includeTrashed bool) (*models.Customer, error) {
	var customer models.Customer
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Area").
		Preload("CustomerType").
		First(&customer, "id = ?", customerId).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer")
	}
	return &customer, nil
}

func (r *CustomerRepositoryImpl) FindByName(tx *gorm.DB, customerName string) (*models.Customer, error) {
	var customer models.Customer
	if err := r.useDB(tx).
		Where("name = ?", customerName).
		First(&customer).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer")
	}
	return &customer, nil
}

func (r *CustomerRepositoryImpl) FindByNomor(tx *gorm.DB, customerNomor string) (*models.Customer, error) {
	var customer models.Customer
	if err := r.useDB(tx).
		Where("nomor = ?", customerNomor).
		First(&customer).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer")
	}
	return &customer, nil
}

// ---------- Mutations ----------

func (r *CustomerRepositoryImpl) Insert(tx *gorm.DB, customer *models.Customer) (*models.Customer, error) {
	if customer.ID == uuid.Nil {
		return nil, fmt.Errorf("customer ID cannot be empty")
	}
	if err := r.useDB(tx).Create(customer).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer")
	}
	return customer, nil
}

func (r *CustomerRepositoryImpl) Update(tx *gorm.DB, customer *models.Customer) (*models.Customer, error) {
	if customer.ID == uuid.Nil {
		return nil, fmt.Errorf("customer ID cannot be empty")
	}
	if err := r.useDB(tx).Save(customer).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer")
	}
	return customer, nil
}

func (r *CustomerRepositoryImpl) Delete(tx *gorm.DB, customerId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var customer models.Customer
	if err := db.Unscoped().First(&customer, "id = ?", customerId).Error; err != nil {
		return HandleDatabaseError(err, "customer")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&customer).Error; err != nil {
			return HandleDatabaseError(err, "customer")
		}
	} else {
		if err := db.Delete(&customer).Error; err != nil {
			return HandleDatabaseError(err, "customer")
		}
	}
	return nil
}

func (r *CustomerRepositoryImpl) Restore(tx *gorm.DB, customerId string) (*models.Customer, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.Customer{}).
		Where("id = ?", customerId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer")
	}

	var restored models.Customer
	if err := db.
		Preload("Area").
		Preload("CustomerType").
		First(&restored, "id = ?", customerId).Error; err != nil {
		return nil, HandleDatabaseError(err, "customer")
	}
	return &restored, nil
}
