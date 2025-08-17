package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AreaRepository interface {
	FindAll() ([]models.Area, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.Area, int64, error)
	FindById(areaId string, isSoftDelete bool) (*models.Area, error)
	FindByName(areaName string) (*models.Area, error)
	FindByCode(areaCode string) (*models.Area, error)
	Insert(area *models.Area) (*models.Area, error)
	Update(area *models.Area) (*models.Area, error)
	Delete(areaId string, isHardDelete bool) error
	Restore(area *models.Area, areaId string) (*models.Area, error)
}

type AreaRepositoryImpl struct{
	DB *gorm.DB
}

func NewAreaRepository(db *gorm.DB) *AreaRepositoryImpl {
	return &AreaRepositoryImpl{DB: db}
}

func (r *AreaRepositoryImpl) FindAll() ([]models.Area, error) {
	var areas []models.Area
	if err := r.DB.
	 Unscoped().
	 Find(&areas).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return areas, nil
}

func (r *AreaRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.Area, int64, error) {
	var (
		areas      []models.Area
		totalCount int64
	)

	query := r.DB.
		Unscoped().
		Model(&models.Area{}) 

	switch req.Status {
	case "active":
		query = query.Where("areas.deleted_at IS NULL")
	case "deleted":
		query = query.Where("areas.deleted_at IS NOT NULL")
	case "all":
	default:
		query = query.Where("areas.deleted_at IS NULL")
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		query = query.Where(`
			LOWER(areas.name)  LIKE ? OR
			LOWER(areas.code)  LIKE ? OR
			LOWER(areas.color) LIKE ?
		`, p, p, p)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "area")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.
		Order("areas.created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&areas).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "area")
	}

	return areas, totalCount, nil
}

func (r *AreaRepositoryImpl) FindById(areaId string, isSoftDelete bool) (*models.Area, error) {
	var area *models.Area
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.First(&area, "id = ?", areaId).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	
	return area, nil
}

func (r *AreaRepositoryImpl) FindByName(areaName string) (*models.Area, error) {
	var area *models.Area
	if err := r.DB.Where("name = ?", areaName).First(&area).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return area, nil
}

func (r *AreaRepositoryImpl) FindByCode(areaCode string) (*models.Area, error) {
	var area *models.Area
	if err := r.DB.Where("code = ?", areaCode).First(&area).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return area, nil
}

func (r *AreaRepositoryImpl) Insert(area *models.Area) (*models.Area, error) {

	if area.ID == uuid.Nil {
		return nil, fmt.Errorf("area ID cannot be empty")
	}

	if err := r.DB.Create(&area).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return area, nil
}

func (r *AreaRepositoryImpl) Update(area *models.Area) (*models.Area, error) {

	if area.ID == uuid.Nil {
		return nil, fmt.Errorf("area ID cannot be empty")
	}

	if err := r.DB.Save(&area).Error; err != nil {
		return nil, HandleDatabaseError(err, "area")
	}
	return area, nil
}

func (r *AreaRepositoryImpl) Delete(areaId string, isHardDelete bool) error {
	var area *models.Area

	if err := r.DB.Unscoped().First(&area, "id = ?", areaId).Error; err != nil {
		return HandleDatabaseError(err, "area")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&area).Error; err != nil {
			return HandleDatabaseError(err, "area")
		}
	} else {
		if err := r.DB.Delete(&area).Error; err != nil {
			return HandleDatabaseError(err, "area")
		}
	}
	return nil
}

func (r *AreaRepositoryImpl) Restore(area *models.Area, areaId string) (*models.Area, error) {
	if err := r.DB.Unscoped().Model(area).Where("id = ?", areaId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredArea *models.Area
	if err := r.DB.Unscoped().First(&restoredArea, "id = ?", areaId).Error; err != nil {
		return nil, err
	}
	
	return restoredArea, nil
}