package repositories

import (
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModuleTypeRepository interface {
	FindAll() ([]models.ModuleType, error)
	FindById(moduleTypeId string, isSoftDelete bool) (*models.ModuleType, error)
	Insert(moduleType *models.ModuleType) (*models.ModuleType, error)
	Update(moduleType *models.ModuleType) (*models.ModuleType, error)
	Delete(moduleTypeId string, isHardDelete bool) error
	Restore(moduleType *models.ModuleType, moduleTypeId string) (*models.ModuleType, error)
}

type ModuleTypeRepositoryImpl struct {
	DB *gorm.DB
}

func NewModuleTypeRepository(db *gorm.DB) *ModuleTypeRepositoryImpl {
	return &ModuleTypeRepositoryImpl{DB: db}
}

func (r *ModuleTypeRepositoryImpl) FindAll() ([]models.ModuleType, error) {
	var moduleTypes []models.ModuleType

	if err := r.DB.
	 Unscoped().
		Find(&moduleTypes).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}

	return moduleTypes, nil
}

func (r *ModuleTypeRepositoryImpl) FindById(moduleTypeId string, isSoftDelete bool) (*models.ModuleType, error) {
	var moduleType *models.ModuleType
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.
		First(&moduleType, "id = ?", moduleTypeId).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}
	
	return moduleType, nil
}

func (r *ModuleTypeRepositoryImpl) Insert(moduleType *models.ModuleType) (*models.ModuleType, error) {

	if moduleType.ID == uuid.Nil {
		return nil, fmt.Errorf("moduleType ID cannot be empty")
	}

	if err := r.DB.Create(&moduleType).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}

	return moduleType, nil
}

func (r *ModuleTypeRepositoryImpl) Update(moduleType *models.ModuleType) (*models.ModuleType, error) {

	if moduleType.ID == uuid.Nil {
		return nil, fmt.Errorf("moduleType ID cannot be empty")
	}

	if err := r.DB.Save(&moduleType).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}

	return moduleType, nil
}

func (r *ModuleTypeRepositoryImpl) Delete(moduleTypeId string, isHardDelete bool) error {
	var moduleType *models.ModuleType

	if err := r.DB.Unscoped().First(&moduleType, "id = ?", moduleTypeId).Error; err != nil {
		return HandleDatabaseError(err, "module_type")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&moduleType).Error; err != nil {
			return HandleDatabaseError(err, "module_type")
		}
	} else {
		if err := r.DB.Delete(&moduleType).Error; err != nil {
			return HandleDatabaseError(err, "module_type")
		}
	}
	return nil
}

func (r *ModuleTypeRepositoryImpl) Restore(moduleType *models.ModuleType, moduleTypeId string) (*models.ModuleType, error) {
	if err := r.DB.Unscoped().Model(moduleType).Where("id = ?", moduleTypeId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredModuleType *models.ModuleType
	if err := r.DB.Unscoped().First(&restoredModuleType, "id = ?", moduleTypeId).Error; err != nil {
		return nil, err
	}
	
	return restoredModuleType, nil
}