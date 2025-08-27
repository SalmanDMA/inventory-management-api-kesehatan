package repositories

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type ItemRepository interface {
	FindAll(tx *gorm.DB) ([]models.Item, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Item, int64, error)
	FindById(tx *gorm.DB, itemId string, includeTrashed bool) (*models.Item, error)
	FindByName(tx *gorm.DB, itemName string) (*models.Item, error)
	FindConsignmentDueBetween(tx *gorm.DB, start, end time.Time) ([]models.Item, error)
	CountAllThisMonth(tx *gorm.DB) (int64, error)
	CountAllLastMonth(tx *gorm.DB) (int64, error)
	CountLowStockNow(tx *gorm.DB) (int64, error)
	CountLowStockLastMonth(tx *gorm.DB) (int64, error)
	Insert(tx *gorm.DB, item *models.Item) (*models.Item, error)
	Update(tx *gorm.DB, item *models.Item) (*models.Item, error)
	Delete(tx *gorm.DB, itemId string, isHardDelete bool) error
	Restore(tx *gorm.DB, itemId string) (*models.Item, error)
}

// ==============================
// Implementation
// ==============================

type ItemRepositoryImpl struct {
	DB *gorm.DB
}

func NewItemRepository(db *gorm.DB) *ItemRepositoryImpl {
	return &ItemRepositoryImpl{DB: db}
}

func (r *ItemRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *ItemRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Item, error) {
	var items []models.Item
	db := r.useDB(tx).Unscoped().
		Preload("Image").
		Preload("UoM").
		Preload("Category").
		Preload("ItemHistories")

	if err := db.Find(&items).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return items, nil
}

func (r *ItemRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Item, int64, error) {
	var (
		items      []models.Item
		totalCount int64
	)

	query := r.useDB(tx).Unscoped().
		Preload("Image").
		Preload("UoM").
		Preload("Category").
		Preload("ItemHistories")

	switch req.Status {
	case "active":
		query = query.Where("items.deleted_at IS NULL")
	case "deleted":
		query = query.Where("items.deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		query = query.Where("items.deleted_at IS NULL")
	}

	if req.CategoryID != "" {
		if catUUID, err := uuid.Parse(req.CategoryID); err == nil {
			query = query.Where("category_id = ?", catUUID)
		}
	}
	if req.UoMID != "" {
		if uomUUID, err := uuid.Parse(req.UoMID); err == nil {
			query = query.Where("uom_id = ?", uomUUID)
		}
	}
	if b := strings.TrimSpace(req.Batch); b != "" {
			n, err := strconv.Atoi(b)
			if err != nil {
					return nil, 0, fmt.Errorf("invalid batch number: %s", b)
			}
			query = query.Where("items.batch = ?", n)
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		query = query.
			Joins("LEFT JOIN categories ON categories.id = items.category_id").
			Where(`
				LOWER(items.name) LIKE ? OR
				LOWER(items.code) LIKE ? OR
				LOWER(COALESCE(items.description, '')) LIKE ? OR
				CAST(items.price AS TEXT) LIKE ? OR
				CAST(items.stock AS TEXT) LIKE ? OR
				LOWER(categories.name) LIKE ?
			`, p, p, p, p, p, p)
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

func (r *ItemRepositoryImpl) FindById(tx *gorm.DB, itemId string, includeTrashed bool) (*models.Item, error) {
	var item models.Item
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Image").
		Preload("UoM").
		Preload("Category").
		Preload("ItemHistories").
		First(&item, "id = ?", itemId).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return &item, nil
}

func (r *ItemRepositoryImpl) FindByName(tx *gorm.DB, itemName string) (*models.Item, error) {
	var item models.Item
	if err := r.useDB(tx).
		Where("name = ?", itemName).
		First(&item).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return &item, nil
}

func (r *ItemRepositoryImpl) FindConsignmentDueBetween(tx *gorm.DB, start, end time.Time) ([]models.Item, error) {
	var items []models.Item
	db := r.useDB(tx).
		Where("is_consignment = ?", true).
		Where("due_date IS NOT NULL").
		Where("due_date >= ? AND due_date <= ?", start, end)

	if err := db.Find(&items).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return items, nil
}

// ---------- Counts ----------

func (r *ItemRepositoryImpl) CountAllThisMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.Item{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())").
		Count(&count).Error
	return count, err
}

func (r *ItemRepositoryImpl) CountAllLastMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.Item{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Count(&count).Error
	return count, err
}

func (r *ItemRepositoryImpl) CountLowStockNow(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.Item{}).
		Where("stock <= low_stock").
		Count(&count).Error
	return count, err
}

func (r *ItemRepositoryImpl) CountLowStockLastMonth(tx *gorm.DB) (int64, error) {
	var count int64
	err := r.useDB(tx).Model(&models.Item{}).
		Where("DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW() - INTERVAL '1 month')").
		Where("stock <= low_stock").
		Count(&count).Error
	return count, err
}

// ---------- Mutations ----------

func (r *ItemRepositoryImpl) Insert(tx *gorm.DB, item *models.Item) (*models.Item, error) {
	if item.ID == uuid.Nil {
		return nil, fmt.Errorf("item ID cannot be empty")
	}
	if err := r.useDB(tx).Create(item).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return item, nil
}

func (r *ItemRepositoryImpl) Update(tx *gorm.DB, item *models.Item) (*models.Item, error) {
	if item.ID == uuid.Nil {
		return nil, fmt.Errorf("item ID cannot be empty")
	}
	if err := r.useDB(tx).Save(item).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return item, nil
}

func (r *ItemRepositoryImpl) Delete(tx *gorm.DB, itemId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var item models.Item
	if err := db.Unscoped().First(&item, "id = ?", itemId).Error; err != nil {
		return HandleDatabaseError(err, "item")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&item).Error; err != nil {
			return HandleDatabaseError(err, "item")
		}
	} else {
		if err := db.Delete(&item).Error; err != nil {
			return HandleDatabaseError(err, "item")
		}
	}
	return nil
}

func (r *ItemRepositoryImpl) Restore(tx *gorm.DB, itemId string) (*models.Item, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.Item{}).
		Where("id = ?", itemId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}

	var restored models.Item
	if err := db.
		Preload("Image").
		Preload("UoM").
		Preload("Category").
		Preload("ItemHistories").
		First(&restored, "id = ?", itemId).Error; err != nil {
		return nil, HandleDatabaseError(err, "item")
	}
	return &restored, nil
}
