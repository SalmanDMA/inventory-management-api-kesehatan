package repositories

import (
	"errors"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleModuleRepository interface {
	FindAll(roleID uuid.UUID) ([]models.RoleModule, error)
	FindByRoleAndModule(roleID uuid.UUID, moduleID int) (*models.RoleModule, error)
	Insert(roleModule *models.RoleModule) (*models.RoleModule, error)
	Update(roleModule *models.RoleModule) (*models.RoleModule, error)
}

type RoleModuleRepositoryImpl struct {
	DB *gorm.DB
}

func NewRoleModuleRepository(db *gorm.DB) *RoleModuleRepositoryImpl {
	return &RoleModuleRepositoryImpl{DB: db}
}

func (r *RoleModuleRepositoryImpl) FindAll(roleID uuid.UUID) ([]models.RoleModule, error) {
	var roleModules []models.RoleModule

	err := r.DB.
		Preload("Role").
		Preload("Module").
		Preload("Module.ModuleType").
		Where("role_id = ?", roleID).
		Find(&roleModules).Error

	if err != nil {
		return nil, HandleDatabaseError(err, "role_module")
	}

	return roleModules, nil
}

func (r *RoleModuleRepositoryImpl) FindByRoleAndModule(roleID uuid.UUID, moduleID int) (*models.RoleModule, error) {
	var roleModule models.RoleModule
	err := r.DB.
	Preload("Role").
	Preload("Module").
	Preload("Module.ModuleType").
	Where("role_id = ? AND module_id = ?", roleID, moduleID).
	First(&roleModule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, HandleDatabaseError(err, "role_module")
	}
	return &roleModule, nil
}

func (r *RoleModuleRepositoryImpl) Insert(roleModule *models.RoleModule) (*models.RoleModule, error) {
	if err := r.DB.Create(&roleModule).Error; err != nil {
		return nil, HandleDatabaseError(err, "role_module")
	}
	return roleModule, nil
}

func (r *RoleModuleRepositoryImpl) Update(roleModule *models.RoleModule) (*models.RoleModule, error) {
	if err := r.DB.Save(&roleModule).Error; err != nil {
		return nil, HandleDatabaseError(err, "role_module")
	}
	return roleModule, nil
}