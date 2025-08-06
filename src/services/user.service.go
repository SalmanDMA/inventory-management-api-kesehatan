package services

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserService struct {
	UserRepository repositories.UserRepository
	UploadRepository repositories.UploadRepository
}

func NewUserService(userRepo repositories.UserRepository, uploadRepo repositories.UploadRepository) *UserService {
	return &UserService{
					UserRepository: userRepo,
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

	passwordHash, err := helpers.HashPassword(userRequest.Password)
	if err != nil {
		return nil, err
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

	file, err := ctx.FormFile("avatar")
	if err == nil {
		uuidStr, err := helpers.SaveFile(ctx, file, "users")
		if err != nil {
			return nil, err
		}
		avatarUUID, err := uuid.Parse(uuidStr)
		if err != nil {
			return nil, err
		}
		newUser.AvatarID = &avatarUUID
	}

	insertedUser, err := service.UserRepository.Insert(newUser)
	if err != nil {
		return nil, err
	}

	return insertedUser, nil
}


func (service *UserService) UpdateUser(userRequest *models.UserUpdate, userId string, ctx *fiber.Ctx, userInfo *models.User) (*models.User, error) {
	user, err := service.UserRepository.FindById(userId, true)
	if err != nil {
		return nil, err
	}

	if userRequest.Email != "" && userRequest.Email != user.Email {
		existingUser, err := service.UserRepository.FindByEmailOrUsername(userRequest.Email)
		if err == nil && existingUser.ID != user.ID {
			return nil, errors.New("email or username already exists")
		}
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
		avatarIDStr, err := helpers.SaveFile(ctx, file, "users")
		if err != nil {
			return nil, fmt.Errorf("failed to save avatar: %w", err)
		}

		newAvatarID, err := uuid.Parse(avatarIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid avatar ID: %w", err)
		}

		oldAvatarID := user.AvatarID
		user.AvatarID = &newAvatarID

		updatedUser, err := service.UserRepository.Update(user)
		if err != nil {
			return nil, fmt.Errorf("error updating user with avatar: %w", err)
		}

		if oldAvatarID != nil {
			service.UploadRepository.Delete(oldAvatarID.String(), false)
		}

		return updatedUser, nil
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


func (service *UserService) UpdateUserProfile(userID string, userUpdate *models.UserUpdateProfileRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.User, error) {
	user, err := service.UserRepository.FindById(userID, true)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	if userUpdate.Email != "" && userUpdate.Email != user.Email {
		existingUser, err := service.UserRepository.FindByEmailOrUsername(userUpdate.Email)
		if err == nil && existingUser.ID != user.ID {
			return nil, errors.New("email already exists")
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

	file, err := ctx.FormFile("avatar")
	if err == nil && file != nil {
		uuidStr, err := helpers.SaveFile(ctx, file, "users")
		if err != nil {
			return nil, fmt.Errorf("failed to save avatar: %w", err)
		}

		newAvatarID, err := uuid.Parse(uuidStr)
		if err != nil {
			return nil, fmt.Errorf("invalid avatar UUID: %w", err)
		}

		oldAvatarID := user.AvatarID
		user.AvatarID = &newAvatarID

		updatedUser, err := service.UserRepository.Update(user)
		if err != nil {
			return nil, fmt.Errorf("error updating user with avatar: %w", err)
		}

		if oldAvatarID != nil {
			service.UploadRepository.Delete(oldAvatarID.String(), false)
		}

		return updatedUser, nil
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

func (service *UserService) DeleteUsers(userRequest *models.UserIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, userId := range userRequest.IDs {
		user, err := service.UserRepository.FindById(userId.String(), false)
		if err != nil {
			if err == repositories.ErrUserNotFound {
				log.Printf("User not found: %v\n", userId)
				continue
			}
			log.Printf("Error finding user %v: %v\n", userId, err)
			return errors.New("error finding user")
		}

		if userRequest.IsHardDelete == "hardDelete" {
			if user.AvatarID != nil && user.AvatarID.String() != "" {
				if err := service.UploadRepository.Delete(user.AvatarID.String(), false); err != nil {
					log.Printf("Error deleting avatar file for user %v: %v\n", userId, err)
					return errors.New("error deleting avatar file")
				}
			}

			if err := service.UserRepository.Delete(userId.String(), true); err != nil {
				log.Printf("Error hard deleting user %v: %v\n", userId, err)
				return errors.New("error hard deleting user")
			}
		} else {
			if err := service.UserRepository.Delete(userId.String(), false); err != nil {
				log.Printf("Error soft deleting user %v: %v\n", userId, err)
				return errors.New("error soft deleting user")
			}
		}

	}

	return nil
}


func (service *UserService) RestoreUsers(userRequest *models.UserRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.User, error) {
	var restoredUsers []models.User

	for _, userId := range userRequest.IDs {
		user := &models.User{ID: userId}

		restoredUser, err := service.UserRepository.Restore(user, userId.String())
		if err != nil {
			if err == repositories.ErrUserNotFound {
				log.Printf("User not found: %v\n", userId)
				continue
			}
			log.Printf("Error restoring user %v: %v\n", userId, err)
			return nil, errors.New("error restoring user")
		}

		restoredUsers = append(restoredUsers, *restoredUser)
	}

	return restoredUsers, nil
}

