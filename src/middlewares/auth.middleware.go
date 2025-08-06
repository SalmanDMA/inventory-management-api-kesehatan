package middlewares

import (
	"log"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
)

func JWTProtected(ctx *fiber.Ctx) error {
	authHeader := ctx.Get("Authorization")

	if authHeader == "" {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Missing Authorization header", nil)
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Invalid Authorization format", nil)
	}

	tokenStr := parts[1]

	claims, err := helpers.ValidateToken(tokenStr)
	if err != nil {
		log.Printf("Invalid token: %v", err)
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Invalid token", nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	user, err := userRepo.FindByEmailOrUsername(claims.Email)
	if err != nil {
		log.Printf("User fetch error: %v", err)
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Internal server error", nil)
	}

	if user == nil {
		return helpers.Response(ctx, fiber.StatusNotFound, "User not found", nil)
	}

	ctx.Locals("userInfo", user)

	return ctx.Next()
}
