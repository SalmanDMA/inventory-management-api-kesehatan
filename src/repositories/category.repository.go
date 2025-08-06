package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	FindAll() ([]models.Category, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.Category, int64, error)
	FindById(categoryId string, isSoftDelete bool) (*models.Category, error)
	FindByName(categoryName string) (*models.Category, error)
	Insert(category *models.Category) (*models.Category, error)
	Update(category *models.Category) (*models.Category, error)
	Delete(categoryId string, isHardDelete bool) error
	Restore(category *models.Category, categoryId string) (*models.Category, error)
}

type CategoryRepositoryImpl struct{
	DB *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{DB: db}
}

func (r *CategoryRepositoryImpl) FindAll() ([]models.Category, error) {
	var categories []models.Category
	if err := r.DB.
	 Unscoped().
		Preload("Items").
	 Find(&categories).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return categories, nil
}

func (r *CategoryRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.Category, int64, error) {
	var categories []models.Category
	var totalCount int64

	query := r.DB.Unscoped().Preload("Items")

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
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(color) LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Model(&models.Category{}).Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "category")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&categories).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "category")
	}

	return categories, totalCount, nil
}

func (r *CategoryRepositoryImpl) FindById(categoryId string, isSoftDelete bool) (*models.Category, error) {
	var category *models.Category
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.
	 Preload("Items").
	 First(&category, "id = ?", categoryId).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	
	return category, nil
}

func (r *CategoryRepositoryImpl) FindByName(categoryName string) (*models.Category, error) {
	var category *models.Category
	if err := r.DB.Where("name = ?", categoryName).First(&category).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return category, nil
}

func (r *CategoryRepositoryImpl) Insert(category *models.Category) (*models.Category, error) {

	if category.ID == uuid.Nil {
		return nil, fmt.Errorf("category ID cannot be empty")
	}

	if err := r.DB.Create(&category).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return category, nil
}

func (r *CategoryRepositoryImpl) Update(category *models.Category) (*models.Category, error) {

	if category.ID == uuid.Nil {
		return nil, fmt.Errorf("category ID cannot be empty")
	}

	if err := r.DB.Save(&category).Error; err != nil {
		return nil, HandleDatabaseError(err, "category")
	}
	return category, nil
}

func (r *CategoryRepositoryImpl) Delete(categoryId string, isHardDelete bool) error {
	var category *models.Category

	if err := r.DB.Unscoped().First(&category, "id = ?", categoryId).Error; err != nil {
		return HandleDatabaseError(err, "category")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&category).Error; err != nil {
			return HandleDatabaseError(err, "category")
		}
	} else {
		if err := r.DB.Delete(&category).Error; err != nil {
			return HandleDatabaseError(err, "category")
		}
	}
	return nil
}

func (r *CategoryRepositoryImpl) Restore(category *models.Category, categoryId string) (*models.Category, error) {
	if err := r.DB.Unscoped().Model(category).Where("id = ?", categoryId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredCategory *models.Category
	if err := r.DB.Unscoped().First(&restoredCategory, "id = ?", categoryId).Error; err != nil {
		return nil, err
	}
	
	return restoredCategory, nil
}