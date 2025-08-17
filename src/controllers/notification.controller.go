package controllers

import (
	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/SalmanDMA/inventory-app/backend/src/services"
	"github.com/gofiber/fiber/v2"
)

// @Summary Get all notifications
// @Description Get all notifications. For now only accessible by users with DEVELOPER or SUPERADMIN notifications.
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {array} models.ResponseGetNotification
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error retrieving notifications"
// @Router /api/v1/notification [get]
func NotificationControllerGetAll(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to view notifications", nil)
	}

	notificationRepo := repositories.NewNotificationRepository(configs.DB)
	notificationService := services.NewNotificationService(notificationRepo)
	notificationsResponse, err := notificationService.GetAllNotifications(userInfo)

	println("Notifications retrieved successfully:", notificationsResponse)

	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error retrieving notifications", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Notifications retrieved successfully", notificationsResponse)
}

// @Summary Get By ID notification
// @Description Get By ID notification
// @Tags notifications
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {object} models.ResponseGetNotification
// @Failure 403 {string} string "Forbidden: You do not have access to this resource"
// @Failure 500 {string} string "Error getting notification"
// @Router /api/v1/notification/{id} [get]
func NotificationControllerGetByID(c *fiber.Ctx) error {
	userInfo, ok := c.Locals("userInfo").(*models.User)
	if !ok {
					return helpers.Response(c, fiber.StatusUnauthorized, "Unauthorized: Notification info not found", nil)
	}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(c, fiber.StatusForbidden, "Forbidden: You do not have access to this resource", nil)
	}

	id := c.Params("id")
	notificationRepo := repositories.NewNotificationRepository(configs.DB)
	notificationService := services.NewNotificationService(notificationRepo)
	notificationResponse, err := notificationService.GetNotificationByID(id)
	
	if err != nil {
		return helpers.Response(c, fiber.StatusInternalServerError, "Error getting notification", nil)
	}
	
	return helpers.Response(c, fiber.StatusOK, "Notification fetched successfully", notificationResponse)
}

// @Summary Mark all notifications as read
// @Description Mark all notifications as read for current user
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Success 200 {string} string "All notifications marked as read"
// @Failure 500 {string} string "Error marking notifications as read"
// @Router /api/v1/notification/mark-all-read [put]
func NotificationControllerMarkAllAsRead(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	notificationRepo := repositories.NewNotificationRepository(configs.DB)
	notificationService := services.NewNotificationService(notificationRepo)
	
	err := notificationService.MarkAllAsRead(userInfo.ID, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error marking notifications as read", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "All notifications marked as read successfully", nil)
}

// @Summary Mark multiple notifications as read
// @Description Mark multiple notifications as read by IDs
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.NotificationMarkMultipleRequest true "Notification IDs to mark as read"
// @Success 200 {string} string "Notifications marked as read"
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Error marking notifications as read"
// @Router /api/v1/notification/mark-multiple-read [put]
func NotificationControllerMarkMultipleAsRead(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	notificationRequest := new(models.NotificationMarkMultipleRequest)
	if err := ctx.BodyParser(notificationRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(notificationRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	notificationRepo := repositories.NewNotificationRepository(configs.DB)
	notificationService := services.NewNotificationService(notificationRepo)
	
	err := notificationService.MarkMultipleAsRead(notificationRequest.IDs, userInfo.ID, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error marking notifications as read", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Notifications marked as read successfully", nil)
}

// @Summary Mark notification as read
// @Description Mark a specific notification as read by ID
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param id path string true "Notification ID"
// @Success 200 {object} models.ResponseGetNotification
// @Failure 403 {string} string "Forbidden: You can only mark your own notifications as read"
// @Failure 404 {string} string "Notification not found"
// @Failure 500 {string} string "Error marking notification as read"
// @Router /api/v1/notification/{id}/read [put]
func NotificationControllerMarkAsRead(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
	if !ok {
		return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Unable to retrieve user information", nil)
	}

	notificationId := ctx.Params("id")
	
	notificationRepo := repositories.NewNotificationRepository(configs.DB)
	notificationService := services.NewNotificationService(notificationRepo)
	
	updatedNotification, err := notificationService.MarkAsRead(notificationId, userInfo.ID, ctx, userInfo)
	if err != nil {
		if err.Error() == "notification not found" {
			return helpers.Response(ctx, fiber.StatusNotFound, "Notification not found", nil)
		}
		if err.Error() == "forbidden: you can only mark your own notifications as read" {
			return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You can only mark your own notifications as read", nil)
		}
		return helpers.Response(ctx, fiber.StatusInternalServerError, "Error marking notification as read", nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Notification marked as read successfully", updatedNotification)
}

// NotificationControllerDelete adalah handler untuk endpoint notification
// @Summary Delete notification
// @Description Delete notification. For now only accessible by notifications with DEVELOPER or SUPERADMIN notifications. If the notification is hard deleted, the notification's avatar will be deleted as well. If the notification is soft deleted, the notification's avatar will be retained. Hard delete mark with is_hard_delete = true and soft delete mark with is_hard_delete = false
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.NotificationIsHardDeleteRequest true "Notification delete request body"
// @Success 200 {array} models.Notification
// @Failure 403 {string} string "Forbidden: You do not have access to delete notification"
// @Failure 404 {string} string "Notification not found"
// @Failure 500 {string} string "Error deleting notification"
// @Router /api/v1/notification/delete [delete]
func NotificationControllerDelete(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Notification info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to delete notifications", nil)
	}

	notificationRequest := new(models.NotificationIsHardDeleteRequest)
	if err := ctx.BodyParser(notificationRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(notificationRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	notificationRepo := repositories.NewNotificationRepository(configs.DB)
	notificationService := services.NewNotificationService(notificationRepo)

	if err := notificationService.DeleteNotifications(notificationRequest, ctx, userInfo); err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	var message string
	if notificationRequest.IsHardDelete == "hardDelete"  {
		message = "Notifications deleted successfully"
	} else {
		message = "Notifications moved to trash successfully"
	}

	return helpers.Response(ctx, fiber.StatusOK, message, nil)
}


// NotificationControllerRestore restores a soft-deleted notification
// @Summary Restore notification
// @Description Restore notification. For now only accessible by notifications with DEVELOPER or SUPERADMIN notifications.
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authorization"
// @Param request body models.NotificationRestoreRequest true "Notification restore request body"
// @Success 200 {array} models.Notification
// @Failure 403 {string} string "Forbidden: You do not have access to restore notification"
// @Failure 404 {string} string "Notification not found"
// @Failure 500 {string} string "Error restoring notification"
// @Router /api/v1/notification/restore [put]
func NotificationControllerRestore(ctx *fiber.Ctx) error {
	userInfo, ok := ctx.Locals("userInfo").(*models.User)
		if !ok {
						return helpers.Response(ctx, fiber.StatusUnauthorized, "Unauthorized: Notification info not found", nil)
		}

	if userInfo.Role.Name != "DEVELOPER" && userInfo.Role.Name != "SUPERADMIN" {
		return helpers.Response(ctx, fiber.StatusForbidden, "Forbidden: You do not have access to restore notifications", nil)
	}

	notificationRequest := new(models.NotificationRestoreRequest)
	if err := ctx.BodyParser(notificationRequest); err != nil {
		return helpers.Response(ctx, fiber.StatusBadRequest, err.Error(), nil)
	}

	if err := helpers.ValidateStruct(notificationRequest); err != nil {
		errorMessage := helpers.ExtractErrorMessages(err)
		return helpers.Response(ctx, fiber.StatusBadRequest, errorMessage, nil)
	}

	notificationRepo := repositories.NewNotificationRepository(configs.DB)
	notificationService := services.NewNotificationService(notificationRepo)

	restoredNotifications, err := notificationService.RestoreNotifications(notificationRequest, ctx, userInfo)
	if err != nil {
		return helpers.Response(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return helpers.Response(ctx, fiber.StatusOK, "Notifications restored successfully", restoredNotifications)
}