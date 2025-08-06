package middlewares

import (
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/gofiber/fiber/v2"
)

func RBACMiddleware(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	var role models.Role
	if err := configs.DB.First(&role, userInfo.RoleID).Error; err != nil {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Role not found", nil)
	}

	if strings.ToLower(role.Name) == "developer" {
		return ctx.Next()
	}

	path := ctx.Path()

	var module models.Module
	if err := configs.DB.Where("path LIKE ? AND deleted_at IS NULL", "%"+path+"%").First(&module).Error; err != nil {
		return helpers.Response(ctx, fiber.StatusNotFound, "Resource not found", nil)
	}

	var roleModule models.RoleModule
	if err := configs.DB.Where("role_id = ? AND module_id = ? AND checked = ? AND deleted_at IS NULL",
		userInfo.RoleID, module.ID, true).First(&roleModule).Error; err != nil {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have permission to access this resource", nil)
	}

	return ctx.Next()
}
