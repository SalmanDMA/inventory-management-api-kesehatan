package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ItemRepository interface {
	FindAll() ([]models.Item, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.Item, int64, error)
	FindById(itemId string, isSoftDelete bool) (*models.Item, error)
	FindByName(itemName string) (*models.Item, error)
	CountAllThisMonth() (int64, error)
	CountAllLastMonth() (int64, error)
	Insert(item *models.Item) (*models.Item, error)
	Update(item *models.Item) (*models.Item, error)
	Delete(itemId string, isHardDelete bool) error
	Restore(item *models.Item, itemId string) (*models.Item, error)
}

type ItemRepositoryImpl struct{
	DB *gorm.DB
}

func NewItemRepository(db *gorm.DB) *ItemRepositoryImpl {
	return &ItemRepositoryImpl{DB: db}
}

func (r *ItemRepositoryImpl) FindAll() ([]models.Item, error) {
	var items []models.Item
	db := r.DB.Unscoped().
		Preload("Image").
		Preload("Category").
		Preload("ItemHistories")

	if err := db.Find(&items).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return items, nil
}

func (r *ItemRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.Item, int64, error) {
	var items []models.Item
	var totalCount int64

	query := r.DB.Unscoped().
		Preload("Image").
		Preload("Category").
		Preload("ItemHistories")

	switch req.Status {
	case "active":
		query = query.Where("deleted_at IS NULL")
	case "deleted":
		query = query.Where("deleted_at IS NOT NULL")
	case "all":
	default:
		query = query.Where("deleted_at IS NULL")
	}

	if req.CategoryID != "" {
		if catUUID, err := uuid.Parse(req.CategoryID); err == nil {
			query = query.Where("category_id = ?", catUUID)
		}
	}

	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.Joins("LEFT JOIN categories ON categories.id = items.category_id").
			Where(`
				LOWER(items.name) LIKE ? OR
				LOWER(items.code) LIKE ? OR
				LOWER(items.description) LIKE ? OR
				CAST(items.price AS TEXT) LIKE ? OR
				CAST(items.stock AS TEXT) LIKE ? OR
				LOWER(categories.name) LIKE ?
			`,
			searchPattern, searchPattern, searchPattern,
			searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Model(&models.Item{}).Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "item")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&items).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "item")
	}

	return items, totalCount, nil
}

func (r *ItemRepositoryImpl) FindById(itemId string, isSoftDelete bool) (*models.Item, error) {
	var item *models.Item
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Image").
	 Preload("Category").
		Preload("ItemHistories").
	 First(&item, "id = ?", itemId).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	
	return item, nil
}

func (r *ItemRepositoryImpl) FindByName(itemName string) (*models.Item, error) {
	var item *models.Item
	if err := r.DB.Where("name = ?", itemName).First(&item).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return item, nil
}

func (r *ItemRepositoryImpl) CountAllThisMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.Item{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())").
		Count(&count).Error
	return count, err
}

func (r *ItemRepositoryImpl) CountAllLastMonth() (int64, error) {
	var count int64
	err := r.DB.Model(&models.Item{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Count(&count).Error
	return count, err
}

func (r *ItemRepositoryImpl) Insert(item *models.Item) (*models.Item, error) {

	if item.ID == uuid.Nil {
		return nil, fmt.Errorf("item ID cannot be empty")
	}

	if err := r.DB.Create(&item).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return item, nil
}

func (r *ItemRepositoryImpl) Update(item *models.Item) (*models.Item, error) {

	if item.ID == uuid.Nil {
		return nil, fmt.Errorf("item ID cannot be empty")
	}

	if err := r.DB.Save(&item).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return item, nil
}

func (r *ItemRepositoryImpl) Delete(itemId string, isHardDelete bool) error {
	var item *models.Item

	if err := r.DB.Unscoped().First(&item, "id = ?", itemId).Error; err != nil {
		return HandleDatabaseError(err, "item")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&item).Error; err != nil {
			return HandleDatabaseError(err, "item")
		}
	} else {
		if err := r.DB.Delete(&item).Error; err != nil {
			return HandleDatabaseError(err, "item")
		}
	}
	return nil
}

func (r *ItemRepositoryImpl) Restore(item *models.Item, itemId string) (*models.Item, error) {
	if err := r.DB.Unscoped().Model(item).Where("id = ?", itemId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredItem *models.Item
	if err := r.DB.Unscoped().First(&restoredItem, "id = ?", itemId).Error; err != nil {
		return nil, err
	}
	
	return restoredItem, nil
}