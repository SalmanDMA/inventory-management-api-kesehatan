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

type AreaRepository interface {
	FindAll(tx *gorm.DB) ([]models.Area, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Area, int64, error)
	FindById(tx *gorm.DB, areaId string, includeTrashed bool) (*models.Area, error)
	FindByName(tx *gorm.DB, areaName string) (*models.Area, error)
	FindByCode(tx *gorm.DB, areaCode string) (*models.Area, error)
	Insert(tx *gorm.DB, area *models.Area) (*models.Area, error)
	Update(tx *gorm.DB, area *models.Area) (*models.Area, error)
	Delete(tx *gorm.DB, areaId string, isHardDelete bool) error
	Restore(tx *gorm.DB, areaId string) (*models.Area, error)
}

// ==============================
// Implementation
// ==============================

type AreaRepositoryImpl struct {
	DB *gorm.DB
}

func NewAreaRepository(db *gorm.DB) *AreaRepositoryImpl {
	return &AreaRepositoryImpl{DB: db}
}

func (r *AreaRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *AreaRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Area, error) {
	var areas []models.Area
	if err := r.useDB(tx).
		Unscoped().
		Find(&areas).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return areas, nil
}

func (r *AreaRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Area, int64, error) {
	var (
		areas      []models.Area
		totalCount int64
	)

	q := r.useDB(tx).
		Unscoped().
		Model(&models.Area{})

	switch req.Status {
	case "active":
		q = q.Where("areas.deleted_at IS NULL")
	case "deleted":
		q = q.Where("areas.deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		q = q.Where("areas.deleted_at IS NULL")
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		q = q.Where(`
			LOWER(areas.name)  LIKE ? OR
			LOWER(areas.code)  LIKE ? OR
			LOWER(COALESCE(areas.color, '')) LIKE ?
		`, p, p, p)
	}

	if err := q.Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "area")
	}

	offset := (req.Page - 1) * req.Limit
	if err := q.
		Order("areas.created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&areas).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "area")
	}

	return areas, totalCount, nil
}

func (r *AreaRepositoryImpl) FindById(tx *gorm.DB, areaId string, includeTrashed bool) (*models.Area, error) {
	var area models.Area
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.First(&area, "id = ?", areaId).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return &area, nil
}

func (r *AreaRepositoryImpl) FindByName(tx *gorm.DB, areaName string) (*models.Area, error) {
	var area models.Area
	if err := r.useDB(tx).
		Where("name = ?", areaName).
		First(&area).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return &area, nil
}

func (r *AreaRepositoryImpl) FindByCode(tx *gorm.DB, areaCode string) (*models.Area, error) {
	var area models.Area
	if err := r.useDB(tx).
		Where("code = ?", areaCode).
		First(&area).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return &area, nil
}

// ---------- Mutations ----------

func (r *AreaRepositoryImpl) Insert(tx *gorm.DB, area *models.Area) (*models.Area, error) {
	if area.ID == uuid.Nil {
		return nil, fmt.Errorf("area ID cannot be empty")
	}
	if err := r.useDB(tx).Create(area).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return area, nil
}

func (r *AreaRepositoryImpl) Update(tx *gorm.DB, area *models.Area) (*models.Area, error) {
	if area.ID == uuid.Nil {
		return nil, fmt.Errorf("area ID cannot be empty")
	}
	if err := r.useDB(tx).Save(area).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return area, nil
}

func (r *AreaRepositoryImpl) Delete(tx *gorm.DB, areaId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var area models.Area
	if err := db.Unscoped().First(&area, "id = ?", areaId).Error; err != nil {
		return HandleDatabaseError(err, "area")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&area).Error; err != nil {
			return HandleDatabaseError(err, "area")
		}
	} else {
		if err := db.Delete(&area).Error; err != nil {
			return HandleDatabaseError(err, "area")
		}
	}
	return nil
}

func (r *AreaRepositoryImpl) Restore(tx *gorm.DB, areaId string) (*models.Area, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.Area{}).
		Where("id = ?", areaId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}

	var restored models.Area
	if err := db.First(&restored, "id = ?", areaId).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return &restored, nil
}
