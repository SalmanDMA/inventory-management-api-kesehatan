package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserService struct {
	UserRepository   repositories.UserRepository
	UploadRepository repositories.UploadRepository
}

func NewUserService(userRepo repositories.UserRepository, uploadRepo repositories.UploadRepository) *UserService {
	return &UserService{
		UserRepository:   userRepo,
		UploadRepository: uploadRepo,
	}
}

func (service *UserService) GetAllUsers(userInfo *models.User) ([]models.ResponseGetUser, error) {
	users, err := service.UserRepository.FindAll()
	if err != nil {
		return nil, err
	}

	usersResponse := []models.ResponseGetUser{}
	for _, user := range users {
		usersResponse = append(usersResponse, models.ResponseGetUser{
			ID:          user.ID,
			Username:    user.Username,
			Name:        user.Name,
			Email:       user.Email,
			Address:     user.Address,
			Phone:       user.Phone,
			Avatar:      user.Avatar,
			AvatarID:    user.AvatarID,
			Role:        user.Role,
			RoleID:      *user.RoleID,
			Description: user.Description,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			DeletedAt:   user.DeletedAt,
		})
	}

	return usersResponse, nil
}

func (service *UserService) GetAllUsersPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.UserPaginatedResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Status == "" {
		req.Status = "active"
	}

	users, totalCount, err := service.UserRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	usersResponse := []models.ResponseGetUser{}
	for _, user := range users {
		if strings.ToLower(userInfo.Role.Name) != "developer" && strings.ToLower(user.Role.Name) == "developer" {
			continue
		}

		usersResponse = append(usersResponse, models.ResponseGetUser{
			ID:          user.ID,
			Username:    user.Username,
			Name:        user.Name,
			Email:       user.Email,
			Address:     user.Address,
			Phone:       user.Phone,
			Avatar:      user.Avatar,
			AvatarID:    user.AvatarID,
			Role:        user.Role,
			RoleID:      *user.RoleID,
			Description: user.Description,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			DeletedAt:   user.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	paginationResponse := models.PaginationResponse{
		CurrentPage:  req.Page,
		PerPage:      req.Limit,
		TotalPages:   totalPages,
		TotalRecords: totalCount,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
	}

	return &models.UserPaginatedResponse{
		Data:       usersResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *UserService) CreateUser(userRequest *models.UserCreate, ctx *fiber.Ctx, userInfo *models.User) (*models.User, error) {
	if _, err := service.UserRepository.FindByEmailOrUsername(userRequest.Email); err == nil {
		return nil, errors.New("email or username already exists")
	} else if err != repositories.ErrUserNotFound {
		return nil, errors.New("error checking email: " + err.Error())
	}

	if userRequest.Username != userRequest.Email {
		if _, err := service.UserRepository.FindByEmailOrUsername(userRequest.Username); err == nil {
			return nil, errors.New("email or username already exists")
		} else if err != repositories.ErrUserNotFound {
			return nil, errors.New("error checking username: " + err.Error())
		}
	}

	passwordHash, err := helpers.HashPassword(userRequest.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	newUser := &models.User{
		ID:          uuid.New(),
		Username:    userRequest.Username,
		Name:        userRequest.Name,
		Email:       userRequest.Email,
		Address:     userRequest.Address,
		Phone:       userRequest.Phone,
		Password:    passwordHash,
		RoleID:      &userRequest.RoleID,
		Description: userRequest.Description,
	}

	var avatarUUIDStr string
	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if avatarUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(avatarUUIDStr)
			}
		}
	}()

	file, err := ctx.FormFile("avatar")
	if err == nil && file != nil {
		avatarUUIDStr, err = helpers.SaveFile(ctx, file, "users")
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to save avatar: %w", err)
		}

		avatarUUID, err := uuid.Parse(avatarUUIDStr)
		if err != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(avatarUUIDStr)
			return nil, fmt.Errorf("invalid avatar UUID: %w", err)
		}
		newUser.AvatarID = &avatarUUID
	}

	result := tx.Create(newUser)
	if result.Error != nil {
		tx.Rollback()
		if avatarUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(avatarUUIDStr)
		}
		
		if strings.Contains(result.Error.Error(), "duplicate key") {
			return nil, errors.New("email or username already exists")
		}
		return nil, fmt.Errorf("error creating user: %w", result.Error)
	}

	if err := tx.Commit().Error; err != nil {
		if avatarUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(avatarUUIDStr)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	insertedUser, err := service.UserRepository.FindById(newUser.ID.String(), true)
	if err != nil {
		log.Printf("Warning: User created but failed to fetch created data: %v", err)
		return newUser, nil
	}

	return insertedUser, nil
}

func (service *UserService) UpdateUser(userRequest *models.UserUpdate, userId string, ctx *fiber.Ctx, userInfo *models.User) (*models.User, error) {
	user, err := service.UserRepository.FindById(userId, true)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	if userRequest.Email != "" && userRequest.Email != user.Email {
		existingUser, err := service.UserRepository.FindByEmailOrUsername(userRequest.Email)
		if err == nil && existingUser.ID != user.ID {
			return nil, errors.New("email or username already exists")
		} else if err != repositories.ErrUserNotFound {
			return nil, fmt.Errorf("error checking existing email: %w", err)
		}
	}

	if userRequest.Username != "" && userRequest.Username != user.Username {
		existingUser, err := service.UserRepository.FindByEmailOrUsername(userRequest.Username)
		if err == nil && existingUser.ID != user.ID {
			return nil, errors.New("email or username already exists")
		} else if err != repositories.ErrUserNotFound {
			return nil, fmt.Errorf("error checking existing username: %w", err)
		}
	}

	if userRequest.Email != "" && userRequest.Email != user.Email {
		user.Email = userRequest.Email
	}
	if userRequest.Name != "" && userRequest.Name != user.Name {
		user.Name = userRequest.Name
	}
	if userRequest.Username != "" && userRequest.Username != user.Username {
		user.Username = userRequest.Username
	}
	if userRequest.Address != "" {
		user.Address = userRequest.Address
	}
	if userRequest.Phone != "" {
		user.Phone = userRequest.Phone
	}
	if userRequest.Description != "" {
		user.Description = userRequest.Description
	}
	if userRequest.RoleID != uuid.Nil && (user.RoleID == nil || *user.RoleID != userRequest.RoleID) {
		user.RoleID = &userRequest.RoleID
	}

	file, err := ctx.FormFile("avatar")
	if err == nil && file != nil {
		oldAvatarID := user.AvatarID
		var newAvatarUUIDStr string
		var newAvatarID uuid.UUID

		tx := configs.DB.Begin()
		if tx.Error != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				if newAvatarUUIDStr != "" {
					helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
				}
			}
		}()

		newAvatarUUIDStr, err := helpers.SaveFile(ctx, file, "users")
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to save avatar: %w", err)
		}

		newAvatarID, err = uuid.Parse(newAvatarUUIDStr)
		if err != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
			return nil, fmt.Errorf("invalid avatar UUID: %w", err)
		}

		user.Avatar = nil
		user.AvatarID = &newAvatarID
		result := tx.Model(user).Select("*").Updates(user)
		if result.Error != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
			
			if strings.Contains(result.Error.Error(), "duplicate key") {
				return nil, errors.New("email or username already exists")
			}
			return nil, fmt.Errorf("error updating user with avatar: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
			return nil, errors.New("no rows affected, user may not exist")
		}

		if err := tx.Commit().Error; err != nil {
			helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		updatedUser, err := service.UserRepository.FindById(user.ID.String(), true)
		if err != nil {
			log.Printf("Warning: User updated but failed to fetch updated data: %v", err)
			return user, nil
		}

		if oldAvatarID != nil {
			go func() {
				if err := helpers.DeleteLocalFileImmediate(oldAvatarID.String()); err != nil {
					log.Printf("Warning: Failed to delete old avatar file %s: %v", oldAvatarID.String(), err)
				}
			}()
		}

		return updatedUser, nil
	}

	updatedUser, err := service.UserRepository.Update(user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, errors.New("email or username already exists")
		}
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return updatedUser, nil
}

func (service *UserService) UpdateUserProfile(userID string, userUpdate *models.UserUpdateProfileRequest) (*models.User, error) {
	user, err := service.UserRepository.FindById(userID, true)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	if userUpdate.Email != "" && userUpdate.Email != user.Email {
		existingUser, err := service.UserRepository.FindByEmailOrUsername(userUpdate.Email)
		switch {
		case err == nil && existingUser.ID != user.ID:
			return nil, errors.New("email already exists")
		case err != nil && err != repositories.ErrUserNotFound:
			return nil, fmt.Errorf("error checking existing email: %w", err)
		}
		user.Email = userUpdate.Email
	}

	if userUpdate.Name != "" && userUpdate.Name != user.Name {
		user.Name = userUpdate.Name
	}
	if userUpdate.Address != "" {
		user.Address = userUpdate.Address
	}
	if userUpdate.Phone != "" {
		user.Phone = userUpdate.Phone
	}
	if userUpdate.Description != "" {
		user.Description = userUpdate.Description
	}

	updatedUser, err := service.UserRepository.Update(user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, errors.New("email already exists")
		}
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return updatedUser, nil
}

func (service *UserService) UpdateAvatarOnly(userId string, ctx *fiber.Ctx) (*models.User, error) {
	user, err := service.UserRepository.FindById(userId, true)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	file, err := ctx.FormFile("avatar")
	if err != nil {
		return nil, fmt.Errorf("failed to read avatar file: %w", err)
	}
	if file == nil {
		return nil, errors.New("avatar file is required")
	}

	oldAvatarID := user.AvatarID
	var newAvatarUUIDStr string
	var newAvatarID uuid.UUID

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if newAvatarUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
			}
		}
	}()

	newAvatarUUIDStr, err = helpers.SaveFile(ctx, file, "users")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to save avatar: %w", err)
	}

	newAvatarID, err = uuid.Parse(newAvatarUUIDStr)
	if err != nil {
		tx.Rollback()
		helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
		return nil, fmt.Errorf("invalid avatar UUID: %w", err)
	}

	user.AvatarID = &newAvatarID
	result := tx.Model(user).Select("avatar_id", "updated_at").Updates(map[string]interface{}{
		"avatar_id":  newAvatarID,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		tx.Rollback()
		helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
		return nil, fmt.Errorf("error updating user avatar_id: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
		return nil, errors.New("no rows affected, user may not exist")
	}

	if err := tx.Commit().Error; err != nil {
		helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	updatedUser, err := service.UserRepository.FindById(userId, true)
	if err != nil {
		log.Printf("Warning: User updated but failed to fetch updated data: %v", err)
		return user, nil
	}

	if oldAvatarID != nil {
		go func() {
			if err := helpers.DeleteLocalFileImmediate(oldAvatarID.String()); err != nil {
				log.Printf("Warning: Failed to delete old avatar file %s: %v", oldAvatarID.String(), err)
			}
		}()
	}

	return updatedUser, nil
}

func (service *UserService) DeleteUsers(userRequest *models.UserIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, userId := range userRequest.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for user %v: %v\n", userId, tx.Error)
			return errors.New("error beginning transaction")
		}

		user, err := service.UserRepository.FindById(userId.String(), false)
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrUserNotFound {
				log.Printf("User not found: %v\n", userId)
				continue
			}
			log.Printf("Error finding user %v: %v\n", userId, err)
			return errors.New("error finding user")
		}

		if userRequest.IsHardDelete == "hardDelete" {
			if user.AvatarID != nil && user.AvatarID.String() != "" {
				if err := tx.Unscoped().Delete(&models.User{}, "id = ?", userId).Error; err != nil {
					tx.Rollback()
					log.Printf("Error hard deleting user %v: %v\n", userId, err)
					return errors.New("error hard deleting user")
				}

				if err := tx.Commit().Error; err != nil {
					log.Printf("Error committing hard delete for user %v: %v\n", userId, err)
					return errors.New("error committing hard delete")
				}

				go func(avatarID string) {
					if err := helpers.DeleteLocalFileImmediate(avatarID); err != nil {
						log.Printf("Warning: Failed to delete avatar file %s: %v", avatarID, err)
					}
				}(user.AvatarID.String())
			} else {
				if err := tx.Unscoped().Delete(&models.User{}, "id = ?", userId).Error; err != nil {
					tx.Rollback()
					log.Printf("Error hard deleting user %v: %v\n", userId, err)
					return errors.New("error hard deleting user")
				}

				if err := tx.Commit().Error; err != nil {
					log.Printf("Error committing hard delete for user %v: %v\n", userId, err)
					return errors.New("error committing hard delete")
				}
			}
		} else {
			if err := tx.Delete(&models.User{}, "id = ?", userId).Error; err != nil {
				tx.Rollback()
				log.Printf("Error soft deleting user %v: %v\n", userId, err)
				return errors.New("error soft deleting user")
			}

			if err := tx.Commit().Error; err != nil {
				log.Printf("Error committing soft delete for user %v: %v\n", userId, err)
				return errors.New("error committing soft delete")
			}
		}
	}

	return nil
}

func (service *UserService) RestoreUsers(userRequest *models.UserRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.User, error) {
	var restoredUsers []models.User

	for _, userId := range userRequest.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for user restore %v: %v\n", userId, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		result := tx.Model(&models.User{}).Unscoped().Where("id = ?", userId).Update("deleted_at", nil)
		if result.Error != nil {
			tx.Rollback()
			log.Printf("Error restoring user %v: %v\n", userId, result.Error)
			return nil, errors.New("error restoring user")
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			log.Printf("User not found for restore: %v\n", userId)
			continue
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing user restore %v: %v\n", userId, err)
			return nil, errors.New("error committing user restore")
		}

		restoredUser, err := service.UserRepository.FindById(userId.String(), true)
		if err != nil {
			log.Printf("Error fetching restored user %v: %v\n", userId, err)
			continue
		}

		restoredUsers = append(restoredUsers, *restoredUser)
	}

	return restoredUsers, nil
}