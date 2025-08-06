package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all items
// @Description Get all items with pagination, filtering, and search.
// @Tags Item
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param category_id query string false "Filter by category ID"
// @Success 200 {array} models.ResponseGetItem
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving items"
// @Router /api/v1/item [get]
func ItemControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to view items", nil)
	}

	paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	uploadRepo := repositories.NewUploadRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	itemService := services.NewItemService(uploadRepo, itemRepo, itemHistoryRepo)

	itemsResponse, err := itemService.GetAllItemsPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error retrieving items", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Items retrieved successfully", itemsResponse)
}


// @Summary Get By ID item
// @Description Get By ID item
// @Tags items
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetItem
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting item"
// @Router /api/v1/item/{id} [get]
func ItemControllerGetByID(c *fiber.Ctx) error {
	userInfo, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(c, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	id := c.Params("id")
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	itemService := services.NewItemService(uploadRepo, itemRepo, itemHistoryRepo)
	itemResponse, err := itemService.GetItemByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting item", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Item fetched successfully", itemResponse)
}

	// @Summary Create item
	// @Description Create item. For now only accessible by users with DEVELOPER or SUPERADMIN items.
	// @Tags Item
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param itemCreateRequest body models.ItemCreateRequest true "Item create request"
	// @Success 200 {object} models.ResponseGetItem
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create items"
	// @Failure 409 {string} string "Item already exists"
	// @Failure 500 {string} string "Error creating item"
	// @Router /api/v1/item [post]
	func ItemControllerCreate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
			return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

		if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
			return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to create items", nil)
		}

		itemRequest := new(models.ItemCreateRequest)
		if err := ctx.BodyParser(itemRequest); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}

		if err := helpers.ValidateStruct(itemRequest); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}

		uploadRepo := repositories.NewUploadRepository(configs.DB)
		itemRepo := repositories.NewItemRepository(configs.DB)
		itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
		itemService := services.NewItemService(uploadRepo, itemRepo, itemHistoryRepo)

		if _, err := itemService.CreateItem(itemRequest, ctx, userInfo); err != nil {
			return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
		}

		return helpers.Response(ctx, fiber.StatusOK, "Item created successfully", nil)
	}


	// @Summary Update item
	// @Description Update item. For now only accessible by users with DEVELOPER or SUPERADMIN items.
	// @Tags Item
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param id path string true "Item ID"
	// @Param itemUpdateRequest body models.ItemUpdateRequest true "Item update request"
	// @Success 200 {object} models.ResponseGetItem
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to update items"
	// @Failure 404 {string} string "Item not found"
	// @Failure 500 {string} string "Error updating item"
	// @Router /api/v1/item/{id} [put]
	func ItemControllerUpdate(ctx *fiber.Ctx) error {
		userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
			return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
		}

		if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
			return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to update items", nil)
		}

		itemRequest := new(models.ItemUpdateRequest)
		if err := ctx.BodyParser(itemRequest); err != nil {
			return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
		}
		
		if err := helpers.ValidateStruct(itemRequest); err != nil {
			errorMessage := helpers.ExtractErrorMessages(err)
			return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
		}

		itemID := ctx.Params("id")

		uploadRepo := repositories.NewUploadRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	itemService := services.NewItemService(uploadRepo, itemRepo, itemHistoryRepo)

		if _, err := itemService.UpdateItem(itemID, itemRequest, ctx, userInfo); err != nil {
			if err == repositories.ErrItemNotFound {
				return helpers.Response(ctx, fiber.StatusNotFound, "Item not found", nil)
			}
			return helpers.Response(ctx, fiber.StatusInternalServerError, "Error updating item", nil)
		}

		return helpers.Response(ctx, fiber.StatusOK, "Item updated successfully", nil)
	}


// ItemControllerDelete adalah handler untuk endpoint item
// @Summary Delete item
// @Description Delete item. For now only accessible by items with DEVELOPER or SUPERADMIN items. If the item is hard deleted, the item's avatar will be deleted as well. If the item is soft deleted, the item's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags Item
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.ItemIsHardDeleteRequest true "Item delete request body"
// @Success 200 {array} models.Item
// @Failure 403 {string} string "Forbidden: You do not have access to delete item"
// @Failure 404 {string} string "Item not found"
// @Failure 500 {string} string "Error deleting item"
// @Router /api/v1/item/delete [delete]
func ItemControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to delete items", nil)
	}

	itemRequest := new(models.ItemIsHardDeleteRequest)
	if err := ctx.BodyParser(itemRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(itemRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	uploadRepo := repositories.NewUploadRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	itemService := services.NewItemService(uploadRepo, itemRepo, itemHistoryRepo)

	if err := itemService.DeleteItems(itemRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if itemRequest.IsHardDelete == "hardDelete"  {
		message = "Items deleted successfully"
	} else {
		message = "Items moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// ItemControllerRestore restores a soft-deleted item
// @Summary Restore item
// @Description Restore item. For now only accessible by items with DEVELOPER or SUPERADMIN items.
// @Tags Item
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.ItemRestoreRequest true "Item restore request body"
// @Success 200 {array} models.Item
// @Failure 403 {string} string "Forbidden: You do not have access to restore item"
// @Failure 404 {string} string "Item not found"
// @Failure 500 {string} string "Error restoring item"
// @Router /api/v1/item/restore [put]
func ItemControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Item info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to restore items", nil)
	}

	itemRequest := new(models.ItemRestoreRequest)
	if err := ctx.BodyParser(itemRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(itemRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	uploadRepo := repositories.NewUploadRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	itemService := services.NewItemService(uploadRepo, itemRepo, itemHistoryRepo)

	restoredItems, err := itemService.RestoreItems(itemRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Items restored successfully", restoredItems)
}