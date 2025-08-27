package helpers

import (
	"log"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/websockets"
	"github.com/google/uuid"
)

var allowedRoles = map[string][]string{
	// Low stock
	"low_stock": {"SUPERADMIN", "DEVELOPER", "SALES"},
	"consigment_item": {"SUPERADMIN", "DEVELOPER", "SALES"},
}

func SendNotificationAuto(
	notifType string,
	title string,
	message string,
	metadata map[string]interface{},
) error {
	rolesAllowed, ok := allowedRoles[notifType]
	if !ok {
		log.Printf("⚠️ No allowed roles defined for notification type: %s", notifType)
		return nil
	}

	var users []models.User
	if err := configs.DB.Preload("Role").Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		role := strings.ToUpper(user.Role.Name)
		for _, allowedRole := range rolesAllowed {
			if role == allowedRole {
				notification := models.Notification{
					ID:       uuid.New(),
					UserID:   user.ID,
					Type:     notifType,
					Title:    title,
					Message:  message,
					IsRead:   false,
					Metadata: metadata,
				}
				
				if err := configs.DB.Create(&notification).Error; err != nil {
					log.Printf("❌ Failed to save notification to DB for user %s: %v", user.ID, err)
					continue
				}

				if err := configs.DB.Preload("User.Role").First(&notification, "id = ?", notification.ID).Error; err != nil {
					log.Printf("❌ Failed to load notification with user relation: %v", err)
					continue
				}

				websockets.SendToUser(user.ID.String(), "notification", notification)
			}
		}
	}

	return nil
}
