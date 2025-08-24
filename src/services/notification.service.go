package services

import (
	"errors"
	"log"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type NotificationService struct {
	NotificationRepository repositories.NotificationRepository
}

func NewNotificationService(notificationRepo repositories.NotificationRepository) *NotificationService {
	return &NotificationService{
		NotificationRepository: notificationRepo,
	}
}

func (service *NotificationService) GetAllNotifications(userInfo *models.User) ([]models.ResponseGetNotification, error) {
	notifications, err := service.NotificationRepository.FindByUserID(nil, userInfo.ID)
	if err != nil {
		return nil, err
	}

	var out []models.ResponseGetNotification
	for _, n := range notifications {
		out = append(out, models.ResponseGetNotification{
			ID:        n.ID,
			UserID:    n.UserID,
			Title:     n.Title,
			Type:      n.Type,
			Message:   n.Message,
			IsRead:    n.IsRead,
			ReadAt:    n.ReadAt,
			Metadata:  n.Metadata,
			User:      n.User,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
			DeletedAt: n.DeletedAt,
		})
	}
	return out, nil
}

func (service *NotificationService) GetNotificationByID(notificationId string) (*models.ResponseGetNotification, error) {
	n, err := service.NotificationRepository.FindById(nil,notificationId, false)
	if err != nil {
		return nil, err
	}

	return &models.ResponseGetNotification{
		ID:        n.ID,
		UserID:    n.UserID,
		Title:     n.Title,
		Type:      n.Type,
		Message:   n.Message,
		IsRead:    n.IsRead,
		ReadAt:    n.ReadAt,
		Metadata:  n.Metadata,
		User:      n.User,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
		DeletedAt: n.DeletedAt,
	}, nil
}

func (service *NotificationService) MarkAllAsRead(userID uuid.UUID, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx
	_ = userInfo

	notifications, err := service.NotificationRepository.FindUnreadByUserID(nil, userID)
	if err != nil {
		return err
	}
	if len(notifications) == 0 {
		return nil
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()
	for i := range notifications {
		notifications[i].IsRead = true
		notifications[i].ReadAt = &now

		if _, err := service.NotificationRepository.Update(tx, &notifications[i]); err != nil {
			_ = tx.Rollback()
			log.Printf("Error updating notification %v: %v\n", notifications[i].ID, err)
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (service *NotificationService) MarkMultipleAsRead(notificationIDs []uuid.UUID, userID uuid.UUID, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx
	_ = userInfo

	if len(notificationIDs) == 0 {
		return nil
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()
	for _, id := range notificationIDs {
		n, err := service.NotificationRepository.FindById(tx, id.String(), false)
		if err != nil {
			if err == repositories.ErrNotificationNotFound {
				log.Printf("Notification not found: %v\n", id)
				continue
			}
			_ = tx.Rollback()
			log.Printf("Error finding notification %v: %v\n", id, err)
			return err
		}

		if n.UserID != userID {
			log.Printf("User %v trying to mark notification %v that doesn't belong to them\n", userID, id)
			continue
		}

		if !n.IsRead {
			n.IsRead = true
			n.ReadAt = &now

			if _, err := service.NotificationRepository.Update(tx, n); err != nil {
				_ = tx.Rollback()
				log.Printf("Error updating notification %v: %v\n", id, err)
				return err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (service *NotificationService) MarkAsRead(notificationID string, userID uuid.UUID, ctx *fiber.Ctx, userInfo *models.User) (*models.ResponseGetNotification, error) {
	_ = ctx
	_ = userInfo

	n, err := service.NotificationRepository.FindById(nil, notificationID, false)
	if err != nil {
		if err == repositories.ErrNotificationNotFound {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}

	if n.UserID != userID {
		return nil, errors.New("forbidden: you can only mark your own notifications as read")
	}

	if n.IsRead {
		return &models.ResponseGetNotification{
			ID:        n.ID,
			UserID:    n.UserID,
			Title:     n.Title,
			Type:      n.Type,
			Message:   n.Message,
			IsRead:    n.IsRead,
			ReadAt:    n.ReadAt,
			Metadata:  n.Metadata,
			User:      n.User,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
			DeletedAt: n.DeletedAt,
		}, nil
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now

	updated, err := service.NotificationRepository.Update(tx, n)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &models.ResponseGetNotification{
		ID:        updated.ID,
		UserID:    updated.UserID,
		Title:     updated.Title,
		Type:      updated.Type,
		Message:   updated.Message,
		IsRead:    updated.IsRead,
		ReadAt:    updated.ReadAt,
		Metadata:  updated.Metadata,
		User:      updated.User,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
		DeletedAt: updated.DeletedAt,
	}, nil
}

func (service *NotificationService) DeleteNotifications(req *models.NotificationIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx
	_ = userInfo

	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			return tx.Error
		}

		if _, err := service.NotificationRepository.FindById(tx, id.String(), false); err != nil {
			_ = tx.Rollback()
			if err == repositories.ErrNotificationNotFound {
				log.Printf("Notification not found: %v\n", id)
				continue
			}
			log.Printf("Error finding notification %v: %v\n", id, err)
			return errors.New("error finding notification")
		}

		if err := service.NotificationRepository.Delete(tx, id.String(), req.IsHardDelete == "hardDelete"); err != nil {
			_ = tx.Rollback()
			log.Printf("Error deleting notification %v: %v\n", id, err)
			if req.IsHardDelete == "hardDelete" {
				return errors.New("error hard deleting notification")
			}
			return errors.New("error soft deleting notification")
		}

		if err := tx.Commit().Error; err != nil {
			return err
		}
	}
	return nil
}

func (service *NotificationService) RestoreNotifications(req *models.NotificationRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Notification, error) {
	_ = ctx
	_ = userInfo

	var restored []models.Notification

	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			return nil, tx.Error
		}

		res, err := service.NotificationRepository.Restore(tx, id.String())
		if err != nil {
			_ = tx.Rollback()
			if err == repositories.ErrNotificationNotFound {
				log.Printf("Notification not found: %v\n", id)
				continue
			}
			log.Printf("Error restoring notification %v: %v\n", id, err)
			return nil, errors.New("error restoring notification")
		}

		if err := tx.Commit().Error; err != nil {
			return nil, err
		}

		restored = append(restored, *res)
	}
	return restored, nil
}
