package repositories

import (
	"errors"
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindAll() ([]models.User, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.User, int64, error)
	FindByEmailOrUsername(email string) (*models.User, error)
	FindById(userId string, isSoftDelete bool) (*models.User, error)
	Insert(user *models.User) (*models.User, error)
	Update(user *models.User) (*models.User, error)
	Delete(userId string, isHardDelete bool) error
	Restore(user *models.User, userId string) (*models.User, error)
}

type UserRepositoryImpl struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{DB: db}
}

func (r *UserRepositoryImpl) FindAll() ([]models.User, error) {
	var users []models.User
	if err := r.DB.Unscoped().
	 Preload("Avatar").
		Preload("Role").
		Find(&users).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}
	return users, nil
}

func (r *UserRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.User, int64, error) {
	var users []models.User
	var totalCount int64

	query := r.DB.Unscoped().
		Preload("Avatar").
		Preload("Role")

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


func (r *UserRepositoryImpl) FindByEmailOrUsername(identifier string) (*models.User, error) {
	var user *models.User

	err := r.DB.
		Preload("Avatar").
		Preload("Role").
		Where("email = ?", identifier).
		First(&user).Error

	if err == nil {
		return user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, HandleDatabaseError(err, "user")
	}

	err = r.DB.
		Preload("Avatar").
		Preload("Role").
		Where("username = ?", identifier).
		First(&user).Error

	if err != nil {
		return nil, HandleDatabaseError(err, "user")
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindById(userId string, isSoftDelete bool) (*models.User, error) {
	var user *models.User
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	err := db.
	 Preload("Avatar").
		Preload("Role").
		First(&user, "id = ?", userId).Error

	if err != nil {
		return nil, HandleDatabaseError(err, "user")
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindByUsername(username string) (*models.User, error) {
	var user *models.User
	err := r.DB.
	 Preload("Avatar").
		Preload("Role").
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		return nil, HandleDatabaseError(err, "user")
	}

	return user, nil
}


func (r *UserRepositoryImpl) Insert(user *models.User) (*models.User, error) {

	if user.ID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	if err := r.DB.Create(&user).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}
	return user, nil
}

func (r *UserRepositoryImpl) Update(user *models.User) (*models.User, error) {

	if user.ID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	if err := r.DB.Save(&user).Error; err != nil {
		return nil, HandleDatabaseError(err, "user")
	}
	return user, nil
}

func (r *UserRepositoryImpl) Delete(userId string, isHardDelete bool) error {
	var user *models.User

	if err := r.DB.Unscoped().First(&user, "id = ?", userId).Error; err != nil {
		return HandleDatabaseError(err, "user")
	}

	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&user).Error; err != nil {
			return HandleDatabaseError(err, "user")
		}
	} else {
		if err := r.DB.Delete(&user).Error; err != nil {
			return HandleDatabaseError(err, "user")
		}
	}
	return nil
}

func (r *UserRepositoryImpl) Restore(user *models.User, userID string) (*models.User, error) {
	if err := r.DB.Unscoped().Model(user).Where("id = ?", userID).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredUser *models.User
	if err := r.DB.Unscoped().First(&restoredUser, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	
	return restoredUser, nil
}


