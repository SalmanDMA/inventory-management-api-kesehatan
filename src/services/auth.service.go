package services

import (
	"errors"

	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepository repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) *AuthService {
	return &AuthService{
		UserRepository: userRepo,
	}
}

func (service *AuthService) Login(userRequest *models.UserLoginRequest, ctx *fiber.Ctx) (*models.User, string, error) {
	user, err := service.UserRepository.FindByEmailOrUsername(userRequest.Identifier)
	if err != nil {
		return nil, "", err
	}

	if err := helpers.VerifyPassword(user.Password, userRequest.Password); err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	token, err := helpers.CreateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (service *AuthService) ResetPassword(userID string, currentPassword string, newPassword string, ctx *fiber.Ctx) error {
	user, err := service.UserRepository.FindById(userID, false)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return errors.New("user not found")
		}
		return err
	}

	if err := helpers.VerifyPassword(user.Password, currentPassword); err != nil {
		return errors.New("current password is incorrect")
	}

	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}

	passwordHash, err := helpers.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = passwordHash

	_, err = service.UserRepository.Update(user)
	if err != nil {
		return err
	}

	return nil
}

func (service *AuthService) CheckIdentifier(req *models.UserCheckIdentifierRequest, ctx *fiber.Ctx) (*models.User, error) {
	user, err := service.UserRepository.FindByEmailOrUsername(req.Identifier)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (service *AuthService) ForgotPassword(req *models.UserForgotPasswordRequest, ctx *fiber.Ctx) (*models.User, error) {
	user, err := service.UserRepository.FindByEmailOrUsername(req.Identifier)
	if err != nil {
					return nil, errors.New("user not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
					return nil, err
	}

	user.Password = string(hashedPassword)

	_, err = service.UserRepository.Update(user)
	if err != nil {
					return nil, err
	}

	return user, nil
}
