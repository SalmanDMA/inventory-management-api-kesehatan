package repositories

import (
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type ModuleRepository interface {
	FindAll(tx *gorm.DB) ([]models.Module, error)
	FindById(tx *gorm.DB, moduleId int, includeTrashed bool) (*models.Module, error)
	Insert(tx *gorm.DB, module *models.Module) (*models.Module, error)
	Update(tx *gorm.DB, module *models.Module) (*models.Module, error)
	Delete(tx *gorm.DB, moduleId int, isHardDelete bool) error
	Restore(tx *gorm.DB, moduleId int) (*models.Module, error)
}

// ==============================
// Implementation
// ==============================

type ModuleRepositoryImpl struct {
	DB *gorm.DB
}

func NewModuleRepository(db *gorm.DB) *ModuleRepositoryImpl {
	return &ModuleRepositoryImpl{DB: db}
}

func (r *ModuleRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *ModuleRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Module, error) {
	var modules []models.Module
	if err := r.useDB(tx).
		Unscoped().
		Preload("Parent").
		Preload("Children").
		Preload("ModuleType").
		Find(&modules).Error; err != nil {
		return nil, HandleDatabaseError(err, "module")
	}
	return modules, nil
}

func (r *ModuleRepositoryImpl) FindById(tx *gorm.DB, moduleId int, includeTrashed bool) (*models.Module, error) {
	var module models.Module
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Parent").
		Preload("Children").
		Preload("ModuleType").
		First(&module, "id = ?", moduleId).Error; err != nil {
		return nil, HandleDatabaseError(err, "module")
	}
	return &module, nil
}

// ---------- Mutations ----------

func (r *ModuleRepositoryImpl) Insert(tx *gorm.DB, module *models.Module) (*models.Module, error) {
	if err := r.useDB(tx).Create(module).Error; err != nil {
		return nil, HandleDatabaseError(err, "module")
	}
	return module, nil
}

func (r *ModuleRepositoryImpl) Update(tx *gorm.DB, module *models.Module) (*models.Module, error) {
	if err := r.useDB(tx).Save(module).Error; err != nil {
		return nil, HandleDatabaseError(err, "module")
	}
	return module, nil
}

func (r *ModuleRepositoryImpl) Delete(tx *gorm.DB, moduleId int, isHardDelete bool) error {
	db := r.useDB(tx)

	var module models.Module
	if err := db.Unscoped().First(&module, "id = ?", moduleId).Error; err != nil {
		return HandleDatabaseError(err, "module")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&module).Error; err != nil {
			return HandleDatabaseError(err, "module")
		}
	} else {
		if err := db.Delete(&module).Error; err != nil {
			return HandleDatabaseError(err, "module")
		}
	}
	return nil
}

func (r *ModuleRepositoryImpl) Restore(tx *gorm.DB, moduleId int) (*models.Module, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.Module{}).
		Where("id = ?", moduleId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "module")
	}

	var restored models.Module
	if err := db.
		Preload("Parent").
		Preload("Children").
		Preload("ModuleType").
		First(&restored, "id = ?", moduleId).Error; err != nil {
		return nil, HandleDatabaseError(err, "module")
	}
	return &restored, nil
}
