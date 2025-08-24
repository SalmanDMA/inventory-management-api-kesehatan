package repositories

import (
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type ModuleTypeRepository interface {
	FindAll(tx *gorm.DB) ([]models.ModuleType, error)
	FindById(tx *gorm.DB, moduleTypeId string, includeTrashed bool) (*models.ModuleType, error)
	Insert(tx *gorm.DB, moduleType *models.ModuleType) (*models.ModuleType, error)
	Update(tx *gorm.DB, moduleType *models.ModuleType) (*models.ModuleType, error)
	Delete(tx *gorm.DB, moduleTypeId string, isHardDelete bool) error
	Restore(tx *gorm.DB, moduleTypeId string) (*models.ModuleType, error)
}

// ==============================
// Implementation
// ==============================

type ModuleTypeRepositoryImpl struct {
	DB *gorm.DB
}

func NewModuleTypeRepository(db *gorm.DB) *ModuleTypeRepositoryImpl {
	return &ModuleTypeRepositoryImpl{DB: db}
}

func (r *ModuleTypeRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *ModuleTypeRepositoryImpl) FindAll(tx *gorm.DB) ([]models.ModuleType, error) {
	var moduleTypes []models.ModuleType
	if err := r.useDB(tx).
		Unscoped().
		Find(&moduleTypes).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}
	return moduleTypes, nil
}

func (r *ModuleTypeRepositoryImpl) FindById(tx *gorm.DB, moduleTypeId string, includeTrashed bool) (*models.ModuleType, error) {
	var mt models.ModuleType
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.First(&mt, "id = ?", moduleTypeId).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}
	return &mt, nil
}

// ---------- Mutations ----------

func (r *ModuleTypeRepositoryImpl) Insert(tx *gorm.DB, moduleType *models.ModuleType) (*models.ModuleType, error) {
	if moduleType.ID == uuid.Nil {
		return nil, fmt.Errorf("moduleType ID cannot be empty")
	}
	if err := r.useDB(tx).Create(moduleType).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}
	return moduleType, nil
}

func (r *ModuleTypeRepositoryImpl) Update(tx *gorm.DB, moduleType *models.ModuleType) (*models.ModuleType, error) {
	if moduleType.ID == uuid.Nil {
		return nil, fmt.Errorf("moduleType ID cannot be empty")
	}
	if err := r.useDB(tx).Save(moduleType).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}
	return moduleType, nil
}

func (r *ModuleTypeRepositoryImpl) Delete(tx *gorm.DB, moduleTypeId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var mt models.ModuleType
	if err := db.Unscoped().First(&mt, "id = ?", moduleTypeId).Error; err != nil {
		return HandleDatabaseError(err, "module_type")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&mt).Error; err != nil {
			return HandleDatabaseError(err, "module_type")
		}
	} else {
		if err := db.Delete(&mt).Error; err != nil {
			return HandleDatabaseError(err, "module_type")
		}
	}
	return nil
}

func (r *ModuleTypeRepositoryImpl) Restore(tx *gorm.DB, moduleTypeId string) (*models.ModuleType, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.ModuleType{}).
		Where("id = ?", moduleTypeId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}

	var restored models.ModuleType
	if err := db.First(&restored, "id = ?", moduleTypeId).Error; err != nil {
		return nil, HandleDatabaseError(err, "module_type")
	}
	return &restored, nil
}
