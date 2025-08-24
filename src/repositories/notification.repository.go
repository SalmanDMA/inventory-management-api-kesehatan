package repositories

import (
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type NotificationRepository interface {
	FindAll(tx *gorm.DB) ([]models.Notification, error)
	FindById(tx *gorm.DB, notificationId string, includeTrashed bool) (*models.Notification, error)
	FindByUserID(tx *gorm.DB, userID uuid.UUID) ([]models.Notification, error)
	FindUnreadByUserID(tx *gorm.DB, userID uuid.UUID) ([]models.Notification, error)
	Update(tx *gorm.DB, notification *models.Notification) (*models.Notification, error)
	Delete(tx *gorm.DB, notificationId string, isHardDelete bool) error
	Restore(tx *gorm.DB, notificationId string) (*models.Notification, error)
}

// ==============================
// Implementation
// ==============================

type NotificationRepositoryImpl struct {
	DB *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepositoryImpl {
	return &NotificationRepositoryImpl{DB: db}
}

func (r *NotificationRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *NotificationRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.useDB(tx).
		Unscoped().
		Preload("User").
		Find(&notifications).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return notifications, nil
}

func (r *NotificationRepositoryImpl) FindById(tx *gorm.DB, notificationId string, includeTrashed bool) (*models.Notification, error) {
	var n models.Notification
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.
		Preload("User").
		First(&n, "id = ?", notificationId).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return &n, nil
}

func (r *NotificationRepositoryImpl) FindByUserID(tx *gorm.DB, userID uuid.UUID) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.useDB(tx).
		Preload("User").
		Where("user_id = ?", userID).
		Find(&notifications).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return notifications, nil
}

func (r *NotificationRepositoryImpl) FindUnreadByUserID(tx *gorm.DB, userID uuid.UUID) ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.useDB(tx).
		Preload("User").
		Where("user_id = ? AND is_read = ?", userID, false).
		Find(&notifications).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return notifications, nil
}

// ---------- Mutations ----------

func (r *NotificationRepositoryImpl) Update(tx *gorm.DB, notification *models.Notification) (*models.Notification, error) {
	if notification.ID == uuid.Nil {
		return nil, fmt.Errorf("notification ID cannot be empty")
	}
	if err := r.useDB(tx).Save(notification).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return notification, nil
}

func (r *NotificationRepositoryImpl) Delete(tx *gorm.DB, notificationId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var n models.Notification
	if err := db.Unscoped().First(&n, "id = ?", notificationId).Error; err != nil {
		return HandleDatabaseError(err, "notification")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&n).Error; err != nil {
			return HandleDatabaseError(err, "notification")
		}
	} else {
		if err := db.Delete(&n).Error; err != nil {
			return HandleDatabaseError(err, "notification")
		}
	}
	return nil
}

func (r *NotificationRepositoryImpl) Restore(tx *gorm.DB, notificationId string) (*models.Notification, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.Notification{}).
		Where("id = ?", notificationId).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}

	var restored models.Notification
	if err := db.
		Preload("User").
		First(&restored, "id = ?", notificationId).Error; err != nil {
		return nil, HandleDatabaseError(err, "notification")
	}
	return &restored, nil
}
