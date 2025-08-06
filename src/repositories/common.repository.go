package repositories

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrRoleNotFound    = errors.New("role not found")
	ErrModuleNotFound  = errors.New("module not found")
	ErrModuleTypeNotFound = errors.New("module type not found")
	ErrRoleModuleNotFound = errors.New("role module not found")
	ErrUploadNotFound = errors.New("upload not found")
	ErrCategoryNotFound = errors.New("category not found")
	ErrItemNotFound    = errors.New("item not found")
	ErrItemHistoryNotFound = errors.New("item history not found")
	ErrDatabase        = errors.New("database error")
	ErrUniqueViolation = errors.New("unique constraint violation")
)

func HandleDatabaseError(err error, entity string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		switch entity {
		case "user":
			return ErrUserNotFound
		case "role":
			return ErrRoleNotFound
		case "module":
			return ErrModuleNotFound
		case "module_type":
			return ErrModuleTypeNotFound
		case "role_module":
			return ErrRoleModuleNotFound
		case "upload":
			return ErrUploadNotFound
		case "category":
			return ErrCategoryNotFound
		case "item":
			return ErrItemNotFound
		case "item_history":
			return ErrItemHistoryNotFound
		default:
			return fmt.Errorf("%w: entity not found", ErrDatabase)
		}
	}

	return fmt.Errorf("%w: %v", ErrDatabase, err)
}