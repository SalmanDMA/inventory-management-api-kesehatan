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
	"gorm.io/gorm"
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

func (s *UserService) GetAllUsers(userInfo *models.User) ([]models.ResponseGetUser, error) {
	users, err := s.UserRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResponseGetUser, 0, len(users))
	for _, u := range users {
		out = append(out, models.ResponseGetUser{
			ID:          u.ID,
			Username:    u.Username,
			Name:        u.Name,
			Email:       u.Email,
			Address:     u.Address,
			Phone:       u.Phone,
			Avatar:      u.Avatar,
			AvatarID:    u.AvatarID,
			Role:        u.Role,
			RoleID:      *u.RoleID,
			Description: u.Description,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
			DeletedAt:   u.DeletedAt,
		})
	}
	return out, nil
}

func (s *UserService) GetAllUsersPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.UserPaginatedResponse, error) {
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

	users, totalCount, err := s.UserRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetUser, 0, len(users))
	for _, u := range users {
		if strings.ToLower(userInfo.Role.Name) != "developer" && strings.ToLower(u.Role.Name) == "developer" {
			continue
		}
		resp = append(resp, models.ResponseGetUser{
			ID:          u.ID,
			Username:    u.Username,
			Name:        u.Name,
			Email:       u.Email,
			Address:     u.Address,
			Phone:       u.Phone,
			Avatar:      u.Avatar,
			AvatarID:    u.AvatarID,
			Role:        u.Role,
			RoleID:      *u.RoleID,
			Description: u.Description,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
			DeletedAt:   u.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.UserPaginatedResponse{
		Data: resp,
		Pagination: models.PaginationResponse{
			CurrentPage:  req.Page,
			PerPage:      req.Limit,
			TotalPages:   totalPages,
			TotalRecords: totalCount,
			HasNext:      req.Page < totalPages,
			HasPrev:      req.Page > 1,
		},
	}, nil
}

func (s *UserService) GetUserByID(userId string, userInfo *models.User) (*models.User, error) {
	u, err := s.UserRepository.FindById(nil, userId, false)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:          u.ID,
		Username:    u.Username,
		Name:        u.Name,
		Email:       u.Email,
		Address:     u.Address,
		Phone:       u.Phone,
		Avatar:      u.Avatar,
		AvatarID:    u.AvatarID,
		Role:        u.Role,
		RoleID:      u.RoleID,
		Description: u.Description,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		DeletedAt:   u.DeletedAt,
	}, nil
}

func (s *UserService) CreateUser(in *models.UserCreate, ctx *fiber.Ctx, userInfo *models.User) (*models.User, error) {
	// Unik email
	if _, err := s.UserRepository.FindByEmailOrUsername(nil, in.Email); err == nil {
		return nil, errors.New("email or username already exists")
	} else if err != repositories.ErrUserNotFound {
		return nil, fmt.Errorf("error checking email: %w", err)
	}

	// Unik username (kalau beda dengan email)
	if in.Username != in.Email {
		if _, err := s.UserRepository.FindByEmailOrUsername(nil, in.Username); err == nil {
			return nil, errors.New("email or username already exists")
		} else if err != repositories.ErrUserNotFound {
			return nil, fmt.Errorf("error checking username: %w", err)
		}
	}

	passwordHash, err := helpers.HashPassword(in.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	u := &models.User{
		ID:          uuid.New(),
		Username:    in.Username,
		Name:        in.Name,
		Email:       in.Email,
		Address:     in.Address,
		Phone:       in.Phone,
		Password:    passwordHash,
		RoleID:      &in.RoleID,
		Description: in.Description,
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

	// Handle avatar (opsional)
	if file, ferr := ctx.FormFile("avatar"); ferr == nil && file != nil {
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
		u.AvatarID = &avatarUUID
	}

	created, err := s.UserRepository.Insert(tx, u)
	if err != nil {
		tx.Rollback()
		if avatarUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(avatarUUIDStr)
		}
		// Normalisasi pesan duplicate
		if strings.Contains(err.Error(), "duplicate") {
			return nil, errors.New("email or username already exists")
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		if avatarUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(avatarUUIDStr)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// fetch ulang (dengan preload) di outside tx
	out, err := s.UserRepository.FindById(nil, created.ID.String(), true)
	if err != nil {
		log.Printf("Warning: User created but failed to fetch created data: %v", err)
		return created, nil
	}
	return out, nil
}

func (s *UserService) UpdateUser(in *models.UserUpdate, userId string, ctx *fiber.Ctx, userInfo *models.User) (*models.User, error) {
	user, err := s.UserRepository.FindById(nil, userId, true)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	// Validasi unik
	if in.Email != "" && in.Email != user.Email {
		ex, err := s.UserRepository.FindByEmailOrUsername(nil, in.Email)
		if err == nil && ex.ID != user.ID {
			return nil, errors.New("email or username already exists")
		} else if err != nil && err != repositories.ErrUserNotFound {
			return nil, fmt.Errorf("error checking existing email: %w", err)
		}
	}
	if in.Username != "" && in.Username != user.Username {
		ex, err := s.UserRepository.FindByEmailOrUsername(nil, in.Username)
		if err == nil && ex.ID != user.ID {
			return nil, errors.New("email or username already exists")
		} else if err != nil && err != repositories.ErrUserNotFound {
			return nil, fmt.Errorf("error checking existing username: %w", err)
		}
	}

	// Map perubahan
	if in.Email != "" && in.Email != user.Email {
		user.Email = in.Email
	}
	if in.Name != "" && in.Name != user.Name {
		user.Name = in.Name
	}
	if in.Username != "" && in.Username != user.Username {
		user.Username = in.Username
	}
	if in.Address != "" {
		user.Address = in.Address
	}
	if in.Phone != "" {
		user.Phone = in.Phone
	}
	if in.Description != "" {
		user.Description = in.Description
	}
	if in.RoleID != uuid.Nil && (user.RoleID == nil || *user.RoleID != in.RoleID) {
		user.RoleID = &in.RoleID
	}

	// Jika ada avatar baru: proses dalam transaksi terpisah
	if file, ferr := ctx.FormFile("avatar"); ferr == nil && file != nil {
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

		user.Avatar = nil
		user.AvatarID = &newAvatarID

		if _, err := s.UserRepository.Update(tx, user); err != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
			if strings.Contains(err.Error(), "duplicate") {
				return nil, errors.New("email or username already exists")
			}
			return nil, fmt.Errorf("error updating user with avatar: %w", err)
		}

		if err := tx.Commit().Error; err != nil {
			helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		updated, err := s.UserRepository.FindById(nil, user.ID.String(), true)
		if err != nil {
			log.Printf("Warning: User updated but failed to fetch updated data: %v", err)
			return user, nil
		}

		// Hapus avatar lama setelah commit
		if oldAvatarID != nil {
			go func(id string) {
				if err := helpers.DeleteLocalFileImmediate(id); err != nil {
					log.Printf("Warning: Failed to delete old avatar file %s: %v", id, err)
				}
			}(oldAvatarID.String())
		}
		return updated, nil
	}

	// Tanpa avatar baru: update biasa
	updated, err := s.UserRepository.Update(nil, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, errors.New("email or username already exists")
		}
		return nil, fmt.Errorf("error updating user: %w", err)
	}
	return updated, nil
}

func (s *UserService) UpdateUserProfile(userID string, in *models.UserUpdateProfileRequest) (*models.User, error) {
	user, err := s.UserRepository.FindById(nil, userID, true)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	if in.Email != "" && in.Email != user.Email {
		ex, err := s.UserRepository.FindByEmailOrUsername(nil, in.Email)
		switch {
		case err == nil && ex.ID != user.ID:
			return nil, errors.New("email already exists")
		case err != nil && err != repositories.ErrUserNotFound:
			return nil, fmt.Errorf("error checking existing email: %w", err)
		}
		user.Email = in.Email
	}
	if in.Name != "" && in.Name != user.Name {
		user.Name = in.Name
	}
	if in.Address != "" {
		user.Address = in.Address
	}
	if in.Phone != "" {
		user.Phone = in.Phone
	}
	if in.Description != "" {
		user.Description = in.Description
	}

	updated, err := s.UserRepository.Update(nil, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, errors.New("email already exists")
		}
		return nil, fmt.Errorf("error updating user: %w", err)
	}
	return updated, nil
}

func (s *UserService) UpdateAvatarOnly(userId string, ctx *fiber.Ctx) (*models.User, error) {
	user, err := s.UserRepository.FindById(nil, userId, true)
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

	// Minimal update kolom
	_ = time.Now() // keep behavior align with previous code
	user.AvatarID = &newAvatarID
	if _, err := s.UserRepository.Update(tx, user); err != nil {
		tx.Rollback()
		helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
		return nil, fmt.Errorf("error updating user avatar_id: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		helpers.DeleteLocalFileImmediate(newAvatarUUIDStr)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	updated, err := s.UserRepository.FindById(nil, userId, true)
	if err != nil {
		log.Printf("Warning: User updated but failed to fetch updated data: %v", err)
		return user, nil
	}

	if oldAvatarID != nil {
		go func(id string) {
			if err := helpers.DeleteLocalFileImmediate(id); err != nil {
				log.Printf("Warning: Failed to delete old avatar file %s: %v", id, err)
			}
		}(oldAvatarID.String())
	}
	return updated, nil
}

func (s *UserService) DeleteUsers(in *models.UserIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for user %v: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		u, err := s.UserRepository.FindById(tx, id.String(), true)
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrUserNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("User not found: %v\n", id)
				continue
			}
			log.Printf("Error finding user %v: %v\n", id, err)
			return errors.New("error finding user")
		}

		isHard := in.IsHardDelete == "hardDelete"
		if err := s.UserRepository.Delete(tx, id.String(), isHard); err != nil {
			tx.Rollback()
			log.Printf("Error deleting user %v: %v\n", id, err)
			return errors.New("error deleting user")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for user %v: %v\n", id, err)
			return errors.New("error committing delete")
		}

		// Hapus file avatar setelah commit jika hard delete
		if isHard && u.AvatarID != nil && u.AvatarID.String() != "" {
			go func(avatarID string) {
				if err := helpers.DeleteLocalFileImmediate(avatarID); err != nil {
					log.Printf("Warning: Failed to delete avatar file %s: %v", avatarID, err)
				}
			}(u.AvatarID.String())
		}
	}
	return nil
}

func (s *UserService) RestoreUsers(in *models.UserRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.User, error) {
	var restored []models.User
	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for user restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		res, err := s.UserRepository.Restore(tx, id.String())
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrUserNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("User not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error restoring user %v: %v\n", id, err)
			return nil, errors.New("error restoring user")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing user restore %v: %v\n", id, err)
			return nil, errors.New("error committing user restore")
		}

		restoredUser, ferr := s.UserRepository.FindById(nil, res.ID.String(), true)
		if ferr != nil {
			log.Printf("Error fetching restored user %v: %v\n", id, ferr)
			continue
		}
		restored = append(restored, *restoredUser)
	}
	return restored, nil
}