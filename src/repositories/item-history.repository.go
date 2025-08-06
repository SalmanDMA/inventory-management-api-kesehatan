package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ItemHistoryRepository interface {
	FindAllByItem(itemID uuid.UUID, changeType string) ([]models.ItemHistory, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.ItemHistory, int64, error)
	FindById(itemHistoryId string, isSoftDelete bool) (*models.ItemHistory, error)
	FindLastByItem(itemID uuid.UUID, changeType string) (*models.ItemHistory, error)
	Insert(itemHistory *models.ItemHistory) (*models.ItemHistory, error)
	Delete(itemHistoryId string, isHardDelete bool) error
	Restore(itemHistory *models.ItemHistory, itemHistoryId string) (*models.ItemHistory, error)
}

type ItemHistoryRepositoryImpl struct{
	DB *gorm.DB
}

func NewItemHistoryRepository(db *gorm.DB) *ItemHistoryRepositoryImpl {
	return &ItemHistoryRepositoryImpl{DB: db}
}

func (r *ItemHistoryRepositoryImpl) FindAllByItem(itemID uuid.UUID, changeType string) ([]models.ItemHistory, error) {
	var itemHistories []models.ItemHistory
	db := r.DB

	var changeGroup []string

	switch strings.ToLower(changeType) {
	case "stock", "create_stock", "update_stock":
		changeGroup = []string{"create_stock", "update_stock"}
	case "price", "create_price", "update_price":
		changeGroup = []string{"create_price", "update_price"}
	default:
		return nil, fmt.Errorf("invalid changeType '%s', must be 'stock' or 'price'", changeType)
	}


	if err := db.
		Where("item_id = ? AND change_type IN ?", itemID, changeGroup).
		Order("created_at DESC").
		Find(&itemHistories).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}

	return itemHistories, nil
}

func (r *ItemHistoryRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.ItemHistory, int64, error) {
	var histories []models.ItemHistory
	var totalCount int64

	query := r.DB.Unscoped().
		Preload("Item").
		Preload("CreatedByUser").
		Preload("UpdatedByUser")

	switch req.Status {
	case "active":
		query = query.Where("deleted_at IS NULL")
	case "deleted":
		query = query.Where("deleted_at IS NOT NULL")
	case "all":
	default:
		query = query.Where("deleted_at IS NULL")
	}

	if req.ItemID != "" {
		if itemUUID, err := uuid.Parse(req.ItemID); err == nil {
			query = query.Where("item_id = ?", itemUUID)
		}
	}

	if req.ChangeType != "" {
		changeType := strings.ToLower(req.ChangeType)
		fmt.Println("changeType", changeType)
		switch changeType {
		case "stock", "create_stock", "update_stock":
			query = query.Where("change_type IN ?", []string{"create_stock", "update_stock"})
		case "price", "create_price", "update_price":
			query = query.Where("change_type IN ?", []string{"create_price", "update_price"})
		default:
			query = query.Where("change_type = ?", changeType)
		}
	}

	if req.Search != "" {
		search := "%" + strings.ToLower(req.Search) + "%"
		query = query.Joins("LEFT JOIN items ON items.id = item_histories.item_id").
			Where(`
				LOWER(items.name) LIKE ? OR
				LOWER(items.code) LIKE ? OR
				LOWER(items.description) LIKE ? OR
				CAST(item_histories.description AS TEXT) LIKE ? OR
				CAST(item_histories.old_price AS TEXT) LIKE ? OR
				CAST(item_histories.new_price AS TEXT) LIKE ? OR
				CAST(item_histories.current_price AS TEXT) LIKE ? OR
				CAST(item_histories.old_stock AS TEXT) LIKE ? OR
				CAST(item_histories.new_stock AS TEXT) LIKE ? OR
				CAST(item_histories.current_stock AS TEXT) LIKE ?
			`,
				search, search, search, 
				search, search, search, 
				search, search, search,
			)
	}

	if err := query.Model(&models.ItemHistory{}).Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "item_history")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&histories).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "item_history")
	}

	return histories, totalCount, nil
}

func (r *ItemHistoryRepositoryImpl) FindById(itemHistoryId string, isSoftDelete bool) (*models.ItemHistory, error) {
	var itemHistory *models.ItemHistory
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Item").
		Preload("CreatedByUser").
		Preload("UpdatedByUser").
	 First(&itemHistory, "id = ?", itemHistoryId).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}
	
	return itemHistory, nil
}

func (r *ItemHistoryRepositoryImpl) FindLastByItem(itemID uuid.UUID, changeType string) (*models.ItemHistory, error) {
	var itemHistory models.ItemHistory
	db := r.DB

	var changeGroup []string
	changeType = strings.ToLower(changeType)

	switch changeType {
	case "stock", "create_stock", "update_stock":
		changeGroup = []string{"create_stock", "update_stock"}
	case "price", "create_price", "update_price":
		changeGroup = []string{"create_price", "update_price"}
	default:
		return nil, fmt.Errorf("invalid changeType '%s', must be 'stock' or 'price'", changeType)
	}

	if err := db.
		Preload("Item").
		Preload("CreatedByUser").
		Preload("UpdatedByUser").
		Where("item_id = ? AND change_type IN ?", itemID, changeGroup).
		Order("created_at DESC").
		First(&itemHistory).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}

	return &itemHistory, nil
}

func (r *ItemHistoryRepositoryImpl) Insert(itemHistory *models.ItemHistory) (*models.ItemHistory, error) {
	if itemHistory.ID == uuid.Nil {
		return nil, fmt.Errorf("itemHistory ID cannot be empty")
	}

	if err := r.DB.Create(&itemHistory).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}
	return itemHistory, nil
}

func (r *ItemHistoryRepositoryImpl) Delete(itemHistoryId string, isHardDelete bool) error {
	var itemHistory *models.ItemHistory

	if err := r.DB.Unscoped().First(&itemHistory, "id = ?", itemHistoryId).Error; err != nil {
		return HandleDatabaseError(err, "item_history")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&itemHistory).Error; err != nil {
			return HandleDatabaseError(err, "item_history")
		}
	} else {
		if err := r.DB.Delete(&itemHistory).Error; err != nil {
			return HandleDatabaseError(err, "item_history")
		}
	}
	return nil
}

func (r *ItemHistoryRepositoryImpl) Restore(itemHistory *models.ItemHistory, itemHistoryId string) (*models.ItemHistory, error) {
	if err := r.DB.Unscoped().Model(itemHistory).Where("id = ?", itemHistoryId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredItemHistory *models.ItemHistory
	if err := r.DB.Unscoped().First(&restoredItemHistory, "id = ?", itemHistoryId).Error; err != nil {
		return nil, err
	}
	
	return restoredItemHistory, nil
}
