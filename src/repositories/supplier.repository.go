package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SupplierRepository interface {
	FindAll() ([]models.Supplier, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.Supplier, int64, error)
	FindById(supplierId string, includeTrashed bool) (*models.Supplier, error)
	FindByName(supplierName string) (*models.Supplier, error)
	FindByCode(supplierCode string) (*models.Supplier, error)
	CountAllThisMonth() (int64, error)
	CountAllLastMonth() (int64, error)
	Insert(supplier *models.Supplier) (*models.Supplier, error)
	Update(supplier *models.Supplier) (*models.Supplier, error)
	Delete(supplierId string, isHardDelete bool) error
	Restore(supplier *models.Supplier, supplierId string) (*models.Supplier, error)
}

// SupplierRepositoryImpl implementation
type SupplierRepositoryImpl struct {
	DB *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) *SupplierRepositoryImpl {
	return &SupplierRepositoryImpl{DB: db}
}

func (r *SupplierRepositoryImpl) FindAll() ([]models.Supplier, error) {
	var suppliers []models.Supplier
	if err := r.DB.Unscoped().Find(&suppliers).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return suppliers, nil
}

func (r *SupplierRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.Supplier, int64, error) {
	var suppliers []models.Supplier
	var totalCount int64

	query := r.DB.Unscoped()

	switch req.Status {
	case "active":
		query = query.Where("deleted_at IS NULL")
	case "deleted":
		query = query.Where("deleted_at IS NOT NULL")
	case "all":
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

func (r *SupplierRepositoryImpl) FindById(supplierId string, includeTrashed bool) (*models.Supplier, error) {
	var supplier *models.Supplier
	db := r.DB

	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.First(&supplier, "id = ?", supplierId).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	
	return supplier, nil
}

func (r *SupplierRepositoryImpl) FindByName(supplierName string) (*models.Supplier, error) {
	var supplier *models.Supplier
	if err := r.DB.Where("name = ?", supplierName).First(&supplier).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return supplier, nil
}

func (r *SupplierRepositoryImpl) FindByCode(supplierCode string) (*models.Supplier, error) {
	var supplier *models.Supplier
	if err := r.DB.Where("code = ?", supplierCode).First(&supplier).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return supplier, nil
}

func (r *SupplierRepositoryImpl) CountAllThisMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.Supplier{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())").
		Count(&count).Error
	return count, err
}

func (r *SupplierRepositoryImpl) CountAllLastMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.Supplier{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Count(&count).Error
	return count, err
}

func (r *SupplierRepositoryImpl) Insert(supplier *models.Supplier) (*models.Supplier, error) {
	if supplier.ID == uuid.Nil {
		return nil, fmt.Errorf("supplier ID cannot be empty")
	}

	if err := r.DB.Create(&supplier).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return supplier, nil
}

func (r *SupplierRepositoryImpl) Update(supplier *models.Supplier) (*models.Supplier, error) {
	if supplier.ID == uuid.Nil {
		return nil, fmt.Errorf("supplier ID cannot be empty")
	}

	if err := r.DB.Save(&supplier).Error; err != nil {
		return nil, HandleDatabaseError(err, "supplier")
	}
	return supplier, nil
}

func (r *SupplierRepositoryImpl) Delete(supplierId string, isHardDelete bool) error {
	var supplier *models.Supplier

	if err := r.DB.Unscoped().First(&supplier, "id = ?", supplierId).Error; err != nil {
		return HandleDatabaseError(err, "supplier")
	}

	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&supplier).Error; err != nil {
			return HandleDatabaseError(err, "supplier")
		}
	} else {
		if err := r.DB.Delete(&supplier).Error; err != nil {
			return HandleDatabaseError(err, "supplier")
		}
	}
	return nil
}

func (r *SupplierRepositoryImpl) Restore(supplier *models.Supplier, supplierId string) (*models.Supplier, error) {
	if err := r.DB.Unscoped().Model(supplier).Where("id = ?", supplierId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredSupplier *models.Supplier
	if err := r.DB.Unscoped().First(&restoredSupplier, "id = ?", supplierId).Error; err != nil {
		return nil, err
	}

	return restoredSupplier, nil
}