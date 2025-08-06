package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all categories
// @Description Get all categories with pagination, filtering, and search.
// @Tags Category
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Success 200 {array} models.ResponseGetCategory
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving categories"
// @Router /api/v1/category [get]
func CategoryControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to view categories", nil)
	}

		paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	categoryRepo := repositories.NewCategoryRepository(configs.DB)
	categoryService := services.NewCategoryService(categoryRepo)
	categoriesResponse, err := categoryService.GetAllCategoriesPaginated(paginationReq, userInfo)

	println("Categories retrieved successfully:", categoriesResponse)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error retrieving categories", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Categories retrieved successfully", categoriesResponse)
}

// @Summary Get By ID category
// @Description Get By ID category
// @Tags categories
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetCategory
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting category"
// @Router /api/v1/category/{id} [get]
func CategoryControllerGetByID(c *fiber.Ctx) error {
	userInfo, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Category info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(c, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	id := c.Params("id")
	categoryRepo := repositories.NewCategoryRepository(configs.DB)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryResponse, err := categoryService.GetCategoryByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting category", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Category fetched successfully", categoryResponse)
}

	// @Summary Create category
	// @Description Create category. For now only accessible by users with DEVELOPER or SUPERADMIN categories.
	// @Tags Category
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param categoryCreateRequest body models.CategoryCreateRequest true "Category create request"
	// @Success 200 {object} models.ResponseGetCategory
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create categories"
	// @Failure 409 {string} string "Category already exists"
	// @Failure 500 {string} string "Error creating category"
	// @Router /api/v1/category [post]
	func CategoryControllerCreate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
			return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to create categories", nil)
		}
	
		category := new(models.CategoryCreateRequest)
	
		if err := ctx.BodyParser(category); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		if err := helpers.ValidateStruct(category); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		categoryRepo := repositories.NewCategoryRepository(configs.DB)
		categoryService := services.NewCategoryService(categoryRepo)
		
		if _ , err := categoryService.CreateCategory(category, ctx, userInfo); err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Category created successfully", nil)
	}

	// @Summary Update category
	// @Description Update category. For now only accessible by users with DEVELOPER or SUPERADMIN categories.
	// @Tags Category
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param id path string true "Category ID"
	// @Param categoryUpdateRequest body models.CategoryUpdateRequest true "Category update request"
	// @Success 200 {object} models.ResponseGetCategory
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to update categories"
	// @Failure 404 {string} string "Category not found"
	// @Failure 500 {string} string "Error updating category"
	// @Router /api/v1/category/{id} [put]
	func CategoryControllerUpdate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}
	
		if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
			return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to update categories", nil)
		}
	
		category := new(models.CategoryUpdateRequest)
	
		if err := ctx.BodyParser(category); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
	
		categoryId := ctx.Params("id")
	
		if err := helpers.ValidateStruct(category); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}
	
		categoryRepo := repositories.NewCategoryRepository(configs.DB)
		categoryService := services.NewCategoryService(categoryRepo)
		if _ , err := categoryService.UpdateCategory(categoryId, category, ctx, userInfo); err != nil {
			if err == repositories.ErrCategoryNotFound {
				return helpers.Response(ctx, fiber.StatusNotFound, "Category not found", nil)
			}
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating category", nil)
		}
	
		return helpers.Response(ctx, fiber.StatusOK, "Category updated successfully", nil)
	}

// CategoryControllerDelete adalah handler untuk endpoint category
// @Summary Delete category
// @Description Delete category. For now only accessible by categories with DEVELOPER or SUPERADMIN categories. If the category is hard deleted, the category's avatar will be deleted as well. If the category is soft deleted, the category's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags Category
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.CategoryIsHardDeleteRequest true "Category delete request body"
// @Success 200 {array} models.Category
// @Failure 403 {string} string "Forbidden: You do not have access to delete category"
// @Failure 404 {string} string "Category not found"
// @Failure 500 {string} string "Error deleting category"
// @Router /api/v1/category/delete [delete]
func CategoryControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Category info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to delete categories", nil)
	}

	categoryRequest := new(models.CategoryIsHardDeleteRequest)
	if err := ctx.BodyParser(categoryRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(categoryRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	categoryRepo := repositories.NewCategoryRepository(configs.DB)
	categoryService := services.NewCategoryService(categoryRepo)

	if err := categoryService.DeleteCategories(categoryRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if categoryRequest.IsHardDelete == "hardDelete"  {
		message = "Categories deleted successfully"
	} else {
		message = "Categories moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// CategoryControllerRestore restores a soft-deleted category
// @Summary Restore category
// @Description Restore category. For now only accessible by categories with DEVELOPER or SUPERADMIN categories.
// @Tags Category
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.CategoryRestoreRequest true "Category restore request body"
// @Success 200 {array} models.Category
// @Failure 403 {string} string "Forbidden: You do not have access to restore category"
// @Failure 404 {string} string "Category not found"
// @Failure 500 {string} string "Error restoring category"
// @Router /api/v1/category/restore [put]
func CategoryControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Category info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to restore categories", nil)
	}

	categoryRequest := new(models.CategoryRestoreRequest)
	if err := ctx.BodyParser(categoryRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(categoryRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	categoryRepo := repositories.NewCategoryRepository(configs.DB)
	categoryService := services.NewCategoryService(categoryRepo)

	restoredCategories, err := categoryService.RestoreCategories(categoryRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Categories restored successfully", restoredCategories)
}