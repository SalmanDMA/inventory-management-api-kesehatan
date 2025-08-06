package controllers

import (
	"errors"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// LoginController adalah handler untuk endpoint login.
// @Summary Login user
// @Description Login user dengan email dan password.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.UserLoginRequest true "Login request body"
// @Success 200 {string} string "Success login"
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Router /api/v1/auth/login [post]
func LoginController(ctx *fiber.Ctx) error {
	loginRequest := new(models.UserLoginRequest)

	if err := ctx.BodyParser(loginRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(loginRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	authService := services.NewAuthService(userRepo)

	user , token, err := authService.Login(loginRequest, ctx)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Invalid email or password", nil)
	}


	return helpers.Response(ctx, fiber.StatusOK, "Success login", map[string]interface{}{
		"user": user,
		"token": token,
	})
}


// ResetPasswordController adalah handler untuk endpoint reset password.
// @Summary Reset password
// @Description Reset password pengguna yang sudah login.
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UserResetPasswordRequest true "Reset password request body"
// @Success 200 {string} string "Success reset password"
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Router /api/v1/auth/reset-password [post]
func ResetPasswordController(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	userRequest := new(models.UserResetPasswordRequest)
	println(userRequest, "userRequest")
	if err := ctx.BodyParser(userRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(userRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	authService := services.NewAuthService(userRepo)
	
	err := authService.ResetPassword(userInfo.ID.String(), userRequest.CurrentPassword, userRequest.NewPassword, ctx)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if errors.Is(err, repositories.ErrUserNotFound) {
			statusCode = fiber.StatusNotFound
		} else if errors.Is(err, errors.New("new password cannot be empty")) {
			statusCode = fiber.StatusBadRequest
		}
		return helpers.Response(ctx, statusCode, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success reset password", nil)
}

// CheckIdentityController adalah handler untuk check identifier
// @Summary Check identifier
// @Description Menangani permintaan check identifier oleh pengguna.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.UserCheckIdentifierRequest true "Check identifier request body"
// @Success 200 {string} string "Success check identifier"
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/auth/check-identifier [post]
func CheckIdentifierController(ctx *fiber.Ctx) error {
	req := new(models.UserCheckIdentifierRequest)

	if err := ctx.BodyParser(req); err != nil {
					return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	authService := services.NewAuthService(userRepo)

	user, err := authService.CheckIdentifier(req, ctx)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if errors.Is(err, repositories.ErrUserNotFound) {
			statusCode = fiber.StatusNotFound
		}
		return helpers.Response(ctx, statusCode, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success check identifier", user)
}

// ForgotPasswordController adalah handler untuk lupa password.
// @Summary Forgot password
// @Description Menangani permintaan lupa password oleh pengguna.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.UserForgotPasswordRequest true "Forgot password request body"
// @Success 200 {string} string "Success reset your password"
// @Failure 400 {string} string "Bad request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/auth/forgot-password [post]
func ForgotPasswordController(ctx *fiber.Ctx) error {
	req := new(models.UserForgotPasswordRequest)

	if err := ctx.BodyParser(req); err != nil {
					return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	authService := services.NewAuthService(userRepo)

	user, err := authService.ForgotPassword(req, ctx)
	if err != nil {
					if err.Error() == "User not found" {
									return helpers.Response(ctx, fiber.StatusNotFound, err.Error(), nil)
					}
					if err.Error() == "Password reset is not allowed, please verify your OTP" {
									return helpers.Response(ctx, fiber.StatusUnauthorized, err.Error(), nil)
					}
					return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success reset your password", user)
}



