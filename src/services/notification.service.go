package services

import (
	"errors"
	"log"
	"time"

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
	var notifications []models.Notification
	var err error

	if userInfo.Role.Name == "DEVELOPER" {
		notifications, err = service.NotificationRepository.FindAll()
	} else {
		notifications, err = service.NotificationRepository.FindByUserID(userInfo.ID)
	}

	if err != nil {
		return nil, err
	}

	var notificationsResponse []models.ResponseGetNotification
	for _, notification := range notifications {
		notificationsResponse = append(notificationsResponse, models.ResponseGetNotification{
			ID:         notification.ID,
			UserID:     notification.UserID,
			Title:      notification.Title,
			Type:       notification.Type,
			Message:    notification.Message,
			IsRead:     notification.IsRead,
			ReadAt:     notification.ReadAt,
			Metadata:   notification.Metadata,
			User:       notification.User,
			CreatedAt:  notification.CreatedAt,
			UpdatedAt:  notification.UpdatedAt,
			DeletedAt:  notification.DeletedAt,
		})
	}

	return notificationsResponse, nil
}

func (service *NotificationService) GetNotificationByID(notificationId string) (*models.ResponseGetNotification, error) {
	notification, err := service.NotificationRepository.FindById(notificationId, false)

	if err != nil {
		return nil, err
	}

	return &models.ResponseGetNotification{
		 ID:          notification.ID,
		 UserID:      notification.UserID,
		 Title:       notification.Title,
		 Type:        notification.Type,
		 Message:     notification.Message,
		 IsRead:      notification.IsRead,
		 ReadAt:      notification.ReadAt,
		 Metadata:    notification.Metadata,
		 User:        notification.User,
	}, nil
}

func (service *NotificationService) MarkAllAsRead(userID uuid.UUID, ctx *fiber.Ctx, userInfo *models.User) error {
	notifications, err := service.NotificationRepository.FindUnreadByUserID(userID)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		notification.IsRead = true
		notification.ReadAt = &time.Time{}
		*notification.ReadAt = time.Now()
		
		_, err := service.NotificationRepository.Update(&notification)
		if err != nil {
			log.Printf("Error updating notification %v: %v\n", notification.ID, err)
			continue
		}
	}

	return nil
}

func (service *NotificationService) MarkMultipleAsRead(notificationIDs []uuid.UUID, userID uuid.UUID, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, notificationID := range notificationIDs {
		notification, err := service.NotificationRepository.FindById(notificationID.String(), false)
		if err != nil {
			if err == repositories.ErrNotificationNotFound {
				log.Printf("Notification not found: %v\n", notificationID)
				continue
			}
			log.Printf("Error finding notification %v: %v\n", notificationID, err)
			continue
		}

		if notification.UserID != userID {
			log.Printf("User %v trying to mark notification %v that doesn't belong to them\n", userID, notificationID)
			continue
		}

		if !notification.IsRead {
			notification.IsRead = true
			notification.ReadAt = &time.Time{}
			*notification.ReadAt = time.Now()
			
			_, err := service.NotificationRepository.Update(notification)
			if err != nil {
				log.Printf("Error updating notification %v: %v\n", notificationID, err)
				continue
			}
		}
	}

	return nil
}

func (service *NotificationService) MarkAsRead(notificationID string, userID uuid.UUID, ctx *fiber.Ctx, userInfo *models.User) (*models.ResponseGetNotification, error) {
	notification, err := service.NotificationRepository.FindById(notificationID, false)
	if err != nil {
		if err == repositories.ErrNotificationNotFound {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}

	// Check if user owns this notification
	if notification.UserID != userID {
		return nil, errors.New("forbidden: you can only mark your own notifications as read")
	}

	if !notification.IsRead {
		notification.IsRead = true
		notification.ReadAt = &time.Time{}
		*notification.ReadAt = time.Now()
		
		updatedNotification, err := service.NotificationRepository.Update(notification)
		if err != nil {
			return nil, err
		}

		return &models.ResponseGetNotification{
			ID:          updatedNotification.ID,
			UserID:      updatedNotification.UserID,
			Title:       updatedNotification.Title,
			Type:        updatedNotification.Type,
			Message:     updatedNotification.Message,
			IsRead:      updatedNotification.IsRead,
			ReadAt:      updatedNotification.ReadAt,
			Metadata:    updatedNotification.Metadata,
			User:        updatedNotification.User,
			CreatedAt:   updatedNotification.CreatedAt,
			UpdatedAt:   updatedNotification.UpdatedAt,
			DeletedAt:   updatedNotification.DeletedAt,
		}, nil
	}

	return &models.ResponseGetNotification{
		ID:          notification.ID,
		UserID:      notification.UserID,
		Title:       notification.Title,
		Type:        notification.Type,
		Message:     notification.Message,
		IsRead:      notification.IsRead,
		ReadAt:      notification.ReadAt,
		Metadata:    notification.Metadata,
		User:        notification.User,
		CreatedAt:   notification.CreatedAt,
		UpdatedAt:   notification.UpdatedAt,
		DeletedAt:   notification.DeletedAt,
	}, nil
}

func (service *NotificationService) DeleteNotifications(notificationRequest *models.NotificationIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, notificationId := range notificationRequest.IDs {
		_, err := service.NotificationRepository.FindById(notificationId.String(), false)
		if err != nil {
			if err == repositories.ErrNotificationNotFound {
				log.Printf("Notification not found: %v\n", notificationId)
				continue
			}
			log.Printf("Error finding notification %v: %v\n", notificationId, err)
			return errors.New("error finding usenotificationr")
		}

		if notificationRequest.IsHardDelete == "hardDelete" {
			if err := service.NotificationRepository.Delete(notificationId.String(), true); err != nil {
				log.Printf("Error hard deleting notification %v: %v\n", notificationId, err)
				return errors.New("error hard deleting notification")
			}
		} else {
			if err := service.NotificationRepository.Delete(notificationId.String(), false); err != nil {
				log.Printf("Error soft deleting notification %v: %v\n", notificationId, err)
				return errors.New("error soft deleting notification")
			}
		}
	}

	return nil
}

func (service *NotificationService) RestoreNotifications(notification *models.NotificationRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Notification, error) {
	var restoredNotifications []models.Notification

	for _, notificationId := range notification.IDs {
		notification := &models.Notification{ID: notificationId}

		restoredNotification, err := service.NotificationRepository.Restore(notification, notificationId.String())
		if err != nil {
			if err == repositories.ErrNotificationNotFound {
				log.Printf("Notification not found: %v\n", notificationId)
				continue
			}
			log.Printf("Error restoring notification %v: %v\n", notificationId, err)
			return nil, errors.New("error restoring notification")
		}

		restoredNotifications = append(restoredNotifications, *restoredNotification)
	}

	return restoredNotifications, nil
}