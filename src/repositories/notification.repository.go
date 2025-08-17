package repositories

import (
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	FindAll() ([]models.Notification, error)
	FindById(notificationId string, isSoftDelete bool) (*models.Notification, error)
	FindByUserID(userID uuid.UUID) ([]models.Notification, error)
	FindUnreadByUserID(userID uuid.UUID) ([]models.Notification, error)
	Update(notification *models.Notification) (*models.Notification, error)
	Delete(notificationId string, isHardDelete bool) error
	Restore(notification *models.Notification, notificationId string) (*models.Notification, error)
}

type NotificationRepositoryImpl struct{
	DB *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepositoryImpl {
	return &NotificationRepositoryImpl{DB: db}
}

func (r *NotificationRepositoryImpl) FindAll() ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.DB.
	 Unscoped().
	 Preload("User").
	 Find(&notifications).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return notifications, nil
}

func (r *NotificationRepositoryImpl) FindById(notificationId string, isSoftDelete bool) (*models.Notification, error) {
	var notification *models.Notification
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	if err := db.
		Preload("User").	
		First(&notification, "id = ?", notificationId).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	
	return notification, nil
}

func (r *NotificationRepositoryImpl) FindByUserID(userID uuid.UUID) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.DB.
		Preload("User").
		Where("user_id = ?", userID).
		Find(&notifications).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return notifications, nil
}

func (r *NotificationRepositoryImpl) FindUnreadByUserID(userID uuid.UUID) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.DB.
		Preload("User").
		Where("user_id = ? AND is_read = ?", userID, false).
		Find(&notifications).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return notifications, nil
}

func (r *NotificationRepositoryImpl) Update(notification *models.Notification) (*models.Notification, error) {

	if notification.ID == uuid.Nil {
		return nil, fmt.Errorf("Notification ID cannot be empty")
	}

	if err := r.DB.Save(&notification).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return notification, nil
}

func (r *NotificationRepositoryImpl) Delete(notificationId string, isHardDelete bool) error {
	var notification *models.Notification

	if err := r.DB.Unscoped().First(&notification, "id = ?", notificationId).Error; err != nil {
		return HandleDatabaseError(err, "notification")
	}
	
	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&notification).Error; err != nil {
			return HandleDatabaseError(err, "notification")
		}
	} else {
		if err := r.DB.Delete(&notification).Error; err != nil {
			return HandleDatabaseError(err, "notification")
		}
	}
	return nil
}

func (r *NotificationRepositoryImpl) Restore(notification *models.Notification, notificationId string) (*models.Notification, error) {
	if err := r.DB.Unscoped().Model(notification).Where("id = ?", notificationId).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredNotification *models.Notification
	if err := r.DB.Unscoped().First(&restoredNotification, "id = ?", notificationId).Error; err != nil {
		return nil, err
	}
	
	return restoredNotification, nil
}