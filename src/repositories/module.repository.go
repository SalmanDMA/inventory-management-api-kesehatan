package repositories

import (
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"gorm.io/gorm"
)

type ModuleRepository interface {
	FindAll() ([]models.Module, error)
	FindById(moduleId int, isSoftDelete bool) (*models.Module, error)
	// FindRootModule() (*models.Module, error)
	Insert(module *models.Module) (*models.Module, error)
	Update(module *models.Module) (*models.Module, error)
	Delete(moduleId int, isHardDelete bool) error
	Restore(module *models.Module, moduleId int) (*models.Module, error)
}

type ModuleRepositoryImpl struct {
	DB *gorm.DB
}

func NewModuleRepository(db *gorm.DB) *ModuleRepositoryImpl {
	return &ModuleRepositoryImpl{DB: db}
}

func (r *ModuleRepositoryImpl) FindAll() ([]models.Module, error) {
	var modules []models.Module
	if err := r.DB.Unscoped().
			 Preload("Parent").
				Preload("Children").
    Preload("ModuleType").
    Find(&modules).Error; err != nil {
    return nil, HandleDatabaseError(err, "module")
	}
	return modules, nil
}

func (r *ModuleRepositoryImpl) FindById(moduleId int, isSoftDelete bool) (*models.Module, error) {
	var module *models.Module
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.
				Preload("Parent").
				Preload("Children").
    Preload("ModuleType").
    Where("id = ?", moduleId).
    First(&module).Error; err != nil {
    return nil, HandleDatabaseError(err, "module")
}

	return module, nil
}

func (r *ModuleRepositoryImpl) Insert(module *models.Module) (*models.Module, error) {

	if err := r.DB.Create(&module).Error; err != nil {
		return nil, HandleDatabaseError(err, "module")
	}

	return module, nil
}

// func (r *ModuleRepositoryImpl) FindRootModule() (*models.Module, error) {
// 	var module *models.Module

// 	if err := r.DB.Unscoped().
// 		Preload("Icon").
// 		Preload("ModuleType").
// 		Where("title = ?", "Root").
// 		First(&module).Error; err != nil {
// 		return nil, HandleDatabaseError(err, "module")
// 	}

// 	return module, nil
// }

func (r *ModuleRepositoryImpl) Update(module *models.Module) (*models.Module, error) {

	if err := r.DB.Save(&module).Error; err != nil {
		return nil, HandleDatabaseError(err, "module")
	}

	return module, nil
}

func (r *ModuleRepositoryImpl) Delete(moduleId int, isHardDelete bool) error {
	var module *models.Module

	if err := r.DB.Unscoped().First(&module, "id = ?", moduleId).Error; err != nil { 
		return HandleDatabaseError(err, "module")
	}

	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&module).Error; err != nil {
			return HandleDatabaseError(err, "module")
		}
	} else {
		if err := r.DB.Delete(&module).Error; err != nil {
			return HandleDatabaseError(err, "module")
		}
	}

	return nil
}

func (r *ModuleRepositoryImpl) Restore(module *models.Module, moduleId int) (*models.Module, error) {

	if err := r.DB.Unscoped().Model(module).Where("id = ?", moduleId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredModule *models.Module
	if err := r.DB.Unscoped().First(&restoredModule, "id = ?", moduleId).Error; err != nil {
		return nil, err
	}

	return restoredModule, nil
}
