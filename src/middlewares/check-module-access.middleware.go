package middlewares

import (
	"fmt"
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

	requestPath := ctx.Path()
	requestMethod := ctx.Method()
	fmt.Printf("Request Path: %s, Method: %s\n", requestPath, requestMethod)

	var modules []models.Module
	if err := configs.DB.Where("deleted_at IS NULL").Find(&modules).Error; err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error while checking permissions", nil)
	}

	var matchedModule *models.Module
	for i := range modules {
		if matchDynamicPath(requestPath, modules[i].Route) {
			matchedModule = &modules[i]
			fmt.Printf("Matched module: %s (ID: %s) with route: %s\n", modules[i].Name, modules[i].ID, modules[i].Route)
			break
		}
	}

	if matchedModule == nil {
		fmt.Println("No matching module found, allowing request to proceed")
		return ctx.Next()
	}

	var roleModule models.RoleModule
	err := configs.DB.Where("role_id = ? AND module_id = ? AND checked = ? AND deleted_at IS NULL",
		userInfo.RoleID, matchedModule.ID, true).First(&roleModule).Error
	
	if err != nil {
		fmt.Printf("Permission denied for role %s on module %s\n", role.Name, matchedModule.Name)
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have permission to access this resource", nil)
	}

	fmt.Printf("Permission granted for role %s on module %s\n", role.Name, matchedModule.Name)
	return ctx.Next()
}

func matchDynamicPath(requestPath, dbPath string) bool {
	requestPath = strings.Split(requestPath, "?")[0]
	requestPath = strings.Trim(requestPath, "/")
	dbPath = strings.Trim(dbPath, "/")
	
	if requestPath == "" && dbPath == "" {
		return true
	}
	
	requestParts := []string{}
	dbParts := []string{}
	
	if requestPath != "" {
		requestParts = strings.Split(requestPath, "/")
	}
	if dbPath != "" {
		dbParts = strings.Split(dbPath, "/")
	}
	
	if len(requestParts) != len(dbParts) {
		return false
	}
	
	for i := range dbParts {
		if strings.HasPrefix(dbParts[i], ":") {
			continue
		}
		if dbParts[i] != requestParts[i] {
			return false
		}
	}
	
	return true
}