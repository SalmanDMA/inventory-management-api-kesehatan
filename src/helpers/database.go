package helpers

import (
	"errors"
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"gorm.io/gorm"
)

func PreloadRole(db *gorm.DB) *gorm.DB {
	return db.Preload("Role")
}

func HandleDatabaseError(err error, entity string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		switch entity {
		case "user":
			return repositories.ErrUserNotFound
		case "role":
			return repositories.ErrRoleNotFound
		case "module":
			return repositories.ErrModuleNotFound
		case "role_module":
			return repositories.ErrRoleModuleNotFound
		default:
			return fmt.Errorf("%w: entity not found", repositories.ErrDatabase)
		}
	}

	return fmt.Errorf("%w: %v", repositories.ErrDatabase, err)
}