package repositories

import (
	"errors"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type RoleModuleRepository interface {
	FindAll(tx *gorm.DB, roleID uuid.UUID) ([]models.RoleModule, error)
	FindByRoleAndModule(tx *gorm.DB, roleID uuid.UUID, moduleID int) (*models.RoleModule, error)
	Insert(tx *gorm.DB, roleModule *models.RoleModule) (*models.RoleModule, error)
	Update(tx *gorm.DB, roleModule *models.RoleModule) (*models.RoleModule, error)
}

// ==============================
// Implementation
// ==============================

type RoleModuleRepositoryImpl struct {
	DB *gorm.DB
}

func NewRoleModuleRepository(db *gorm.DB) *RoleModuleRepositoryImpl {
	return &RoleModuleRepositoryImpl{DB: db}
}

func (r *RoleModuleRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *RoleModuleRepositoryImpl) FindAll(tx *gorm.DB, roleID uuid.UUID) ([]models.RoleModule, error) {
	var roleModules []models.RoleModule
	if err := r.useDB(tx).
		Preload("Role").
		Preload("Module").
		Preload("Module.ModuleType").
		Where("role_id = ?", roleID).
		Find(&roleModules).Error; err != nil {
		return nil, HandleDatabaseError(err, "role_module")
	}
	return roleModules, nil
}

func (r *RoleModuleRepositoryImpl) FindByRoleAndModule(tx *gorm.DB, roleID uuid.UUID, moduleID int) (*models.RoleModule, error) {
	var rm models.RoleModule
	err := r.useDB(tx).
		Preload("Role").
		Preload("Module").
		Preload("Module.ModuleType").
		Where("role_id = ? AND module_id = ?", roleID, moduleID).
		First(&rm).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, HandleDatabaseError(err, "role_module")
	}
	return &rm, nil
}

// ---------- Mutations ----------

func (r *RoleModuleRepositoryImpl) Insert(tx *gorm.DB, roleModule *models.RoleModule) (*models.RoleModule, error) {
	if err := r.useDB(tx).Create(roleModule).Error; err != nil {
		return nil, HandleDatabaseError(err, "role_module")
	}
	return roleModule, nil
}

func (r *RoleModuleRepositoryImpl) Update(tx *gorm.DB, roleModule *models.RoleModule) (*models.RoleModule, error) {
	if err := r.useDB(tx).Save(roleModule).Error; err != nil {
		return nil, HandleDatabaseError(err, "role_module")
	}
	return roleModule, nil
}
