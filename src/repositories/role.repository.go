package repositories

import (
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type RoleRepository interface {
	FindAll(tx *gorm.DB) ([]models.Role, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Role, int64, error)
	FindById(tx *gorm.DB, roleId string, includeTrashed bool) (*models.Role, error)
	FindByName(tx *gorm.DB, roleName string) (*models.Role, error)
	FindByAlias(tx *gorm.DB, roleAlias string) (*models.Role, error)
	Insert(tx *gorm.DB, role *models.Role) (*models.Role, error)
	Update(tx *gorm.DB, role *models.Role) (*models.Role, error)
	Delete(tx *gorm.DB, roleId string, isHardDelete bool) error
	Restore(tx *gorm.DB, roleId string) (*models.Role, error)
}

// ==============================
// Implementation
// ==============================

type RoleRepositoryImpl struct {
	DB *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepositoryImpl {
	return &RoleRepositoryImpl{DB: db}
}

func (r *RoleRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *RoleRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Role, error) {
	var roles []models.Role
	if err := r.useDB(tx).
		Unscoped().
		Find(&roles).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return roles, nil
}

func (r *RoleRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.Role, int64, error) {
	var (
		roles      []models.Role
		totalCount int64
	)

	query := r.useDB(tx).Unscoped().Model(&models.Role{})

	switch req.Status {
	case "active":
		query = query.Where("deleted_at IS NULL")
	case "deleted":
		query = query.Where("deleted_at IS NOT NULL")
	case "all":
		// no filter
	default:
		query = query.Where("deleted_at IS NULL")
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		p := "%" + strings.ToLower(s) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(alias) LIKE ? OR LOWER(COALESCE(description, '')) LIKE ? OR LOWER(COALESCE(color, '')) LIKE ?",
			p, p, p, p,
		)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "role")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&roles).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "role")
	}

	return roles, totalCount, nil
}

func (r *RoleRepositoryImpl) FindById(tx *gorm.DB, roleId string, includeTrashed bool) (*models.Role, error) {
	var role models.Role
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.First(&role, "id = ?", roleId).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return &role, nil
}

func (r *RoleRepositoryImpl) FindByName(tx *gorm.DB, roleName string) (*models.Role, error) {
	var role models.Role
	if err := r.useDB(tx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return &role, nil
}

func (r *RoleRepositoryImpl) FindByAlias(tx *gorm.DB, roleAlias string) (*models.Role, error) {
	var role models.Role
	if err := r.useDB(tx).Where("alias = ?", roleAlias).First(&role).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return &role, nil
}

// ---------- Mutations ----------

func (r *RoleRepositoryImpl) Insert(tx *gorm.DB, role *models.Role) (*models.Role, error) {
	if role.ID == uuid.Nil {
		return nil, fmt.Errorf("role ID cannot be empty")
	}
	if err := r.useDB(tx).Create(role).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return role, nil
}

func (r *RoleRepositoryImpl) Update(tx *gorm.DB, role *models.Role) (*models.Role, error) {
	if role.ID == uuid.Nil {
		return nil, fmt.Errorf("role ID cannot be empty")
	}
	if err := r.useDB(tx).Save(role).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return role, nil
}

func (r *RoleRepositoryImpl) Delete(tx *gorm.DB, roleId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var role models.Role
	if err := db.Unscoped().First(&role, "id = ?", roleId).Error; err != nil {
		return HandleDatabaseError(err, "role")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&role).Error; err != nil {
			return HandleDatabaseError(err, "role")
		}
	} else {
		if err := db.Delete(&role).Error; err != nil {
			return HandleDatabaseError(err, "role")
		}
	}
	return nil
}

func (r *RoleRepositoryImpl) Restore(tx *gorm.DB, roleId string) (*models.Role, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.Role{}).
		Where("id = ?", roleId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}

	var restored models.Role
	if err := db.First(&restored, "id = ?", roleId).Error; err != nil {
		return nil, HandleDatabaseError(err, "role")
	}
	return &restored, nil
}
