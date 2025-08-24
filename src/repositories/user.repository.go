package repositories

import (
	"errors"
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type UserRepository interface {
	FindAll(tx *gorm.DB) ([]models.User, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.User, int64, error)
	FindByEmailOrUsername(tx *gorm.DB, identifier string) (*models.User, error)
	FindById(tx *gorm.DB, userId string, includeTrashed bool) (*models.User, error)
	FindByUsername(tx *gorm.DB, username string) (*models.User, error)
	Insert(tx *gorm.DB, user *models.User) (*models.User, error)
	Update(tx *gorm.DB, user *models.User) (*models.User, error)
	Delete(tx *gorm.DB, userId string, isHardDelete bool) error
	Restore(tx *gorm.DB, userID string) (*models.User, error)
}

// ==============================
// Implementation
// ==============================

type UserRepositoryImpl struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{DB: db}
}

func (r *UserRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *UserRepositoryImpl) FindAll(tx *gorm.DB) ([]models.User, error) {
	var users []models.User
	if err := r.useDB(tx).
		Unscoped().
		Preload("Avatar").
		Preload("Role").
		Find(&users).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}
	return users, nil
}

func (r *UserRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.User, int64, error) {
	var (
		users      []models.User
		totalCount int64
	)

	query := r.useDB(tx).
		Unscoped().
		Preload("Avatar").
		Preload("Role")

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

	if req.RoleID != "" {
		if roleUUID, err := uuid.Parse(req.RoleID); err == nil {
			query = query.Where("role_id = ?", roleUUID)
		}
	}

	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where(
			"LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(username) LIKE LOWER(?)",
			searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Model(&models.User{}).Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "user")
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Find(&users).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "user")
	}

	return users, totalCount, nil
}

func (r *UserRepositoryImpl) FindByEmailOrUsername(tx *gorm.DB, identifier string) (*models.User, error) {
	var user models.User
	db := r.useDB(tx)

	err := db.
		Preload("Avatar").
		Preload("Role").
		Where("email = ?", identifier).
		First(&user).Error

	if err == nil {
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, HandleDatabaseError(err, "user")
	}

	if err := db.
		Preload("Avatar").
		Preload("Role").
		Where("username = ?", identifier).
		First(&user).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}

	return &user, nil
}

func (r *UserRepositoryImpl) FindById(tx *gorm.DB, userId string, includeTrashed bool) (*models.User, error) {
	var user models.User
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Avatar").
		Preload("Role").
		First(&user, "id = ?", userId).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindByUsername(tx *gorm.DB, username string) (*models.User, error) {
	var user models.User
	if err := r.useDB(tx).
		Preload("Avatar").
		Preload("Role").
		Where("username = ?", username).
		First(&user).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}
	return &user, nil
}

// ---------- Mutations ----------

func (r *UserRepositoryImpl) Insert(tx *gorm.DB, user *models.User) (*models.User, error) {
	if user.ID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if err := r.useDB(tx).Create(user).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}
	return user, nil
}

func (r *UserRepositoryImpl) Update(tx *gorm.DB, user *models.User) (*models.User, error) {
	if user.ID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if err := r.useDB(tx).Save(user).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}
	return user, nil
}

func (r *UserRepositoryImpl) Delete(tx *gorm.DB, userId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var user models.User
	if err := db.Unscoped().First(&user, "id = ?", userId).Error; err != nil {
		return HandleDatabaseError(err, "user")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&user).Error; err != nil {
			return HandleDatabaseError(err, "user")
		}
	} else {
		if err := db.Delete(&user).Error; err != nil {
			return HandleDatabaseError(err, "user")
		}
	}
	return nil
}

func (r *UserRepositoryImpl) Restore(tx *gorm.DB, userID string) (*models.User, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}

	var restored models.User
	if err := db.
		Preload("Avatar").
		Preload("Role").
		First(&restored, "id = ?", userID).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}
	return &restored, nil
}
