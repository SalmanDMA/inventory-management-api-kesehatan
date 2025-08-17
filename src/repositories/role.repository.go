package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository interface {
	FindAll() ([]models.Role, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.Role, int64, error)
	FindById(roleId string, isSoftDelete bool) (*models.Role, error)
	FindByName(roleName string) (*models.Role, error)
	FindByAlias(roleAlias string) (*models.Role, error)
	Insert(role *models.Role) (*models.Role, error)
	Update(role *models.Role) (*models.Role, error)
	Delete(roleId string, isHardDelete bool) error
	Restore(role *models.Role, roleId string) (*models.Role, error)
}

type RoleRepositoryImpl struct{
	DB *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepositoryImpl {
	return &RoleRepositoryImpl{DB: db}
}

func (r *RoleRepositoryImpl) FindAll() ([]models.Role, error) {
	var roles []models.Role
	if err := r.DB.
	 Unscoped().
	 Find(&roles).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return roles, nil
}

func (r *RoleRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.Role, int64, error) {
	var roles []models.Role
	var totalCount int64

	query := r.DB.Unscoped()

	switch req.Status {
	case "active":
		query = query.Where("deleted_at IS NULL")
	case "deleted":
		query = query.Where("deleted_at IS NOT NULL")
	case "all":
	default:
		query = query.Where("deleted_at IS NULL")
	}

	if req.RoleID != "" {
		if roleUUID, err := uuid.Parse(req.RoleID); err == nil {
			query = query.Where("role_id = ?", roleUUID)
		}
	}

	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(alias) LIKE ? OR LOWER(description) LIKE ? OR LOWER(color) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Model(&models.Role{}).Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "role")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&roles).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "role")
	}

	return roles, totalCount, nil
}

func (r *RoleRepositoryImpl) FindById(roleId string, isSoftDelete bool) (*models.Role, error) {
	var role *models.Role
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.First(&role, "id = ?", roleId).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	
	return role, nil
}

func (r *RoleRepositoryImpl) FindByName(roleName string) (*models.Role, error) {
	var role *models.Role
	if err := r.DB.Where("name = ?", roleName).First(&role).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return role, nil
}

func (r *RoleRepositoryImpl) FindByAlias(roleAlias string) (*models.Role, error) {
	var role *models.Role
	if err := r.DB.Where("alias = ?", roleAlias).First(&role).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return role, nil
}

func (r *RoleRepositoryImpl) Insert(role *models.Role) (*models.Role, error) {

	if role.ID == uuid.Nil {
		return nil, fmt.Errorf("role ID cannot be empty")
	}

	if err := r.DB.Create(&role).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return role, nil
}

func (r *RoleRepositoryImpl) Update(role *models.Role) (*models.Role, error) {

	if role.ID == uuid.Nil {
		return nil, fmt.Errorf("role ID cannot be empty")
	}

	if err := r.DB.Save(&role).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return role, nil
}

func (r *RoleRepositoryImpl) Delete(roleId string, isHardDelete bool) error {
	var role *models.Role

	if err := r.DB.Unscoped().First(&role, "id = ?", roleId).Error; err != nil {
		return HandleDatabaseError(err, "role")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&role).Error; err != nil {
			return HandleDatabaseError(err, "role")
		}
	} else {
		if err := r.DB.Delete(&role).Error; err != nil {
			return HandleDatabaseError(err, "role")
		}
	}
	return nil
}

func (r *RoleRepositoryImpl) Restore(role *models.Role, roleId string) (*models.Role, error) {
	if err := r.DB.Unscoped().Model(role).Where("id = ?", roleId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredRole *models.Role
	if err := r.DB.Unscoped().First(&restoredRole, "id = ?", roleId).Error; err != nil {
		return nil, err
	}
	
	return restoredRole, nil
}