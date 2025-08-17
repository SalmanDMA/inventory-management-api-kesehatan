package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FacilityTypeRepository interface {
	FindAll() ([]models.FacilityType, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.FacilityType, int64, error)
	FindById(facilityTypeId string, isSoftDelete bool) (*models.FacilityType, error)
	FindByName(facilityTypeName string) (*models.FacilityType, error)
	Insert(facilityType *models.FacilityType) (*models.FacilityType, error)
	Update(facilityType *models.FacilityType) (*models.FacilityType, error)
	Delete(facilityTypeId string, isHardDelete bool) error
	Restore(facilityType *models.FacilityType, facilityTypeId string) (*models.FacilityType, error)
}

type FacilityTypeRepositoryImpl struct{
	DB *gorm.DB
}

func NewFacilityTypeRepository(db *gorm.DB) *FacilityTypeRepositoryImpl {
	return &FacilityTypeRepositoryImpl{DB: db}
}

func (r *FacilityTypeRepositoryImpl) FindAll() ([]models.FacilityType, error) {
	var facilityTypes []models.FacilityType
	if err := r.DB.
	 Unscoped().
	 Find(&facilityTypes).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility_type")
	}
	return facilityTypes, nil
}

func (r *FacilityTypeRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.FacilityType, int64, error) {
	var facilityTypes []models.FacilityType
	var totalCount int64

	query := r.DB.
		Unscoped().
		Model(&models.FacilityType{}) 

	switch req.Status {
	case "active":
		query = query.Where("facility_types.deleted_at IS NULL")
	case "deleted":
		query = query.Where("facility_types.deleted_at IS NOT NULL")
	case "all":
	default:
		query = query.Where("facility_types.deleted_at IS NULL")
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		query = query.Where(`
			LOWER(facility_types.name) LIKE ? OR
			LOWER(facility_types.color) LIKE ? OR
			LOWER(COALESCE(facility_types.description, '')) LIKE ?
		`, p, p, p)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "facility_type")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.
		Order("facility_types.created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Find(&facilityTypes).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "facility_type")
	}

	return facilityTypes, totalCount, nil
}

func (r *FacilityTypeRepositoryImpl) FindById(facilityTypeId string, isSoftDelete bool) (*models.FacilityType, error) {
	var facilityType *models.FacilityType
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.First(&facilityType, "id = ?", facilityTypeId).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility_type")
	}
	
	return facilityType, nil
}

func (r *FacilityTypeRepositoryImpl) FindByName(facilityTypeName string) (*models.FacilityType, error) {
	var facilityType *models.FacilityType
	if err := r.DB.Where("name = ?", facilityTypeName).First(&facilityType).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility_type")
	}
	return facilityType, nil
}

func (r *FacilityTypeRepositoryImpl) Insert(facilityType *models.FacilityType) (*models.FacilityType, error) {

	if facilityType.ID == uuid.Nil {
		return nil, fmt.Errorf("facilityType ID cannot be empty")
	}

	if err := r.DB.Create(&facilityType).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility_type")
	}
	return facilityType, nil
}

func (r *FacilityTypeRepositoryImpl) Update(facilityType *models.FacilityType) (*models.FacilityType, error) {

	if facilityType.ID == uuid.Nil {
		return nil, fmt.Errorf("facilityType ID cannot be empty")
	}

	if err := r.DB.Save(&facilityType).Error; err != nil {
		return nil, HandleDatabaseError(err, "facility_type")
	}
	return facilityType, nil
}

func (r *FacilityTypeRepositoryImpl) Delete(facilityTypeId string, isHardDelete bool) error {
	var facilityType *models.FacilityType

	if err := r.DB.Unscoped().First(&facilityType, "id = ?", facilityTypeId).Error; err != nil {
		return HandleDatabaseError(err, "facility_type")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&facilityType).Error; err != nil {
			return HandleDatabaseError(err, "facility_type")
		}
	} else {
		if err := r.DB.Delete(&facilityType).Error; err != nil {
			return HandleDatabaseError(err, "facility_type")
		}
	}
	return nil
}

func (r *FacilityTypeRepositoryImpl) Restore(facilityType *models.FacilityType, facilityTypeId string) (*models.FacilityType, error) {
	if err := r.DB.Unscoped().Model(facilityType).Where("id = ?", facilityTypeId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredFacilityType *models.FacilityType
	if err := r.DB.Unscoped().First(&restoredFacilityType, "id = ?", facilityTypeId).Error; err != nil {
		return nil, err
	}
	
	return restoredFacilityType, nil
}