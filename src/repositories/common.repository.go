package repositories

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

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
	ErrAreaNotFound    = errors.New("area not found")
	ErrCustomerTypeNotFound = errors.New("customer type not found")
	ErrCustomerNotFound = errors.New("customer not found")
	ErrSalesPersonNotFound = errors.New("sales person not found")
	ErrSalesAssignmentNotFound = errors.New("sales assignment not found")
	ErrSupplierNotFound = errors.New("supplier not found")
	ErrSalesOrderNotFound = errors.New("sales order not found")
	ErrSalesOrderItemNotFound = errors.New("sales order item not found")
	ErrPurchaseOrderNotFound = errors.New("purchase order not found")
	ErrPurchaseOrderItemNotFound = errors.New("purchase order item not found")
	ErrNotificationNotFound = errors.New("notification not found")
	ErrUoMNotFound = errors.New("UoM not found")
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
		case "area":
			return ErrAreaNotFound
		case "customer_type":
			return ErrCustomerTypeNotFound
		case "customer":
			return ErrCustomerNotFound
		case "sales_person":
			return ErrSalesPersonNotFound
		case "sales_assignment":
			return ErrSalesAssignmentNotFound
		case "supplier":
			return ErrSupplierNotFound
		case "sales_order":
			return ErrSalesOrderNotFound
		case "sales_order_item":
			return ErrSalesOrderItemNotFound
		case "purchase_order":
			return ErrPurchaseOrderNotFound
		case "purchase_order_item":
			return ErrPurchaseOrderItemNotFound
		case "notification":
			return ErrNotificationNotFound
		case "uom":
			return ErrUoMNotFound
		default:
			return fmt.Errorf("%w: entity not found", ErrDatabase)
		}
	}

	return fmt.Errorf("%w: %v", ErrDatabase, err)
}

func IsUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate key value")
}

func IsValidEmail(s string) bool {
	if len(s) < 6 || !strings.Contains(s, "@") || !strings.Contains(s, ".") {
		return false
	}
	return true
}

func IsHexColor(s string) bool {
	var hexColorRx = regexp.MustCompile(`^#(?:[0-9a-f]{3}|[0-9a-f]{6})$`)
	return hexColorRx.MatchString(strings.ToLower(strings.TrimSpace(s)))
}