package controllers

import (
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all item histories
// @Description Get all item histories with pagination, filtering, and search.
// @Tags Item History
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Item Histories per page (default: 10, max: 100)"
// @Param search query string false "Search term for name, email, or username"
// @Param status query string false "Filter by status: active, deleted, all (default: active)"
// @Param category_id query string false "Filter by category ID"
// @Success 200 {array} models.ResponseGetItemHistory
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving item histories"
// @Router /api/v1/item-history [get]
func ItemHistoryControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to view itemHistories", nil)
	}

	paginationReq := &models.PaginationRequest{}
	if err := ctx.QueryParser(paginationReq); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	if err := helpers.ValidateStruct(paginationReq); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	itemHistoryService := services.NewItemHistoryService(itemHistoryRepo, itemRepo)

	itemHistoriesResponse, err := itemHistoryService.GetAllItemHistoriesPaginated(paginationReq, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error retrieving itemHistories", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Item Histories retrieved successfully", itemHistoriesResponse)
}

	// @Summary Create item history
	// @Description Create item history. For now only accessible by users with DEVELOPER or SUPERADMIN itemHistories.
	// @Tags Item history
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Param Authorization header string true "Authorization"
	// @Param itemHistoryCreateRequest body models.ItemHistoryCreateRequest true "Item history create request"
	// @Success 200 {object} models.ResponseGetItemHistory
	// @Failure 400 {string} string "Invalid request body"
	// @Failure 403 {string} string "Forbidden: You do not have access to create itemHistories"
	// @Failure 409 {string} string "Item history already exists"
	// @Failure 500 {string} string "Error creating item history"
	// @Router /api/v1/item-history [post]
	func ItemHistoryControllerCreate(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: User info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to create item changes", nil)
	}

	itemHistoryRequest := new(models.ItemHistoryCreateRequest)
	if err := ctx.BodyParser(itemHistoryRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(itemHistoryRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	itemHistoryService := services.NewItemHistoryService(itemHistoryRepo, itemRepo)

	_, err := itemHistoryService.CreateItemHistory(itemHistoryRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var action string
	switch strings.ToLower(itemHistoryRequest.ChangeType) {
	case "create_price":
		action = "Item price created successfully"
	case "update_price":
		action = "Item price updated successfully"
	case "create_stock":
		action = "Item stock created successfully"
	case "update_stock":
		action = "Item stock updated successfully"
	default:
		action = "Item updated successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, action, nil)
}

// ItemHistoryControllerDelete adalah handler untuk endpoint itemHistory
// @Summary Delete item history
// @Description Delete item history. For now only accessible by itemHistories with DEVELOPER or SUPERADMIN itemHistories. If the itemHistory is hard deleted, the itemHistory's avatar will be deleted as well. If the itemHistory is soft deleted, the itemHistory's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags Item History
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.ItemHistoryIsHardDeleteRequest true "Item History delete request body"
// @Success 200 {array} models.ItemHistory
// @Failure 403 {string} string "Forbidden: You do not have access to delete itemHistory"
// @Failure 404 {string} string "Item History not found"
// @Failure 500 {string} string "Error deleting item history"
// @Router /api/v1/item-history/delete [delete]
func ItemHistoryControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: ItemHistory info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to delete itemHistories", nil)
	}

	itemHistoryRequest := new(models.ItemHistoryIsHardDeleteRequest)
	if err := ctx.BodyParser(itemHistoryRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(itemHistoryRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	itemHistoryService := services.NewItemHistoryService(itemHistoryRepo, itemRepo)

	if err := itemHistoryService.DeleteItemHistories(itemHistoryRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if itemHistoryRequest.IsHardDelete == "hardDelete"  {
		message = "Item Histories deleted successfully"
	} else {
		message = "Item Histories moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// ItemHistoryControllerRestore restores a soft-deleted item history
// @Summary Restore item history
// @Description Restore item history. For now only accessible by item histories with DEVELOPER or SUPERADMIN item histories.
// @Tags ItemHistory
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.ItemHistoryRestoreRequest true "Item History restore request body"
// @Success 200 {array} models.ItemHistory
// @Failure 403 {string} string "Forbidden: You do not have access to restore item history"
// @Failure 404 {string} string "Item History not found"
// @Failure 500 {string} string "Error restoring item history"
// @Router /api/v1/item-history/restore [put]
func ItemHistoryControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: ItemHistory info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to restore itemHistories", nil)
	}

	itemHistoryRequest := new(models.ItemHistoryRestoreRequest)
	if err := ctx.BodyParser(itemHistoryRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(itemHistoryRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	itemHistoryRepo := repositories.NewItemHistoryRepository(configs.DB)
	itemRepo := repositories.NewItemRepository(configs.DB)
	itemHistoryService := services.NewItemHistoryService(itemHistoryRepo, itemRepo)

	restoredItemHistories, err := itemHistoryService.RestoreItemHistories(itemHistoryRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Item Histories restored successfully", restoredItemHistories)
}