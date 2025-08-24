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

type CategoryRepository interface {
	FindAll(tx *gorm.DB) ([]models.Category, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Category, int64, error)
	FindById(tx *gorm.DB, categoryId string, includeTrashed bool) (*models.Category, error)
	FindByName(tx *gorm.DB, categoryName string) (*models.Category, error)
	Insert(tx *gorm.DB, category *models.Category) (*models.Category, error)
	Update(tx *gorm.DB, category *models.Category) (*models.Category, error)
	Delete(tx *gorm.DB, categoryId string, isHardDelete bool) error
	Restore(tx *gorm.DB, categoryId string) (*models.Category, error)
}

// ==============================
// Implementation
// ==============================

type CategoryRepositoryImpl struct {
	DB *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{DB: db}
}

func (r *CategoryRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *CategoryRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Category, error) {
	var categories []models.Category
	if err := r.useDB(tx).
		Unscoped().
		Preload("Items").
		Find(&categories).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return categories, nil
}

func (r *CategoryRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Category, int64, error) {
	var (
		categories  []models.Category
		totalCount  int64
	)

	q := r.useDB(tx).
		Unscoped().
		Model(&models.Category{}).
		Preload("Items")

	switch req.Status {
	case "active":
		q = q.Where("categories.deleted_at IS NULL")
	case "deleted":
		q = q.Where("categories.deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		q = q.Where("categories.deleted_at IS NULL")
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		q = q.Where(`
			LOWER(categories.name) LIKE ? OR
			LOWER(COALESCE(categories.description, '')) LIKE ? OR
			LOWER(COALESCE(categories.color, '')) LIKE ?
		`, p, p, p)
	}

	if err := q.Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "category")
	}

	offset := (req.Page - 1) * req.Limit
	if err := q.
		Offset(offset).
		Limit(req.Limit).
		Order("categories.created_at DESC").
		Find(&categories).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "category")
	}

	return categories, totalCount, nil
}

func (r *CategoryRepositoryImpl) FindById(tx *gorm.DB, categoryId string, includeTrashed bool) (*models.Category, error) {
	var cat models.Category
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Items").
		First(&cat, "id = ?", categoryId).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return &cat, nil
}

func (r *CategoryRepositoryImpl) FindByName(tx *gorm.DB, categoryName string) (*models.Category, error) {
	var cat models.Category
	if err := r.useDB(tx).
		Where("name = ?", categoryName).
		First(&cat).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return &cat, nil
}

// ---------- Mutations ----------

func (r *CategoryRepositoryImpl) Insert(tx *gorm.DB, category *models.Category) (*models.Category, error) {
	if category.ID == uuid.Nil {
		return nil, fmt.Errorf("category ID cannot be empty")
	}
	if err := r.useDB(tx).Create(category).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return category, nil
}

func (r *CategoryRepositoryImpl) Update(tx *gorm.DB, category *models.Category) (*models.Category, error) {
	if category.ID == uuid.Nil {
		return nil, fmt.Errorf("category ID cannot be empty")
	}
	if err := r.useDB(tx).Save(category).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return category, nil
}

func (r *CategoryRepositoryImpl) Delete(tx *gorm.DB, categoryId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var cat models.Category
	if err := db.Unscoped().First(&cat, "id = ?", categoryId).Error; err != nil {
		return HandleDatabaseError(err, "category")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&cat).Error; err != nil {
			return HandleDatabaseError(err, "category")
		}
	} else {
		if err := db.Delete(&cat).Error; err != nil {
			return HandleDatabaseError(err, "category")
		}
	}
	return nil
}

func (r *CategoryRepositoryImpl) Restore(tx *gorm.DB, categoryId string) (*models.Category, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.Category{}).
		Where("id = ?", categoryId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}

	var restored models.Category
	if err := db.
		Preload("Items").
		First(&restored, "id = ?", categoryId).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return &restored, nil
}
