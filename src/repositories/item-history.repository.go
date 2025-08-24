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

type ItemHistoryRepository interface {
	FindAllByItem(tx *gorm.DB, itemID uuid.UUID, changeType string) ([]models.ItemHistory, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.ItemHistory, int64, error)
	FindById(tx *gorm.DB, itemHistoryId string, includeTrashed bool) (*models.ItemHistory, error)
	FindLastByItem(tx *gorm.DB, itemID uuid.UUID, changeType string) (*models.ItemHistory, error)
	Insert(tx *gorm.DB, itemHistory *models.ItemHistory) (*models.ItemHistory, error)
	Delete(tx *gorm.DB, itemHistoryId string, isHardDelete bool) error
	Restore(tx *gorm.DB, itemHistoryId string) (*models.ItemHistory, error)
}

// ==============================
// Implementation
// ==============================

type ItemHistoryRepositoryImpl struct {
	DB *gorm.DB
}

func NewItemHistoryRepository(db *gorm.DB) *ItemHistoryRepositoryImpl {
	return &ItemHistoryRepositoryImpl{DB: db}
}

func (r *ItemHistoryRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *ItemHistoryRepositoryImpl) FindAllByItem(tx *gorm.DB, itemID uuid.UUID, changeType string) ([]models.ItemHistory, error) {
	var histories []models.ItemHistory
	db := r.useDB(tx)

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
		Find(&histories).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}
	return histories, nil
}

func (r *ItemHistoryRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.ItemHistory, int64, error) {
	var (
		histories  []models.ItemHistory
		totalCount int64
	)

	query := r.useDB(tx).Unscoped().
		Preload("Item").
		Preload("CreatedByUser").
		Preload("UpdatedByUser")

	switch req.Status {
	case "active":
		query = query.Where("item_histories.deleted_at IS NULL")
	case "deleted":
		query = query.Where("item_histories.deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		query = query.Where("item_histories.deleted_at IS NULL")
	}

	if req.ItemID != "" {
		if itemUUID, err := uuid.Parse(req.ItemID); err == nil {
			query = query.Where("item_id = ?", itemUUID)
		}
	}

	if req.ChangeType != "" {
		switch strings.ToLower(req.ChangeType) {
		case "stock", "create_stock", "update_stock":
			query = query.Where("change_type IN ?", []string{"create_stock", "update_stock"})
		case "price", "create_price", "update_price":
			query = query.Where("change_type IN ?", []string{"create_price", "update_price"})
		default:
			query = query.Where("change_type = ?", strings.ToLower(req.ChangeType))
		}
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		query = query.
			Joins("LEFT JOIN items ON items.id = item_histories.item_id").
			Where(`
				LOWER(items.name) LIKE ? OR
				LOWER(items.code) LIKE ? OR
				LOWER(COALESCE(items.description, '')) LIKE ? OR
				CAST(item_histories.description AS TEXT) LIKE ? OR
				CAST(item_histories.old_price AS TEXT) LIKE ? OR
				CAST(item_histories.new_price AS TEXT) LIKE ? OR
				CAST(item_histories.current_price AS TEXT) LIKE ? OR
				CAST(item_histories.old_stock AS TEXT) LIKE ? OR
				CAST(item_histories.new_stock AS TEXT) LIKE ? OR
				CAST(item_histories.current_stock AS TEXT) LIKE ?
			`, p, p, p, p, p, p, p, p, p, p)
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

func (r *ItemHistoryRepositoryImpl) FindById(tx *gorm.DB, itemHistoryId string, includeTrashed bool) (*models.ItemHistory, error) {
	var ih models.ItemHistory
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Item").
		Preload("CreatedByUser").
		Preload("UpdatedByUser").
		First(&ih, "id = ?", itemHistoryId).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}
	return &ih, nil
}

func (r *ItemHistoryRepositoryImpl) FindLastByItem(tx *gorm.DB, itemID uuid.UUID, changeType string) (*models.ItemHistory, error) {
	var ih models.ItemHistory
	db := r.useDB(tx)

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
		Preload("Item").
		Preload("CreatedByUser").
		Preload("UpdatedByUser").
		Where("item_id = ? AND change_type IN ?", itemID, changeGroup).
		Order("created_at DESC").
		First(&ih).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}
	return &ih, nil
}

// ---------- Mutations ----------

func (r *ItemHistoryRepositoryImpl) Insert(tx *gorm.DB, itemHistory *models.ItemHistory) (*models.ItemHistory, error) {
	if itemHistory.ID == uuid.Nil {
		return nil, fmt.Errorf("itemHistory ID cannot be empty")
	}
	if err := r.useDB(tx).Create(itemHistory).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}
	return itemHistory, nil
}

func (r *ItemHistoryRepositoryImpl) Delete(tx *gorm.DB, itemHistoryId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var ih models.ItemHistory
	if err := db.Unscoped().First(&ih, "id = ?", itemHistoryId).Error; err != nil {
		return HandleDatabaseError(err, "item_history")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&ih).Error; err != nil {
			return HandleDatabaseError(err, "item_history")
		}
	} else {
		if err := db.Delete(&ih).Error; err != nil {
			return HandleDatabaseError(err, "item_history")
		}
	}
	return nil
}

func (r *ItemHistoryRepositoryImpl) Restore(tx *gorm.DB, itemHistoryId string) (*models.ItemHistory, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.ItemHistory{}).
		Where("id = ?", itemHistoryId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}

	var restored models.ItemHistory
	if err := db.
		Preload("Item").
		Preload("CreatedByUser").
		Preload("UpdatedByUser").
		First(&restored, "id = ?", itemHistoryId).Error; err != nil {
		return nil, HandleDatabaseError(err, "item_history")
	}
	return &restored, nil
}
