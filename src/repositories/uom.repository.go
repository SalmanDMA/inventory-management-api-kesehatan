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

type UoMRepository interface {
	FindAll(tx *gorm.DB) ([]models.UoM, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.UoM, int64, error)
	FindById(tx *gorm.DB, uomId string, includeTrashed bool) (*models.UoM, error)
	FindByName(tx *gorm.DB, uomName string) (*models.UoM, error)
	Insert(tx *gorm.DB, uom *models.UoM) (*models.UoM, error)
	Update(tx *gorm.DB, uom *models.UoM) (*models.UoM, error)
	Delete(tx *gorm.DB, uomId string, isHardDelete bool) error
	Restore(tx *gorm.DB, uomId string) (*models.UoM, error)
}

// ==============================
// Implementation
// ==============================

type UoMRepositoryImpl struct {
	DB *gorm.DB
}

func NewUoMRepository(db *gorm.DB) *UoMRepositoryImpl {
	return &UoMRepositoryImpl{DB: db}
}

func (r *UoMRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *UoMRepositoryImpl) FindAll(tx *gorm.DB) ([]models.UoM, error) {
	var uoms []models.UoM
	if err := r.useDB(tx).Unscoped().Find(&uoms).Error; err != nil {
		return nil, HandleDatabaseError(err, "uom")
	}
	return uoms, nil
}

func (r *UoMRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.UoM, int64, error) {
	var (
		uoms       []models.UoM
		totalCount int64
	)

	query := r.useDB(tx).Unscoped().Model(&models.UoM{})

	switch req.Status {
	case "active":
		query = query.Where("uoms.deleted_at IS NULL")
	case "deleted":
		query = query.Where("uoms.deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		query = query.Where("uoms.deleted_at IS NULL")
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		query = query.Where(`
			LOWER(uoms.name) LIKE ? OR
			LOWER(COALESCE(uoms.color, '')) LIKE ? OR
			LOWER(COALESCE(uoms.description, '')) LIKE ?
		`, p, p, p)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "uom")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.
		Order("uoms.created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&uoms).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "uom")
	}

	return uoms, totalCount, nil
}

func (r *UoMRepositoryImpl) FindById(tx *gorm.DB, uomId string, includeTrashed bool) (*models.UoM, error) {
	var uom models.UoM
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.First(&uom, "id = ?", uomId).Error; err != nil {
		return nil, HandleDatabaseError(err, "uom")
	}
	return &uom, nil
}

func (r *UoMRepositoryImpl) FindByName(tx *gorm.DB, uomName string) (*models.UoM, error) {
	var uom models.UoM
	if err := r.useDB(tx).Where("name = ?", uomName).First(&uom).Error; err != nil {
		return nil, HandleDatabaseError(err, "uom")
	}
	return &uom, nil
}

// ---------- Mutations ----------

func (r *UoMRepositoryImpl) Insert(tx *gorm.DB, uom *models.UoM) (*models.UoM, error) {
	if uom.ID == uuid.Nil {
		return nil, fmt.Errorf("uom ID cannot be empty")
	}
	if err := r.useDB(tx).Create(uom).Error; err != nil {
		return nil, HandleDatabaseError(err, "uom")
	}
	return uom, nil
}

func (r *UoMRepositoryImpl) Update(tx *gorm.DB, uom *models.UoM) (*models.UoM, error) {
	if uom.ID == uuid.Nil {
		return nil, fmt.Errorf("uom ID cannot be empty")
	}
	if err := r.useDB(tx).Save(uom).Error; err != nil {
		return nil, HandleDatabaseError(err, "uom")
	}
	return uom, nil
}

func (r *UoMRepositoryImpl) Delete(tx *gorm.DB, uomId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var uom models.UoM
	if err := db.Unscoped().First(&uom, "id = ?", uomId).Error; err != nil {
		return HandleDatabaseError(err, "uom")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&uom).Error; err != nil {
			return HandleDatabaseError(err, "uom")
		}
	} else {
		if err := db.Delete(&uom).Error; err != nil {
			return HandleDatabaseError(err, "uom")
		}
	}
	return nil
}

func (r *UoMRepositoryImpl) Restore(tx *gorm.DB, uomId string) (*models.UoM, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.UoM{}).
		Where("id = ?", uomId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "uom")
	}

	var restored models.UoM
	if err := db.First(&restored, "id = ?", uomId).Error; err != nil {
		return nil, HandleDatabaseError(err, "uom")
	}
	return &restored, nil
}
