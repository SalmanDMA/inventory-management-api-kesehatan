package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserControllerGetAll adalah handler untuk endpoint user dengan pagination
// @Summary Get all users with pagination
// @Description Get all users with pagination, filtering, and search.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param role_id query string false "Filter by role ID"
// @Success 200 {object} models.UserPaginatedResponse
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting users"
// @Router /api/v1/user [get]
func UserControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	userService := services.NewUserService(userRepo, uploadRepo)
	
	usersResponse, err := userService.GetAllUsersPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting users", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get all users", usersResponse)
}

// UserControllerCreate adalah handler untuk endpoint user
// @Summary Create user
// @Description Create user. Forn now only accessible by users with DEVELOPER or SUPERADMIN roles.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param user body models.UserCreate true "User"
// @Success 200 {string} string "Success creating user"
// @Failure 400 {string} string "Invalid request body"
// @Failure 403 {string} string "Forbidden: You do not have access to create users"
// @Failure 409 {string} string "Email already exists"
// @Failure 409 {string} string "Username already exists"
// @Failure 500 {string} string "Error creating user"
// @Router /api/v1/user [post]
func UserControllerCreate(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to create users", nil)
	}

	userRequest := new(models.UserCreate)
	if err := ctx.BodyParser(userRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(userRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	userService := services.NewUserService(userRepo,uploadRepo)

	newUser, err := userService.CreateUser(userRequest, ctx, userInfo)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			return helpers.Response(ctx, fiber.StatusConflict, err.Error(), nil)
		}
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error creating user: "+err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusCreated, "Success creating user", newUser)
}

// UserControllerGetById adalah handler untuk endpoint user
// @Summary Get user by ID
// @Description Get user by ID. Forn now only accessible by users with DEVELOPER or SUPERADMIN roles.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "User ID"
// @Success 200 {object} models.ResponseGetUser
// @Failure 403 {string} string "Forbidden: You do not have access to view user details"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error getting user"
// @Router /api/v1/user/{id} [get]
func UserControllerGetById(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
					return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to view user details", nil)
	}

	userId := ctx.Params("id")

	userRepo := repositories.NewUserRepository(configs.DB)
	user, err := userRepo.FindById(userId, false)
	if err != nil {
					if err == repositories.ErrUserNotFound {
									return helpers.Response(ctx, fiber.StatusNotFound, "User not found", nil)
					}
					return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting user: "+err.Error(), nil)
	}

	var roleId uuid.UUID
	if user.RoleID != nil {
		roleId = *user.RoleID
	}

	userResponse := &models.ResponseGetUser{
					ID:          user.ID,
					Username:    user.Username,
					Name:        user.Name,
					Email:       user.Email,
					Address:     user.Address,
					Phone:       user.Phone,
					Avatar:      user.Avatar,
					AvatarID:    user.AvatarID,
					Role:        user.Role,
					RoleID:      roleId,
					Description: user.Description,
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success get user", userResponse)
}

// UserControllerGetProfile adalah handler untuk endpoint user
// @Summary Get user profile
// @Description Get user profile. For now only accessible by users with DEVELOPER or SUPERADMIN roles.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetUser
// @Failure 404 {string} string "User not found"
// @Router /api/v1/user/me [get]
func UserControllerGetProfile(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "User info not found", nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	user, err := userRepo.FindById(userInfo.ID.String(), false)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return helpers.Response(ctx, fiber.StatusNotFound, "User not found", nil)
		}
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error getting user: "+err.Error(), nil)
	}

	userResponse := &models.ResponseGerUserProfile{
		ID:                user.ID,
		Username:          user.Username,
		Name:              user.Name,
		Email:             user.Email,
		Address:           user.Address,
		Phone:             user.Phone,
		AvatarID:          user.AvatarID,
		Avatar:            user.Avatar,
		Role:              user.Role,
		Description:       user.Description,
		RoleID:            user.RoleID,
	}


	return helpers.Response(ctx, fiber.StatusOK, "Success get user", userResponse)
}

// UserControllerUpdateProfile adalah handler untuk endpoint user update profile
// @Summary Update user profile
// @Description Update user profile. For now only accessible by users with DEVELOPER or SUPERADMIN roles.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.UserUpdateProfileRequest true "User update request body"
// @Success 200 {string} string "Success update user profile"
// @Failure 403 {string} string "Forbidden: You do not have access to update user profile"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error updating user profile"
// @Router /api/v1/user/me [put]
func UserControllerUpdateProfile(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

	userUpdate := new(models.UserUpdateProfileRequest)
	if err := ctx.BodyParser(userUpdate); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	userService := services.NewUserService(userRepo,uploadRepo)

	updatedUser, err := userService.UpdateUserProfile(userInfo.ID.String(), userUpdate)
	if err != nil {
		if err.Error() == "user not found" {
			return helpers.Response(ctx, fiber.StatusNotFound, err.Error(), nil)
		}
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success update user profile", updatedUser)
}

// @Summary Upload avatar
// @Description Update user avatar (multipart/form-data, field name: avatar)
// @Tags User
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param avatar formData file true "Avatar image (jpg/jpeg/png, <=2MB)"
// @Success 200 {object} map[string]interface{} "data: { user, avatar_url }"
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 500 {string} string
// @Router /api/v1/user/me/avatar [post]
func UserControllerUploadAvatar(ctx *fiber.Ctx) error {
  userInfo, ok := ctx.Locals("userInfo").(*models.User)
  if !ok {
    return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
  }

  userRepo := repositories.NewUserRepository(configs.DB)
  uploadRepo := repositories.NewUploadRepository(configs.DB)
  userService := services.NewUserService(userRepo, uploadRepo)

  updatedUser, svcErr := userService.UpdateAvatarOnly(userInfo.ID.String(), ctx)
  if svcErr != nil {
    return helpers.Response(ctx, fiber.StatusInternalServerError, svcErr.Error(), nil)
  }

  return helpers.Response(ctx, fiber.StatusOK, "Avatar updated", updatedUser)
}



// UserControllerUpdate adalah handler untuk endpoint user
// @Summary Update user
// @Description Update user. For now only accessible by users with DEVELOPER or SUPERADMIN roles.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "User ID"
// @Param request body models.UserUpdate true "User update request body"
// @Success 200 {string} string "Success update user"
// @Failure 403 {string} string "Forbidden: You do not have access to update user"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error updating user"
// @Router /api/v1/user/{id} [put]
func UserControllerUpdate(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
					return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to update user", nil)
	}

	userRequest := new(models.UserUpdate)
	if err := ctx.BodyParser(userRequest); err != nil {
					return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(userRequest); err != nil {
					errorMessage := helpers.ExtractErrorMessages(err)
					return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	userService := services.NewUserService(userRepo,uploadRepo)
	updatedUser, err := userService.UpdateUser(userRequest, ctx.Params("id"), ctx, userInfo)
	
	if err != nil {
					if err.Error() == "email already exists" {
									return helpers.Response(ctx, fiber.StatusConflict, "Email already exists", nil)
					}
					return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating user: "+err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Success updating user", updatedUser)
}

// UserControllerDelete adalah handler untuk endpoint user
// @Summary Delete user
// @Description Delete user. For now only accessible by users with DEVELOPER or SUPERADMIN roles. If the user is hard deleted, the user's avatar will be deleted as well. If the user is soft deleted, the user's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.UserIsHardDeleteRequest true "User delete request body"
// @Success 200 {array} models.User
// @Failure 403 {string} string "Forbidden: You do not have access to delete user"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error deleting user"
// @Router /api/v1/user/delete [delete]
func UserControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to delete users", nil)
	}

	userRequest := new(models.UserIsHardDeleteRequest)
	if err := ctx.BodyParser(userRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(userRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	userService := services.NewUserService(userRepo,uploadRepo)

	if err := userService.DeleteUsers(userRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if userRequest.IsHardDelete == "hardDelete"  {
		message = "Users deleted successfully"
	} else {
		message = "Users moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// UserControllerRestore restores a soft-deleted user
// @Summary Restore user
// @Description Restore user. For now only accessible by users with DEVELOPER or SUPERADMIN roles.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.UserRestoreRequest true "User restore request body"
// @Success 200 {array} models.User
// @Failure 403 {string} string "Forbidden: You do not have access to restore user"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error restoring user"
// @Router /api/v1/user/restore [put]
func UserControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to restore users", nil)
	}

	userRequest := new(models.UserRestoreRequest)
	if err := ctx.BodyParser(userRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(userRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	userRepo := repositories.NewUserRepository(configs.DB)
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	userService := services.NewUserService(userRepo,uploadRepo)	

	restoredUsers, err := userService.RestoreUsers(userRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Users restored successfully", restoredUsers)
}
