package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all roles with pagination
// @Description Get all roles with pagination, filtering, and search.
// @Tags Role
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param role_id query string false "Filter by role ID"
// @Success 200 {array} models.ResponseGetRole
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving roles"
// @Router /api/v1/role [get]
func RoleControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	roleRepo := repositories.NewRoleRepository(configs.DB)
	roleService := services.NewRoleService(roleRepo)
	
	rolesResponse, err := roleService.GetAllRolesPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting roles", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get all roles", rolesResponse)
}

// @Summary Get By ID role
// @Description Get By ID role
// @Tags roles
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetRole
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting role"
// @Router /api/v1/role/{id} [get]
func RoleControllerGetByID(c *fiber.Ctx) error {
	userInfo, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Role info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(c, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	id := c.Params("id")
	roleRepo := repositories.NewRoleRepository(configs.DB)
	roleService := services.NewRoleService(roleRepo)
	roleResponse, err := roleService.GetRoleByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting role", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Role fetched successfully", roleResponse)
}

	// @Summary Create role
	// @Description Create role. For now only accessible by users with DEVELOPER or SUPERADMIN roles.
	// @Tags Role
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param roleCreateRequest body models.RoleCreateRequest true "Role create request"
	// @Success 200 {object} models.ResponseGetRole
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create roles"
	// @Failure 409 {string} string "Role already exists"
	// @Failure 500 {string} string "Error creating role"
	// @Router /api/v1/role [post]
	func RoleControllerCreate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
			return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to create roles", nil)
		}
	
		role := new(models.RoleCreateRequest)
	
		if err := ctx.BodyParser(role); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		if err := helpers.ValidateStruct(role); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		roleRepo := repositories.NewRoleRepository(configs.DB)
		roleService := services.NewRoleService(roleRepo)
		
		if _ , err := roleService.CreateRole(role, ctx, userInfo); err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Role created successfully", nil)
	}

	// @Summary Update role
	// @Description Update role. For now only accessible by users with DEVELOPER or SUPERADMIN roles.
	// @Tags Role
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param id path string true "Role ID"
	// @Param roleUpdateRequest body models.RoleUpdateRequest true "Role update request"
	// @Success 200 {object} models.ResponseGetRole
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to update roles"
	// @Failure 404 {string} string "Role not found"
	// @Failure 500 {string} string "Error updating role"
	// @Router /api/v1/role/{id} [put]
	func RoleControllerUpdate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
			return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to update roles", nil)
		}
	
		role := new(models.RoleUpdateRequest)
	
		if err := ctx.BodyParser(role); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		roleId := ctx.Params("id")
	
		if err := helpers.ValidateStruct(role); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		roleRepo := repositories.NewRoleRepository(configs.DB)
		roleService := services.NewRoleService(roleRepo)
		if _ , err := roleService.UpdateRole(roleId, role, ctx, userInfo); err != nil {
			if err == repositories.ErrRoleNotFound {
				return helpers.Response(ctx, fiber.StatusNotFound, "Role not found", nil)
			}
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating role", nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Role updated successfully", nil)
	}

// RoleControllerDelete adalah handler untuk endpoint role
// @Summary Delete role
// @Description Delete role. For now only accessible by roles with DEVELOPER or SUPERADMIN roles. If the role is hard deleted, the role's avatar will be deleted as well. If the role is soft deleted, the role's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags Role
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.RoleIsHardDeleteRequest true "Role delete request body"
// @Success 200 {array} models.Role
// @Failure 403 {string} string "Forbidden: You do not have access to delete role"
// @Failure 404 {string} string "Role not found"
// @Failure 500 {string} string "Error deleting role"
// @Router /api/v1/role/delete [delete]
func RoleControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Role info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to delete roles", nil)
	}

	roleRequest := new(models.RoleIsHardDeleteRequest)
	if err := ctx.BodyParser(roleRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(roleRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	roleRepo := repositories.NewRoleRepository(configs.DB)
	roleService := services.NewRoleService(roleRepo)

	if err := roleService.DeleteRoles(roleRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if roleRequest.IsHardDelete == "hardDelete"  {
		message = "Roles deleted successfully"
	} else {
		message = "Roles moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// RoleControllerRestore restores a soft-deleted role
// @Summary Restore role
// @Description Restore role. For now only accessible by roles with DEVELOPER or SUPERADMIN roles.
// @Tags Role
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.RoleRestoreRequest true "Role restore request body"
// @Success 200 {array} models.Role
// @Failure 403 {string} string "Forbidden: You do not have access to restore role"
// @Failure 404 {string} string "Role not found"
// @Failure 500 {string} string "Error restoring role"
// @Router /api/v1/role/restore [put]
func RoleControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Role info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to restore roles", nil)
	}

	roleRequest := new(models.RoleRestoreRequest)
	if err := ctx.BodyParser(roleRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(roleRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	roleRepo := repositories.NewRoleRepository(configs.DB)
	roleService := services.NewRoleService(roleRepo)

	restoredRoles, err := roleService.RestoreRoles(roleRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Roles restored successfully", restoredRoles)
}