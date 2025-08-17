package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FacilityRepository interface {
	FindAll() ([]models.Facility, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.Facility, int64, error)
	FindById(facilityId string, isSoftDelete bool) (*models.Facility, error)
	FindByName(facilityName string) (*models.Facility, error)
	FindByCode(facilityCode string) (*models.Facility, error)
	Insert(facility *models.Facility) (*models.Facility, error)
	Update(facility *models.Facility) (*models.Facility, error)
	Delete(facilityId string, isHardDelete bool) error
	Restore(facility *models.Facility, facilityId string) (*models.Facility, error)
}

type FacilityRepositoryImpl struct{
	DB *gorm.DB
}

func NewFacilityRepository(db *gorm.DB) *FacilityRepositoryImpl {
	return &FacilityRepositoryImpl{DB: db}
}

func (r *FacilityRepositoryImpl) FindAll() ([]models.Facility, error) {
	var facilities []models.Facility
	if err := r.DB.
	 Unscoped().Preload("Area").Preload("FacilityType").
	 Find(&facilities).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility")
	}
	return facilities, nil
}

func (r *FacilityRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.Facility, int64, error) {
	var (
		facilities  []models.Facility
		totalCount  int64
	)

	// Base query
	query := r.DB.Unscoped().
		Model(&models.Facility{}).
		Preload("Area").
		Preload("FacilityType").
		Select("facilities.*") 

	switch req.Status {
	case "active":
		query = query.Where("facilities.deleted_at IS NULL")
	case "deleted":
		query = query.Where("facilities.deleted_at IS NOT NULL")
	case "all":
	default:
		query = query.Where("facilities.deleted_at IS NULL")
	}

	if req.AreaID != "" {
		if areaUUID, err := uuid.Parse(req.AreaID); err == nil {
			query = query.Where("facilities.area_id = ?", areaUUID)
		}
	}
	if req.FacilityTypeID != "" {
		if ftUUID, err := uuid.Parse(req.FacilityTypeID); err == nil {
			query = query.Where("facilities.facility_type_id = ?", ftUUID)
		}
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		query = query.
			Joins("LEFT JOIN areas a ON a.id = facilities.area_id").
			Joins("LEFT JOIN facility_types ft ON ft.id = facilities.facility_type_id").
			Where(`
				LOWER(facilities.name) LIKE ? OR
				LOWER(facilities.code) LIKE ? OR
				LOWER(a.name)          LIKE ? OR
				LOWER(ft.name)         LIKE ?
			`, p, p, p, p)
	}

	countQ := query.Session(&gorm.Session{})
	if err := countQ.Distinct("facilities.id").Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "facility")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.
		Offset(offset).
		Limit(req.Limit).
		Order("facilities.created_at DESC").
		Find(&facilities).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "facility")
	}

	return facilities, totalCount, nil
}

func (r *FacilityRepositoryImpl) FindById(facilityId string, isSoftDelete bool) (*models.Facility, error) {
	var facility *models.Facility
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped().Preload("Area").Preload("FacilityType")
	}

	if err := db.First(&facility, "id = ?", facilityId).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility")
	}
	
	return facility, nil
}

func (r *FacilityRepositoryImpl) FindByName(facilityName string) (*models.Facility, error) {
	var facility *models.Facility
	if err := r.DB.Where("name = ?", facilityName).First(&facility).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility")
	}
	return facility, nil
}

func (r *FacilityRepositoryImpl) FindByCode(facilityCode string) (*models.Facility, error) {
	var facility *models.Facility
	if err := r.DB.Where("code = ?", facilityCode).First(&facility).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility")
	}
	return facility, nil
}

func (r *FacilityRepositoryImpl) Insert(facility *models.Facility) (*models.Facility, error) {

	if facility.ID == uuid.Nil {
		return nil, fmt.Errorf("facility ID cannot be empty")
	}

	if err := r.DB.Create(&facility).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility")
	}
	return facility, nil
}

func (r *FacilityRepositoryImpl) Update(facility *models.Facility) (*models.Facility, error) {

	if facility.ID == uuid.Nil {
		return nil, fmt.Errorf("facility ID cannot be empty")
	}

	if err := r.DB.Save(&facility).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility")
	}
	return facility, nil
}

func (r *FacilityRepositoryImpl) Delete(facilityId string, isHardDelete bool) error {
	var facility *models.Facility

	if err := r.DB.Unscoped().First(&facility, "id = ?", facilityId).Error; err != nil {
		return HandleDatabaseError(err, "facility")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&facility).Error; err != nil {
			return HandleDatabaseError(err, "facility")
		}
	} else {
		if err := r.DB.Delete(&facility).Error; err != nil {
			return HandleDatabaseError(err, "facility")
		}
	}
	return nil
}

func (r *FacilityRepositoryImpl) Restore(facility *models.Facility, facilityId string) (*models.Facility, error) {
	if err := r.DB.Unscoped().Model(facility).Where("id = ?", facilityId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredFacility *models.Facility
	if err := r.DB.Unscoped().First(&restoredFacility, "id = ?", facilityId).Error; err != nil {
		return nil, err
	}
	
	return restoredFacility, nil
}