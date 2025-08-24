package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	UserRepository repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) *AuthService {
	return &AuthService{
		UserRepository: userRepo,
	}
}

// ==============================
// Reads (tanpa transaction)
// ==============================

func (s *AuthService) Login(in *models.UserLoginRequest, ctx *fiber.Ctx) (*models.User, string, error) {
	_ = ctx

	// read only
	user, err := s.UserRepository.FindByEmailOrUsername(nil, in.Identifier)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return nil, "", errors.New("invalid email or password")
		}
		return nil, "", err
	}

	if err := helpers.VerifyPassword(user.Password, in.Password); err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	token, err := helpers.CreateToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) CheckIdentifier(in *models.UserCheckIdentifierRequest, ctx *fiber.Ctx) (*models.User, error) {
	_ = ctx

	// read only
	user, err := s.UserRepository.FindByEmailOrUsername(nil, in.Identifier)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// ==============================
// Mutations (pakai transaction)
// ==============================

func (s *AuthService) ResetPassword(userID, currentPassword, newPassword string, ctx *fiber.Ctx) error {
	_ = ctx

	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	user, err := s.UserRepository.FindById(tx, userID, false)
	if err != nil {
		tx.Rollback()
		if err == repositories.ErrUserNotFound {
			return errors.New("user not found")
		}
		return err
	}

	if err := helpers.VerifyPassword(user.Password, currentPassword); err != nil {
		tx.Rollback()
		return errors.New("current password is incorrect")
	}

	hashed, err := helpers.HashPassword(newPassword)
	if err != nil {
		tx.Rollback()
		return err
	}
	user.Password = hashed

	if _, err := s.UserRepository.Update(tx, user); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *AuthService) ForgotPassword(in *models.UserForgotPasswordRequest, ctx *fiber.Ctx) (*models.User, error) {
	_ = ctx

	if strings.TrimSpace(in.Password) == "" {
		return nil, errors.New("new password cannot be empty")
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	user, err := s.UserRepository.FindByEmailOrUsername(tx, in.Identifier)
	if err != nil {
		tx.Rollback()
		if err == repositories.ErrUserNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	hashed, err := helpers.HashPassword(in.Password)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	user.Password = hashed

	updated, err := s.UserRepository.Update(tx, user)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return updated, nil
}
