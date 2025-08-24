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

type SupplierRepository interface {
	FindAll(tx *gorm.DB) ([]models.Supplier, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Supplier, int64, error)
	FindById(tx *gorm.DB, supplierId string, includeTrashed bool) (*models.Supplier, error)
	FindByName(tx *gorm.DB, supplierName string) (*models.Supplier, error)
	FindByCode(tx *gorm.DB, supplierCode string) (*models.Supplier, error)
	CountAllThisMonth(tx *gorm.DB) (int64, error)
	CountAllLastMonth(tx *gorm.DB) (int64, error)
	Insert(tx *gorm.DB, supplier *models.Supplier) (*models.Supplier, error)
	Update(tx *gorm.DB, supplier *models.Supplier) (*models.Supplier, error)
	Delete(tx *gorm.DB, supplierId string, isHardDelete bool) error
	Restore(tx *gorm.DB, supplierId string) (*models.Supplier, error)
}

// ==============================
// Implementation
// ==============================

type SupplierRepositoryImpl struct {
	DB *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) *SupplierRepositoryImpl {
	return &SupplierRepositoryImpl{DB: db}
}

func (r *SupplierRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *SupplierRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Supplier, error) {
	var suppliers []models.Supplier
	if err := r.useDB(tx).Unscoped().Find(&suppliers).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return suppliers, nil
}

func (r *SupplierRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Supplier, int64, error) {
	var (
		suppliers   []models.Supplier
		totalCount  int64
	)
	query := r.useDB(tx).Unscoped()

	switch req.Status {
	case "active":
		query = query.Where("deleted_at IS NULL")
	case "deleted":
		query = query.Where("deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		query = query.Where("deleted_at IS NULL")
	}

	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where(`
			LOWER(name) LIKE ? OR
			LOWER(code) LIKE ? OR
			LOWER(email) LIKE ? OR
			LOWER(phone) LIKE ? OR
			LOWER(contact_person) LIKE ?
		`, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if err := query.Model(&models.Supplier{}).Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "supplier")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&suppliers).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "supplier")
	}

	return suppliers, totalCount, nil
}

func (r *SupplierRepositoryImpl) FindById(tx *gorm.DB, supplierId string, includeTrashed bool) (*models.Supplier, error) {
	var supplier models.Supplier
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.First(&supplier, "id = ?", supplierId).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return &supplier, nil
}

func (r *SupplierRepositoryImpl) FindByName(tx *gorm.DB, supplierName string) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := r.useDB(tx).Where("name = ?", supplierName).First(&supplier).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return &supplier, nil
}

func (r *SupplierRepositoryImpl) FindByCode(tx *gorm.DB, supplierCode string) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := r.useDB(tx).Where("code = ?", supplierCode).First(&supplier).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return &supplier, nil
}

// ---------- Counts ----------

func (r *SupplierRepositoryImpl) CountAllThisMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.Supplier{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())").
		Count(&count).Error
	return count, err
}

func (r *SupplierRepositoryImpl) CountAllLastMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.Supplier{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Count(&count).Error
	return count, err
}

// ---------- Mutations ----------

func (r *SupplierRepositoryImpl) Insert(tx *gorm.DB, supplier *models.Supplier) (*models.Supplier, error) {
	if supplier.ID == uuid.Nil {
		return nil, fmt.Errorf("supplier ID cannot be empty")
	}
	if err := r.useDB(tx).Create(supplier).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return supplier, nil
}

func (r *SupplierRepositoryImpl) Update(tx *gorm.DB, supplier *models.Supplier) (*models.Supplier, error) {
	if supplier.ID == uuid.Nil {
		return nil, fmt.Errorf("supplier ID cannot be empty")
	}
	if err := r.useDB(tx).Save(supplier).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return supplier, nil
}

func (r *SupplierRepositoryImpl) Delete(tx *gorm.DB, supplierId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var supplier models.Supplier
	if err := db.Unscoped().First(&supplier, "id = ?", supplierId).Error; err != nil {
		return HandleDatabaseError(err, "supplier")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&supplier).Error; err != nil {
			return HandleDatabaseError(err, "supplier")
		}
	} else {
		if err := db.Delete(&supplier).Error; err != nil {
			return HandleDatabaseError(err, "supplier")
		}
	}
	return nil
}

func (r *SupplierRepositoryImpl) Restore(tx *gorm.DB, supplierId string) (*models.Supplier, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.Supplier{}).
		Where("id = ?", supplierId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}

	var restored models.Supplier
	if err := db.First(&restored, "id = ?", supplierId).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return &restored, nil
}
