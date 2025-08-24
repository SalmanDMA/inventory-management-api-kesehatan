package seeders

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func strptr(s string) *string { return &s }
func fptr(f float64) *float64 { return &f }
func tptr(loc *time.Location, y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, loc)
	return &t
}

//
// ROLE SEEDER
//

func SeedRoles(db *gorm.DB) error {
	log.Println("Seeding roles...")

	roles := []models.Role{
		{Name: "SUPERADMIN", Alias: "SA", Color: "#f00f00", Description: "Akun Super Admin"},
		{Name: "DEVELOPER", Alias: "DEV", Color: "#000000", Description: "Akun Developer"},
		{Name: "SALES", Alias: "SP", Color: "#00f000", Description: "Akun Sales"},
	}

	for _, role := range roles {
		role.ID = uuid.New()
		if err := db.Where("name = ?", role.Name).FirstOrCreate(&role).Error; err != nil {
			return fmt.Errorf("failed to seed role '%s': %w", role.Name, err)
		}
	}
	return nil
}

//
// MODULE SEEDER
//

func SeedModules(db *gorm.DB) error {
	log.Println("Seeding modules...")

	appVersion := os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = "/api/v1"
	}

	// Seed module types
	moduleTypes := []models.ModuleType{
		{Name: "Menu Directory", Description: "Directory menu", Icon: "mdi:folder"},
		{Name: "Route Menu", Description: "Route-based menu", Icon: "mdi:routes"},
		{Name: "Route Hidden", Description: "Hidden routes not displayed in frontend", Icon: "mdi:eye-off"},
		{Name: "Service API", Description: "CRUD operations and API endpoints", Icon: "mdi:server"},
	}

	for _, mt := range moduleTypes {
		mt.ID = uuid.New()
		db.Where("name = ?", mt.Name).FirstOrCreate(&mt)
	}

	// Build a map for module type lookup
	moduleTypeMap := make(map[string]uuid.UUID)
	for _, name := range []string{"Menu Directory", "Route Menu", "Service API", "Route Hidden"} {
		var mt models.ModuleType
		if err := db.Where("name = ?", name).First(&mt).Error; err != nil {
			return fmt.Errorf("failed to fetch module type '%s': %w", name, err)
		}
		moduleTypeMap[name] = mt.ID
	}

	// Parent modules
	var accessControlModule models.Module
	db.Where("name = ?", "Access Control").FirstOrCreate(&accessControlModule, models.Module{
		Name:         "Access Control",
		Icon:         "mdi:shield",
		ModuleTypeID: moduleTypeMap["Menu Directory"],
		Description:  "Access Control Page",
	})

	var masterDataModule models.Module
	db.Where("name = ?", "Master Data").FirstOrCreate(&masterDataModule, models.Module{
		Name:         "Master Data",
		Icon:         "mdi:database",
		ModuleTypeID: moduleTypeMap["Menu Directory"],
		Description:  "Master Data Page",
	})

	var transactionDataModule models.Module
	db.Where("name = ?", "Transactions").FirstOrCreate(&transactionDataModule, models.Module{
		Name:         "Transactions",
		Icon:         "mdi:cart",
		ModuleTypeID: moduleTypeMap["Menu Directory"],
		Description:  "Transactions Data Page",
	})

	var analyticsModule models.Module
	db.Where("name = ?", "Analytics").FirstOrCreate(&analyticsModule, models.Module{
		Name:         "Analytics",
		Icon:         "mdi:chart-line",
		ModuleTypeID: moduleTypeMap["Menu Directory"],
		Description:  "Analytics Page",
	})

	// Child modules
	modules := []models.Module{
		{Name: "Dashboard", Route: "/dashboard", Icon: "mdi:home", ModuleTypeID: moduleTypeMap["Menu Directory"], Description: "Main Dashboard"},
		{Name: "Settings", Route: "/dashboard/settings", Icon: "mdi:settings", ModuleTypeID: moduleTypeMap["Menu Directory"], Description: "Settings Page"},
		{Name: "Users", Route: "/dashboard/users", Icon: "mdi:account", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &accessControlModule.ID, Description: "User Management Page"},
		{Name: "Roles", Route: "/dashboard/roles", Icon: "mdi:account-group", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &accessControlModule.ID, Description: "Role Management Page"},
		{Name: "Modules", Route: "/dashboard/modules", Icon: "mdi:layers-triple", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &accessControlModule.ID, Description: "Module Management Page"},
		{Name: "Module Types", Route: "/dashboard/module-types", Icon: "mdi:layers", ModuleTypeID: moduleTypeMap["Route Hidden"], ParentID: &accessControlModule.ID, Description: "Module Type Management Page"},
		{Name: "Items", Route: "/dashboard/items", Icon: "mdi:inboxes", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &masterDataModule.ID, Description: "Item Management Page"},
		{Name: "UoMs", Route: "/dashboard/uoms", Icon: "mdi:shape", ModuleTypeID: moduleTypeMap["Route Hidden"], ParentID: &masterDataModule.ID, Description: "Unit of Measurement Management Page"},
		{Name: "Categories", Route: "/dashboard/categories", Icon: "mdi:category", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &masterDataModule.ID, Description: "Category Management Page"},
		{Name: "Item History", Route: "/dashboard/item-history", Icon: "mdi:history", ModuleTypeID: moduleTypeMap["Route Hidden"], ParentID: &masterDataModule.ID, Description: "Item History Page - Hidden Route"},
		{Name: "Areas", Route: "/dashboard/areas", Icon: "mdi:map", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &masterDataModule.ID, Description: "Area Management Page"},
		{Name: "Customers", Route: "/dashboard/customers", Icon: "mdi:map-marker", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &masterDataModule.ID, Description: "Customer Management Page"},
		{Name: "Customer Types", Route: "/dashboard/customer-types", Icon: "mdi:map-marker", ModuleTypeID: moduleTypeMap["Route Hidden"], ParentID: &masterDataModule.ID, Description: "Customer Type Management Page"},
		{Name: "Sales", Route: "/dashboard/sales", Icon: "mdi:calendar-user-outline", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &masterDataModule.ID, Description: "Sales Management Page"},
		{Name: "Suppliers", Route: "/dashboard/suppliers", Icon: "mdi:account-group", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &masterDataModule.ID, Description: "Supplier Management Page"},
		{Name: "Purchase Orders", Route: "/dashboard/purchase-orders", Icon: "mdi:order-bool-descending", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &transactionDataModule.ID, Description: "Purchase Order Management Page"},
		{Name: "Sales Orders", Route: "/dashboard/sales-orders", Icon: "mdi:order-bool-ascending", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &transactionDataModule.ID, Description: "Sales Order Management Page"},
		{Name: "Sales Reports", Route: "/dashboard/sales-reports", Icon: "mdi:chart-line", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &analyticsModule.ID, Description: "Sales Report Management Page"},
		{Name: "Notifications", Route: "/dashboard/notifications", Icon: "mdi:bell", ModuleTypeID: moduleTypeMap["Route Menu"], ParentID: &analyticsModule.ID, Description: "Notification Management Page"},
	}

	for _, m := range modules {
		db.Where("name = ?", m.Name).FirstOrCreate(&m)
	}

	// Get "Users" module for service parent
	var usersModule models.Module
	if err := db.Where("name = ?", "Users").First(&usersModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Users module: %w", err)
	}

	// Service routes for user management
	userServiceModules := []models.Module{
		{Name: "Get User Profile", Path: fmt.Sprintf("%s/user/me", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get current user profile", ParentID: &usersModule.ID},
		{Name: "Update User Profile", Path: fmt.Sprintf("%s/user/me", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update current user profile", ParentID: &usersModule.ID},
		{Name: "Restore User", Path: fmt.Sprintf("%s/user/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted user", ParentID: &usersModule.ID},
		{Name: "Delete User", Path: fmt.Sprintf("%s/user/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete user permanently", ParentID: &usersModule.ID},
		{Name: "Get All Users", Path: fmt.Sprintf("%s/user", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all users", ParentID: &usersModule.ID},
		{Name: "Create User", Path: fmt.Sprintf("%s/user", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new user", ParentID: &usersModule.ID},
		{Name: "Get User By ID", Path: fmt.Sprintf("%s/user/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get user by ID", ParentID: &usersModule.ID},
		{Name: "Update User", Path: fmt.Sprintf("%s/user/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update user by ID", ParentID: &usersModule.ID},
	}

	for _, sm := range userServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Roles" module for service parent
	var rolesModule models.Module
	if err := db.Where("name = ?", "Roles").First(&rolesModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Roles module: %w", err)
	}

	// Service routes for role management
	roleServiceModules := []models.Module{
		{Name: "Get All Roles", Path: fmt.Sprintf("%s/role", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all roles", ParentID: &rolesModule.ID},
		{Name: "Create Role", Path: fmt.Sprintf("%s/role", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new role", ParentID: &rolesModule.ID},
		{Name: "Restore Role", Path: fmt.Sprintf("%s/role/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted role", ParentID: &rolesModule.ID},
		{Name: "Delete Role", Path: fmt.Sprintf("%s/role/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete role permanently", ParentID: &rolesModule.ID},
		{Name: "Get Role By ID", Path: fmt.Sprintf("%s/role/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get role by ID", ParentID: &rolesModule.ID},
		{Name: "Update Role", Path: fmt.Sprintf("%s/role/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update role by ID", ParentID: &rolesModule.ID},
		{Name: "Get Role Modules", Path: fmt.Sprintf("%s/role/:roleId/module", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get modules with role information", ParentID: &rolesModule.ID},
		{Name: "Create or Update Role Module", Path: fmt.Sprintf("%s/role/:roleId/module", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create or update role module assignment", ParentID: &rolesModule.ID},
	}

	for _, sm := range roleServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Modules" module for service parent
	var modulesModule models.Module
	if err := db.Where("name = ?", "Modules").First(&modulesModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Modules module: %w", err)
	}

	// Service routes for module management
	moduleServiceModules := []models.Module{
		{Name: "Restore Module", Path: fmt.Sprintf("%s/module/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted module", ParentID: &modulesModule.ID},
		{Name: "Delete Module", Path: fmt.Sprintf("%s/module/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete module permanently", ParentID: &modulesModule.ID},
		{Name: "Get All Modules", Path: fmt.Sprintf("%s/module", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all modules", ParentID: &modulesModule.ID},
		{Name: "Create Module", Path: fmt.Sprintf("%s/module", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new module", ParentID: &modulesModule.ID},
		{Name: "Get Module By ID", Path: fmt.Sprintf("%s/module/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get module by ID", ParentID: &modulesModule.ID},
		{Name: "Update Module", Path: fmt.Sprintf("%s/module/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update module by ID", ParentID: &modulesModule.ID},
	}

	for _, sm := range moduleServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Module Types" module for service parent
	var moduleTypesModule models.Module
	if err := db.Where("name = ?", "Module Types").First(&moduleTypesModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Module Types module: %w", err)
	}

	// Service routes for module type management
	moduleTypeServiceModules := []models.Module{
		{Name: "Get All Module Types", Path: fmt.Sprintf("%s/module-type", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all module types", ParentID: &moduleTypesModule.ID},
		{Name: "Create Module Type", Path: fmt.Sprintf("%s/module-type", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new module type", ParentID: &moduleTypesModule.ID},
		{Name: "Restore Module Type", Path: fmt.Sprintf("%s/module-type/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted module type", ParentID: &moduleTypesModule.ID},
		{Name: "Delete Module Type", Path: fmt.Sprintf("%s/module-type/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete module type permanently", ParentID: &moduleTypesModule.ID},
		{Name: "Get Module Type By ID", Path: fmt.Sprintf("%s/module-type/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get module type by ID", ParentID: &moduleTypesModule.ID},
		{Name: "Update Module Type", Path: fmt.Sprintf("%s/module-type/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update module type by ID", ParentID: &moduleTypesModule.ID},
	}

	for _, sm := range moduleTypeServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Items" module for service parent
	var itemsModule models.Module
	if err := db.Where("name = ?", "Items").First(&itemsModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Items module: %w", err)
	}

	// Service routes for item management
	itemServiceModules := []models.Module{
		{Name: "Get All Items", Path: fmt.Sprintf("%s/item", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all items", ParentID: &itemsModule.ID},
		{Name: "Create Item", Path: fmt.Sprintf("%s/item", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new item", ParentID: &itemsModule.ID},
		{Name: "Restore Item", Path: fmt.Sprintf("%s/item/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted item", ParentID: &itemsModule.ID},
		{Name: "Delete Item", Path: fmt.Sprintf("%s/item/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete item permanently", ParentID: &itemsModule.ID},
		{Name: "Get Item By ID", Path: fmt.Sprintf("%s/item/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get item by ID", ParentID: &itemsModule.ID},
		{Name: "Update Item", Path: fmt.Sprintf("%s/item/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update item by ID", ParentID: &itemsModule.ID},
	}

	for _, sm := range itemServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "UoMs" module for service parent
	var uomsModule models.Module
	if err := db.Where("name = ?", "UoMs").First(&uomsModule).Error; err != nil {
		return fmt.Errorf("failed to fetch UoMs module: %w", err)
	}

	// Service routes for UoM management
	uomServiceModules := []models.Module{
		{Name: "Get All UoMs", Path: fmt.Sprintf("%s/uom", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all UoMs", ParentID: &uomsModule.ID},
		{Name: "Create UoM", Path: fmt.Sprintf("%s/uom", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new UoM", ParentID: &uomsModule.ID},
		{Name: "Restore UoM", Path: fmt.Sprintf("%s/uom/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted UoM", ParentID: &uomsModule.ID},
		{Name: "Delete UoM", Path: fmt.Sprintf("%s/uom/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete UoM permanently", ParentID: &uomsModule.ID},
		{Name: "Get UoM By ID", Path: fmt.Sprintf("%s/uom/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get UoM by ID", ParentID: &uomsModule.ID},
		{Name: "Update UoM", Path: fmt.Sprintf("%s/uom/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update UoM by ID", ParentID: &uomsModule.ID},
	}
	
	for _, sm := range uomServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Categories" module for service parent
	var categoriesModule models.Module
	if err := db.Where("name = ?", "Categories").First(&categoriesModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Categories module: %w", err)
	}

	// Service routes for category management
	categoryServiceModules := []models.Module{
		{Name: "Get All Categories", Path: fmt.Sprintf("%s/category", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all categories", ParentID: &categoriesModule.ID},
		{Name: "Create Category", Path: fmt.Sprintf("%s/category", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new category", ParentID: &categoriesModule.ID},
		{Name: "Restore Category", Path: fmt.Sprintf("%s/category/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted category", ParentID: &categoriesModule.ID},
		{Name: "Delete Category", Path: fmt.Sprintf("%s/category/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete category permanently", ParentID: &categoriesModule.ID},
		{Name: "Get Category By ID", Path: fmt.Sprintf("%s/category/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get category by ID", ParentID: &categoriesModule.ID},
		{Name: "Update Category", Path: fmt.Sprintf("%s/category/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update category by ID", ParentID: &categoriesModule.ID},
	}

	for _, sm := range categoryServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Item History" module for service parent
	var itemHistoryModule models.Module
	if err := db.Where("name = ?", "Item History").First(&itemHistoryModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Item History module: %w", err)
	}

	// Service routes for item history management
	itemHistoryServiceModules := []models.Module{
		{Name: "Get All Item Histories", Path: fmt.Sprintf("%s/item-history", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all item histories", ParentID: &itemHistoryModule.ID},
		{Name: "Create Item History", Path: fmt.Sprintf("%s/item-history", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new item history", ParentID: &itemHistoryModule.ID},
		{Name: "Restore Item History", Path: fmt.Sprintf("%s/item-history/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted item history", ParentID: &itemHistoryModule.ID},
		{Name: "Delete Item History", Path: fmt.Sprintf("%s/item-history/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete item history permanently", ParentID: &itemHistoryModule.ID},
	}

	for _, sm := range itemHistoryServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Areas" module for service parent
	var areasModule models.Module
	if err := db.Where("name = ?", "Areas").First(&areasModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Areas module: %w", err)
	}

	// Service routes for area management
	areasServiceModules := []models.Module{
		{Name: "Get All Areas", Path: fmt.Sprintf("%s/area", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all areas", ParentID: &areasModule.ID},
		{Name: "Create Area", Path: fmt.Sprintf("%s/area", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new area", ParentID: &areasModule.ID},
		{Name: "Restore Area", Path: fmt.Sprintf("%s/area/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted area", ParentID: &areasModule.ID},
		{Name: "Delete Area", Path: fmt.Sprintf("%s/area/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete area permanently", ParentID: &areasModule.ID},
	}

	for _, sm := range areasServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Customers" module for service parent
	var customersModule models.Module
	if err := db.Where("name = ?", "Customers").First(&customersModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Customers module: %w", err)
	}

	// Service routes for customer management
	customersModules := []models.Module{
		{Name: "Get All Customers", Path: fmt.Sprintf("%s/customer", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all customers", ParentID: &customersModule.ID},
		{Name: "Create Customer", Path: fmt.Sprintf("%s/customer", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new customer", ParentID: &customersModule.ID},
		{Name: "Restore Customer", Path: fmt.Sprintf("%s/customer/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted customer", ParentID: &customersModule.ID},
		{Name: "Delete Customer", Path: fmt.Sprintf("%s/customer/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete customer permanently", ParentID: &customersModule.ID},
	}

	for _, sm := range customersModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Customer Types" module for service parent
	var customerTypesModule models.Module
	if err := db.Where("name = ?", "Customer Types").First(&customerTypesModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Customer Types module: %w", err)
	}

	// Service routes for customer type management 
	customerTypesModules := []models.Module{
		{Name: "Get All Customer Types", Path: fmt.Sprintf("%s/customer-type", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all customer types", ParentID: &customerTypesModule.ID},
		{Name: "Create Customer Type", Path: fmt.Sprintf("%s/customer-type", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new customer type", ParentID: &customerTypesModule.ID},
		{Name: "Restore Customer Type", Path: fmt.Sprintf("%s/customer-type/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted customer type", ParentID: &customerTypesModule.ID},
		{Name: "Delete Customer Type", Path: fmt.Sprintf("%s/customer-type/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete customer type permanently", ParentID: &customerTypesModule.ID},
	}

	for _, sm := range customerTypesModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Sales" module for service parent
	var salesModule models.Module
	if err := db.Where("name = ?", "Sales").First(&salesModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Sales module: %w", err)
	}

	// Service routes for sales management
	salesServiceModules := []models.Module{
		{Name: "Get All Sales", Path: fmt.Sprintf("%s/sales-person", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all sales", ParentID: &salesModule.ID},
		{Name: "Create Sales", Path: fmt.Sprintf("%s/sales-person", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new sales", ParentID: &salesModule.ID},
		{Name: "Restore Sales", Path: fmt.Sprintf("%s/sales-person/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted sales", ParentID: &salesModule.ID},
		{Name: "Delete Sales", Path: fmt.Sprintf("%s/sales-person/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete sales permanently", ParentID: &salesModule.ID},
		{Name: "Get Sales By ID", Path: fmt.Sprintf("%s/sales-person/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get sales by ID", ParentID: &salesModule.ID},
		{Name: "Update Sales", Path: fmt.Sprintf("%s/sales-person/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update sales by ID", ParentID: &salesModule.ID},
		{Name: "Get Sales Assignment", Path: fmt.Sprintf("%s/sales-person/:salesPersonId/area", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get areas with sales information", ParentID: &salesModule.ID},
		{Name: "Create or Update Sales Assignment", Path: fmt.Sprintf("%s/sales-person/:salesPersonId/area", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create or update sales assignment", ParentID: &salesModule.ID},
	}

	for _, sm := range salesServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}


	// Get "Suppliers" module for service parent
	var suppliersModule models.Module
	if err := db.Where("name = ?", "Suppliers").First(&suppliersModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Suppliers module: %w", err)
	}

	// Service routes for supplier management
	suppliersServiceModules := []models.Module{
		{Name: "Get All Paginated Suppliers", Path: fmt.Sprintf("%s/supplier", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get all paginatedsuppliers", ParentID: &suppliersModule.ID},
		{Name: "Get All Suppliers", Path: fmt.Sprintf("%s/supplier/all", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all suppliers", ParentID: &suppliersModule.ID},
		{Name: "Create Suppliers", Path: fmt.Sprintf("%s/supplier", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new suppliers", ParentID: &suppliersModule.ID},
		{Name: "Restore Suppliers", Path: fmt.Sprintf("%s/supplier/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted suppliers", ParentID: &suppliersModule.ID},
		{Name: "Delete Suppliers", Path: fmt.Sprintf("%s/supplier/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete suppliers permanently", ParentID: &suppliersModule.ID},
		{Name: "Get Suppliers By ID", Path: fmt.Sprintf("%s/supplier/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get suppliers by ID", ParentID: &suppliersModule.ID},
		{Name: "Update Suppliers", Path: fmt.Sprintf("%s/supplier/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update suppliers by ID", ParentID: &suppliersModule.ID},
	}

	for _, sm := range suppliersServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Purchase Orders" module for service parent
	var purchaseOrdersModule models.Module
	if err := db.Where("name = ?", "Purchase Orders").First(&purchaseOrdersModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Purchase Orders module: %w", err)
	}

	// Service routes for purchase order management
	purchaseOrdersServiceModules := []models.Module{
		{Name: "Get All Paginated Purchase Orders", Path: fmt.Sprintf("%s/purchase-order", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get all paginated purchase orders", ParentID: &purchaseOrdersModule.ID},
		{Name: "Get All Purchase Orders", Path: fmt.Sprintf("%s/purchase-order/all", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all purchase orders", ParentID: &purchaseOrdersModule.ID},
		{Name: "Get Purchase Order By ID", Path: fmt.Sprintf("%s/purchase-order/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get purchase order by ID", ParentID: &purchaseOrdersModule.ID},
		{Name: "Create Purchase Order", Path: fmt.Sprintf("%s/purchase-order", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new purchase order", ParentID: &purchaseOrdersModule.ID},
		{Name: "Update Purchase Order", Path: fmt.Sprintf("%s/purchase-order/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update purchase order by ID", ParentID: &purchaseOrdersModule.ID},
		{Name: "Update Purchase Order Status", Path: fmt.Sprintf("%s/purchase-order/:id/status", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update purchase order status", ParentID: &purchaseOrdersModule.ID},
		{Name: "Receive Purchase Order Items", Path: fmt.Sprintf("%s/purchase-order/:id/receive", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Receive items for a purchase order", ParentID: &purchaseOrdersModule.ID},
		{Name: "Delete Purchase Orders", Path: fmt.Sprintf("%s/purchase-order/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete purchase orders permanently", ParentID: &purchaseOrdersModule.ID},
		{Name: "Restore Purchase Orders", Path: fmt.Sprintf("%s/purchase-order/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted purchase orders", ParentID: &purchaseOrdersModule.ID},
	}

	for _, sm := range purchaseOrdersServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Sales Orders" module for service parent
	var salesOrdersModule models.Module
	if err := db.Where("name = ?", "Sales Orders").First(&salesOrdersModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Sales Orders module: %w", err)
	}

	// Service routes for sales order management
	salesOrdersServiceModules := []models.Module{
		{Name: "Get All Paginated Sales Orders", Path: fmt.Sprintf("%s/sales-order", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get all paginated sales orders", ParentID: &salesOrdersModule.ID},
		{Name: "Get All Sales Orders", Path: fmt.Sprintf("%s/sales-order/all", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get list of all sales orders", ParentID: &salesOrdersModule.ID},
		{Name: "Get Sales Order By ID", Path: fmt.Sprintf("%s/sales-order/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get sales order by ID", ParentID: &salesOrdersModule.ID},
		{Name: "Create Sales Order", Path: fmt.Sprintf("%s/sales-order", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Create a new sales order", ParentID: &salesOrdersModule.ID},
		{Name: "Update Sales Order", Path: fmt.Sprintf("%s/sales-order/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update sales order by ID", ParentID: &salesOrdersModule.ID},
		{Name: "Update Sales Order Status", Path: fmt.Sprintf("%s/sales-order/:id/status", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Update sales order status", ParentID: &salesOrdersModule.ID},
		{Name: "Delete Sales Orders", Path: fmt.Sprintf("%s/sales-order/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete sales orders permanently", ParentID: &salesOrdersModule.ID},
		{Name: "Restore Sales Orders", Path: fmt.Sprintf("%s/sales-order/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted sales orders", ParentID: &salesOrdersModule.ID},
	}

	for _, sm := range salesOrdersServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Notifications" module for notification parent
	var notificationsModule models.Module
	if err := db.Where("name = ?", "Notifications").First(&notificationsModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Notifications module: %w", err)
	}

	// Service routes for notification management
	notificationsServiceModules := []models.Module{
		{Name: "Get All Notifications", Path: fmt.Sprintf("%s/notification", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get all notifications", ParentID: &notificationsModule.ID},
		{Name: "Mark All Notifications Read", Path: fmt.Sprintf("%s/notification/mark-all-read", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Mark all notifications as read", ParentID: &notificationsModule.ID},
		{Name: "Mark Multiple Notifications Read", Path: fmt.Sprintf("%s/notification/mark-multiple-read", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Mark multiple notifications as read", ParentID: &notificationsModule.ID},
		{Name: "Restore Notifications", Path: fmt.Sprintf("%s/notification/restore", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Restore soft-deleted notifications", ParentID: &notificationsModule.ID},
		{Name: "Delete Notifications", Path: fmt.Sprintf("%s/notification/delete", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Delete notifications permanently", ParentID: &notificationsModule.ID},
		{Name: "Mark Notification As Read", Path: fmt.Sprintf("%s/notification/:id/read", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Mark a notification as read by ID", ParentID: &notificationsModule.ID},
		{Name: "Get Notification By ID", Path: fmt.Sprintf("%s/notification/:id", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get a notification by ID", ParentID: &notificationsModule.ID},
	}

	for _, sm := range notificationsServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	// Get "Sales Reports" module for sales report parent
	var salesReportsModule models.Module
	if err := db.Where("name = ?", "Sales Reports").First(&salesReportsModule).Error; err != nil {
		return fmt.Errorf("failed to fetch Sales Reports module: %w", err)
	}

	// Service routes for sales report management
	salesReportsServiceModules := []models.Module{
		{Name: "Get All Sales Reports", Path: fmt.Sprintf("%s/sales-report", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get all sales reports", ParentID: &salesReportsModule.ID},
		{Name: "Get Summary Sales Report", Path: fmt.Sprintf("%s/sales-report/summary", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get summary sales report", ParentID: &salesReportsModule.ID},
		{Name: "Get Chart Sales Report", Path: fmt.Sprintf("%s/sales-report/charts", appVersion), ModuleTypeID: moduleTypeMap["Service API"], Description: "Get chart sales report", ParentID: &salesReportsModule.ID},
	}

	for _, sm := range salesReportsServiceModules {
		db.Where("name = ?", sm.Name).FirstOrCreate(&sm)
	}

	log.Println("Modules seeded successfully!")
	return nil
}

//
// ROLE-MODULE SEEDER
//

func SeedRoleModules(db *gorm.DB) error {
	log.Println("Seeding role-modules...")

	var modules []models.Module
	if err := db.Find(&modules).Error; err != nil {
		return fmt.Errorf("failed to get modules: %w", err)
	}

	if err := seedRoleModulesForRole(db, "DEVELOPER", modules); err != nil {
		return err
	}
	if err := seedRoleModulesForRole(db, "SUPERADMIN", modules); err != nil {
		return err
	}

	return nil
}

func seedRoleModulesForRole(db *gorm.DB, roleName string, modules []models.Module) error {
	var role models.Role
	if err := db.First(&role, "name = ?", roleName).Error; err != nil {
		return fmt.Errorf("failed to find role '%s': %w", roleName, err)
	}

	for _, module := range modules {
		var existing models.RoleModule
		err := db.Where("role_id = ? AND module_id = ?", role.ID, module.ID).First(&existing).Error
		if err == nil {
			continue
		}

		roleModule := models.RoleModule{
			ID:       uuid.New(),
			RoleID:   &role.ID,
			ModuleID: &module.ID,
			Checked:  true,
		}
		if err := db.Create(&roleModule).Error; err != nil {
			return fmt.Errorf("failed to create role-module for role '%s' module '%s': %w", roleName, module.Name, err)
		}
	}
	return nil
}

// 
// CATEGORY SEEDER
// 

func SeedCategories(db *gorm.DB) error {
	log.Println("Seeding categories...")

	categories := []models.Category{
		{ID: uuid.New(), Name: "Pharmaceuticals", Color: "#1f77b4", Description: "Medicines & pharmaceutical products"},
		{ID: uuid.New(), Name: "Personal Protective Equipment", Color: "#d62728", Description: "Masks, gloves, gowns, respirators"},
		{ID: uuid.New(), Name: "Medical Equipment", Color: "#9467bd", Description: "Durable medical equipment for hospitals/clinics"},
		{ID: uuid.New(), Name: "Diagnostics", Color: "#ff7f0e", Description: "Diagnostic tools & devices"},
		{ID: uuid.New(), Name: "Consumables", Color: "#2ca02c", Description: "Medical consumables & disposables"},
	}

	for _, category := range categories {
		if err := db.Create(&category).Error; err != nil {
			return fmt.Errorf("failed to seed category '%s': %w", category.Name, err)
		}
	}

	return nil
}


// 
// UOM SEEDER
// 


func SeedUoMs(db *gorm.DB) error {
	log.Println("Seeding UoMs...")

	type uomSeed struct {
		Name        string
		Color       string
		Description *string
	}

	desc := func(s string) *string { return &s }

	uoms := []uomSeed{
		{Name: "PCS",   Color: "#0ea5e9", Description: desc("Piece / unit satuan umum")},
		{Name: "BOX",   Color: "#22c55e", Description: desc("Kotak / box")},
		{Name: "PACK",  Color: "#84cc16", Description: desc("Kemasan / pack")},
		{Name: "PAIR",  Color: "#f97316", Description: desc("Sepasang (contoh sarung tangan)")},
		{Name: "BOTOL", Color: "#a855f7", Description: desc("Botol")},
		{Name: "LEMBAR",Color: "#eab308", Description: desc("Sheet / lembar")},
	}

	for _, s := range uoms {
		u := models.UoM{
			ID:          uuid.New(),
			Name:        s.Name,
			Color:       s.Color,
			Description: s.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"color", "description", "updated_at"}),
		}).Create(&u).Error; err != nil {
			return fmt.Errorf("seed uom '%s' failed: %w", s.Name, err)
		}
	}

	return nil
}


// 
// ITEM SEEDER
// 

func SeedItems(db *gorm.DB) error {
	log.Println("Seeding items (idempotent)…")

	loc := mustLoc("Asia/Jakarta")
	now := time.Now().In(loc)

	return db.Transaction(func(tx *gorm.DB) error {
		// ---------- refs ----------
		var uoms []models.UoM
		if err := tx.Find(&uoms).Error; err != nil {
			return fmt.Errorf("failed to load uoms: %w", err)
		}
		uomID := map[string]uuid.UUID{}
		for _, u := range uoms { uomID[u.Name] = u.ID }
		getUoM := func(name string) (uuid.UUID, error) {
			id, ok := uomID[name]; if !ok || id == uuid.Nil { return uuid.Nil, fmt.Errorf("UoM '%s' not found; seed UoMs first", name) }
			return id, nil
		}

		var cats []models.Category
		if err := tx.Find(&cats).Error; err != nil {
			return fmt.Errorf("failed to load categories: %w", err)
		}
		catID := map[string]uuid.UUID{}
		for _, c := range cats { catID[c.Name] = c.ID }
		getCat := func(name string) (uuid.UUID, error) {
			id, ok := catID[name]; if !ok || id == uuid.Nil { return uuid.Nil, fmt.Errorf("category '%s' not found; seed categories first", name) }
			return id, nil
		}

		var developerUser models.User
		if err := tx.First(&developerUser, "username = ?", "developer").Error; err != nil {
			return fmt.Errorf("failed to find user 'developer': %w", err)
		}

		// ---------- config ----------
		shelfLifeMonths := func(categoryName string) int {
			switch categoryName {
			case "Pharmaceuticals": return 24
			case "Consumables": return 12
			case "Personal Protective Equipment": return 18
			case "Diagnostics": return 36
			case "Medical Equipment": return 60
			default: return 24
			}
		}

		type itemSeed struct {
			Name, Code, CategoryName, UoMName, Description string
			Price, Stock, LowStock                          int
		}
		templates := []itemSeed{
			{"Paracetamol 500mg Tablets", "MED-PARA-500TAB", "Pharmaceuticals", "BOX",   "Analgesic/antipyretic", 1200, 250,  50},
			{"Amoxicillin 500mg Capsules","MED-AMOX-500CAP", "Pharmaceuticals", "BOX",   "Antibiotic",            2500, 150,  30},
			{"Syringe 5 mL",               "CON-SYR-05ML",   "Consumables",     "PCS",   "Disposable syringe 5 mL", 700, 500, 100},
			{"IV Cannula 22G",            "CON-IVC-22G",     "Consumables",     "PCS",   "Peripheral IV cannula 22G", 1800,200,40},
			{"Surgical Gloves (Latex) - L (pair)", "PPE-GLV-L", "Personal Protective Equipment", "PAIR", "Latex gloves size L", 1500, 400, 80},
			{"3-Ply Surgical Mask",       "PPE-MSK-3PLY",    "Personal Protective Equipment", "BOX",  "Disposable 3-ply mask", 500, 1000, 200},
			{"N95 Respirator",            "PPE-N95",         "Personal Protective Equipment", "PCS",  "N95 respirator mask", 12000, 300, 60},
			{"Digital Thermometer",       "DIA-THERM-DIG",   "Diagnostics",     "PCS",   "Digital body thermometer", 75000, 40, 10},
			{"Blood Pressure Monitor (Auto)", "DIA-BP-AUTO", "Diagnostics",     "PCS",   "Automatic BP monitor", 550000, 15, 5},
			{"Infusion Pump",             "EQP-INF-PUMP",    "Medical Equipment","PCS",  "Volumetric infusion pump", 12500000, 5, 2},
			{"Ultrasound Gel 5L",         "CON-USG-5L",      "Consumables",     "BOTOL", "Ultrasound transmission gel 5L", 95000, 25, 5},
			{"Hand Sanitizer 500ml",      "CON-HS-500",      "Consumables",     "BOTOL", "Alcohol-based hand rub 500ml", 20000, 300, 50},
		}

		// ---------- tentukan yang SUDAH ADA (by code), dan bangun daftar yang AKAN DIBUAT ----------
		// siapkan semua code calon insert
		type plan struct {
			T itemSeed
			Batch int
			Code  string
		}
		var plans []plan
		var allCodes []string

		for _, t := range templates {
			for b := 1; b <= 3; b++ {
				c := fmt.Sprintf("%s-B%02d", t.Code, b)
				allCodes = append(allCodes, c)
				plans = append(plans, plan{T: t, Batch: b, Code: c})
			}
		}

		// ambil code yang sudah ada
		var existing []struct{ Code string }
		if err := tx.Model(&models.Item{}).Select("code").Where("code IN ?", allCodes).Find(&existing).Error; err != nil {
			return fmt.Errorf("check existing codes: %w", err)
		}
		exists := map[string]struct{}{}
		for _, e := range existing { exists[e.Code] = struct{}{} }

		// filter hanya yang belum ada
		var missing []plan
		for _, p := range plans {
			if _, ok := exists[p.Code]; !ok {
				missing = append(missing, p)
			}
		}

		if len(missing) == 0 {
			log.Println("SeedItems: nothing to insert (all present).")
			return nil
		}

		// ---------- siapkan timeline sejumlah missing (elemen terakhir = now) ----------
		buildTimeline := func(n int) []time.Time {
			ts := make([]time.Time, 0, n)
			for k := n - 1; k >= 0; k-- {
				t := now.AddDate(0, -k, 0) // k bulan lalu s/d sekarang
				yr, mo := t.Year(), t.Month()
				day := min(28, 1+rand.Intn(28)) // biar aman di semua bulan
				h, m, s := randHourMinute()
				tt := time.Date(yr, mo, day, h, m, s, 0, loc)
				if k == 0 { tt = now } // terakhir pasti now
				ts = append(ts, tt)
			}
			return ts
		}
		timeline := buildTimeline(len(missing))

		// ---------- insert missing ----------
		for i, p := range missing {
			createdAt := timeline[i]
			updatedAt := createdAt.Add(time.Duration(rand.Intn(20*24)) * time.Hour)
			if updatedAt.After(now) { updatedAt = createdAt }
			b := p.Batch

			cid, err := getCat(p.T.CategoryName)
			if err != nil { return err }
			uid, err := getUoM(p.T.UoMName)
			if err != nil { return err }

			baseMonths := shelfLifeMonths(p.T.CategoryName)
			if baseMonths <= 0 { baseMonths = 24 }
			monthsToAdd := baseMonths - (b-1)*6
			if monthsToAdd < 3 { monthsToAdd = 3 }
			exp := createdAt.AddDate(0, monthsToAdd, 0)

			name := fmt.Sprintf("%s (B%02d)", p.T.Name, b) // unik by name juga
			item := models.Item{
				ID:          uuid.New(),
				Name:        name,
				Code:        p.Code,
				CategoryID:  cid,
				UoMID:       uid,
				Price:       p.T.Price,
				Stock:       p.T.Stock,
				LowStock:    p.T.LowStock,
				Description: p.T.Description,
				Batch:       b,
				ExpiredAt:   exp,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			}
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("create item '%s': %w", item.Code, err)
			}

			// histories pakai createdAt
			hPrice := models.ItemHistory{
				ID:           uuid.New(),
				ItemID:       item.ID,
				ChangeType:   "create_price",
				Description:  fmt.Sprintf("Initial price set to %d (B%02d)", item.Price, b),
				OldPrice:     0,
				NewPrice:     item.Price,
				CurrentPrice: item.Price,
				OldStock:     0,
				NewStock:     0,
				CurrentStock: 0,
				CreatedBy:    &developerUser.ID,
				UpdatedBy:    &developerUser.ID,
				CreatedAt:    createdAt,
				UpdatedAt:    createdAt,
			}
			if err := tx.Create(&hPrice).Error; err != nil {
				return fmt.Errorf("price history '%s': %w", item.Code, err)
			}

			hStock := models.ItemHistory{
				ID:           uuid.New(),
				ItemID:       item.ID,
				ChangeType:   "create_stock",
				Description:  fmt.Sprintf("Initial stock set to %d (B%02d)", item.Stock, b),
				OldStock:     0,
				NewStock:     item.Stock,
				CurrentStock: item.Stock,
				OldPrice:     0,
				NewPrice:     0,
				CurrentPrice: 0,
				CreatedBy:    &developerUser.ID,
				UpdatedBy:    &developerUser.ID,
				CreatedAt:    createdAt,
				UpdatedAt:    createdAt,
			}
			if err := tx.Create(&hStock).Error; err != nil {
				return fmt.Errorf("stock history '%s': %w", item.Code, err)
			}
		}

		log.Printf("SeedItems: inserted %d new items (last at now()).\n", len(missing))
		return nil
	})
}

// ======================================================================
// SEED AREAS
// ======================================================================
func SeedAreas(db *gorm.DB) error {
	log.Println("Seeding areas...")

	areas := []models.Area{
		{
			ID:        uuid.New(),
			Code:      "AR-JKT-CTR",
			Name:      "DKI Jakarta - Central",
			Color:     "#2563eb",
			Latitude:  fptr(-6.1818),
			Longitude: fptr(106.8283),
		},
		{
			ID:        uuid.New(),
			Code:      "AR-JKT-SOU",
			Name:      "DKI Jakarta - South",
			Color:     "#16a34a",
			Latitude:  fptr(-6.2626),
			Longitude: fptr(106.8103),
		},
		{
			ID:        uuid.New(),
			Code:      "AR-JKT-EAS",
			Name:      "DKI Jakarta - East",
			Color:     "#f59e0b",
			Latitude:  fptr(-6.2250),
			Longitude: fptr(106.9006),
		},
		{
			ID:        uuid.New(),
			Code:      "AR-JKT-WES",
			Name:      "DKI Jakarta - West",
			Color:     "#ef4444",
			Latitude:  fptr(-6.1683),
			Longitude: fptr(106.7586),
		},
		{
			ID:        uuid.New(),
			Code:      "AR-JKT-NOR",
			Name:      "DKI Jakarta - North",
			Color:     "#8b5cf6",
			Latitude:  fptr(-6.1256),
			Longitude: fptr(106.8470),
		},
	}

	for _, a := range areas {
		ar := a
		if err := db.Create(&ar).Error; err != nil {
			return fmt.Errorf("failed to create area '%s': %w", ar.Name, err)
		}
	}
	return nil
}

// ======================================================================
// SEED FACILITIES (beserta CustomerType)
// ======================================================================
func SeedCustomers(db *gorm.DB) error {
	log.Println("Seeding customer types and customers...")

	ftSeeds := []struct {
		ID          uuid.UUID
		Name, Color string
		Desc        *string
	}{
		{uuid.New(), "Hospital", "#2563eb", strptr("General & specialized hospitals")},
		{uuid.New(), "Clinic", "#16a34a", strptr("Primary/secondary care clinics")},
		{uuid.New(), "Pharmacy", "#f59e0b", strptr("Retail & in-hospital pharmacies")},
		{uuid.New(), "Laboratory", "#8b5cf6", strptr("Medical diagnostics laboratories")},
		{uuid.New(), "Puskesmas", "#ef4444", strptr("Community health centers")},
		{uuid.New(), "Medical Supply Warehouse", "#0ea5e9", strptr("Central medical supply warehouse")},
		{uuid.New(), "Ambulance Station", "#64748b", strptr("Ambulance & EMS base station")},
	}

	for _, s := range ftSeeds {
		ft := models.CustomerType{
			ID:          s.ID,
			Name:        s.Name,
			Color:       s.Color,
			Description: s.Desc,
		}
		if err := db.Create(&ft).Error; err != nil {
			return fmt.Errorf("failed to create customer type '%s': %w", s.Name, err)
		}
	}

	var types []models.CustomerType
	if err := db.Find(&types).Error; err != nil {
		return fmt.Errorf("failed to load customer types: %w", err)
	}
	typeID := map[string]uuid.UUID{}
	for _, t := range types {
		typeID[t.Name] = t.ID
	}

	var areas []models.Area
	if err := db.Find(&areas).Error; err != nil {
		return fmt.Errorf("failed to load areas: %w", err)
	}
	areaID := map[string]uuid.UUID{}
	for _, a := range areas {
		areaID[a.Code] = a.ID
	}
	requireArea := func(code string) (uuid.UUID, error) {
		id, ok := areaID[code]
		if !ok || id == uuid.Nil {
			return uuid.Nil, fmt.Errorf("area with code '%s' not found (run SeedAreas first)", code)
		}
		return id, nil
	}

	facSeeds := []struct {
		Nomor, Name string
		TypeName   string
		AreaCode   string
		Address    *string
		Phone      *string
		Email      *string
		Lat, Lng   *float64
	}{
		{"FAC-RSUD-JKT-CTR", "RSUD Jakarta Pusat", "Hospital", "AR-JKT-CTR",
			strptr("Jl. Kesehatan No. 10, Jakarta Pusat"), strptr("+62-21-1234567"), strptr("rsud.ctr@jakarta.go.id"),
			fptr(-6.1723), fptr(106.8311)},
		{"FAC-RS-FATMAWATI", "RSUP Fatmawati", "Hospital", "AR-JKT-SOU",
			strptr("Jl. RS Fatmawati No. 4, Jakarta Selatan"), strptr("+62-21-7654321"), strptr("info@rsfatmawati.id"),
			fptr(-6.2922), fptr(106.7973)},
		{"FAC-KLINIK-SEHAT-1", "Klinik Sehat Medika", "Clinic", "AR-JKT-EAS",
			strptr("Jl. Pahlawan No. 21, Jakarta Timur"), strptr("+62-21-22223333"), strptr("cs@sehatmedika.co.id"),
			fptr(-6.2189), fptr(106.9062)},
		{"FAC-APOTEK-SENTOSA", "Apotek Sehat Sentosa", "Pharmacy", "AR-JKT-WES",
			strptr("Jl. Meruya Raya No. 88, Jakarta Barat"), strptr("+62-21-88997766"), strptr("halo@sentosa-apotek.id"),
			fptr(-6.2007), fptr(106.7508)},
		{"FAC-LAB-NUSANTARA", "Lab Diagnostik Nusantara", "Laboratory", "AR-JKT-NOR",
			strptr("Jl. Gunung Sahari No. 45, Jakarta Utara"), strptr("+62-21-99887766"), strptr("contact@labnusantara.id"),
			fptr(-6.1379), fptr(106.8367)},
		{"FAC-PUSKESMAS-TEBET", "Puskesmas Kecamatan Tebet", "Puskesmas", "AR-JKT-SOU",
			strptr("Jl. Tebet Raya No. 3, Jakarta Selatan"), strptr("+62-21-7778889"), strptr("puskesmas.tebet@jakarta.go.id"),
			fptr(-6.2373), fptr(106.8541)},
		{"FAC-GDG-FARMASI-DKI", "Gudang Farmasi DKI", "Medical Supply Warehouse", "AR-JKT-EAS",
			strptr("Jl. Industri No. 12, Jakarta Timur"), strptr("+62-21-66007788"), strptr("gudangfarmasi@jakarta.go.id"),
			fptr(-6.2104), fptr(106.9152)},
		{"FAC-AMBULANS-HARMONI", "Pos Ambulans Harmoni", "Ambulance Station", "AR-JKT-CTR",
			strptr("Jl. Gajah Mada No. 1, Jakarta Pusat"), strptr("+62-21-119"), strptr("dispatch@ems-jkt.id"),
			fptr(-6.1646), fptr(106.8210)},
	}

	for _, s := range facSeeds {
		ftID, ok := typeID[s.TypeName]
		if !ok || ftID == uuid.Nil {
			return fmt.Errorf("customer type '%s' not found", s.TypeName)
		}
		arID, err := requireArea(s.AreaCode)
		if err != nil {
			return err
		}

		f := models.Customer{
			ID:             uuid.New(),
			Nomor:           s.Nomor,
			Name:           s.Name,
			CustomerTypeID: ftID,
			AreaID:         arID,
			Address:        s.Address,
			Phone:          s.Phone,
			Email:          s.Email,
			Latitude:       s.Lat,
			Longitude:      s.Lng,
		}
		if err := db.Create(&f).Error; err != nil {
			return fmt.Errorf("failed to create customer '%s': %w", f.Name, err)
		}
	}

	return nil
}

// ======================================================================
// SEED SALES PERSONS (+ assignments)
// ======================================================================
func SeedSalesPersons(db *gorm.DB) error {
	log.Println("Seeding sales persons (and assignments)...")

	loc, _ := time.LoadLocation("Asia/Jakarta")

	type spSeed struct {
		Name     string
		Phone    *string
		Email    *string
		HireDate *time.Time
		Address  *string
		NPWP     *string
	}
	seeds := []spSeed{
		{"Andi Pratama",  strptr("+62-811-1000-001"), strptr("andi.pratama@company.id"),  tptr(loc, 2023, time.January, 9),  strptr("Jl. Melati No. 12, Jakarta Selatan"), nil},
		{"Siti Rahma",    strptr("+62-811-1000-002"), strptr("siti.rahma@company.id"),    tptr(loc, 2023, time.March, 20),   strptr("Jl. Mawar No. 8, Jakarta Timur"), nil},
		{"Budi Santoso",  strptr("+62-811-1000-003"), strptr("budi.santoso@company.id"),  tptr(loc, 2022, time.November, 1), strptr("Jl. Cendana No. 5, Jakarta Barat"), nil},
		{"Dewi Lestari",  strptr("+62-811-1000-004"), strptr("dewi.lestari@company.id"),  tptr(loc, 2024, time.May, 6),      strptr("Jl. Kenanga No. 33, Jakarta Pusat"), nil},
		{"Fajar Nugroho", strptr("+62-811-1000-005"), strptr("fajar.nugroho@company.id"), tptr(loc, 2024, time.February, 12),strptr("Jl. Dahlia No. 2, Jakarta Utara"), nil},
		{"Gita Putri",    strptr("+62-811-1000-006"), strptr("gita.putri@company.id"),    tptr(loc, 2025, time.January, 13), strptr("Jl. Flamboyan No. 17, Depok"), nil},
	}

	spID := map[string]uuid.UUID{}
	for _, s := range seeds {
		sp := models.SalesPerson{
			ID:        uuid.New(),
			Name:      s.Name,
			Phone:     s.Phone,
			Email:     s.Email,
			HireDate:  s.HireDate,
			Address:   s.Address,
		}
		if err := db.Create(&sp).Error; err != nil {
			return fmt.Errorf("failed to create sales person '%s': %w", sp.Name, err)
		}
		spID[sp.Name] = sp.ID
	}

	var areas []models.Area
	if err := db.Find(&areas).Error; err != nil {
		return fmt.Errorf("failed to load areas: %w", err)
	}
	areaIDByCode := map[string]uuid.UUID{}
	for _, a := range areas {
		areaIDByCode[a.Code] = a.ID
	}
	requireArea := func(code string) (uuid.UUID, error) {
		id, ok := areaIDByCode[code]
		if !ok || id == uuid.Nil {
			return uuid.Nil, fmt.Errorf("area with code '%s' not found (run SeedAreas first)", code)
		}
		return id, nil
	}

	assignPlan := map[string][]string{
		"Andi Pratama":  {"AR-JKT-CTR", "AR-JKT-WES"},
		"Siti Rahma":    {"AR-JKT-SOU", "AR-JKT-EAS"},
		"Budi Santoso":  {"AR-JKT-NOR"},
		"Dewi Lestari":  {"AR-JKT-CTR", "AR-JKT-SOU"},
		"Fajar Nugroho": {"AR-JKT-EAS"},
		"Gita Putri":    {"AR-JKT-WES", "AR-JKT-NOR"},
	}

	for name, codes := range assignPlan {
		salesID, ok := spID[name]
		if !ok {
			return fmt.Errorf("sales person '%s' not found after seeding", name)
		}
		for _, code := range codes {
			areaID, err := requireArea(code)
			if err != nil {
				return err
			}
			asg := models.SalesAssignment{
				ID:            uuid.New(),
				SalesPersonID: salesID,
				AreaID:        areaID,
				Checked:       true, 
			}
			if err := db.Create(&asg).Error; err != nil {
				return fmt.Errorf("failed to create assignment (%s -> %s): %w", name, code, err)
			}
		}
	}

	return nil
}

// ======================================================================
// SEED SUPPLIERS
// ======================================================================
func SeedSuppliers(db *gorm.DB) error {
	log.Println("Seeding suppliers...")

	type supSeed struct {
		Name          string
		Code          string
		Email         *string
		Phone         *string
		Address       *string
		ContactPerson *string
	}

	suppliers := []supSeed{
		{"Medika Nusantara", "SUP-MEDNUS", strptr("sales@medikanusantara.id"), strptr("+62-21-5550101"), strptr("Jl. Industri No. 10, Jakarta Timur"), strptr("Rina Widjaja")},
		{"Sehat Abadi Pharma", "SUP-SEHAB", strptr("cs@sehatabadi.co.id"), strptr("+62-21-5550102"), strptr("Jl. Rawa Terate No. 22, Jakarta Timur"), strptr("Andreas T")},
		{"Prima Diagnostik", "SUP-PRIMDIA", strptr("info@primadiagnostik.id"), strptr("+62-21-5550103"), strptr("Jl. Gading Kirana No. 5, Jakarta Utara"), strptr("Rudi Hartono")},
		{"Alkes Sentosa", "SUP-ALKESEN", strptr("order@alkessentosa.id"), strptr("+62-21-5550104"), strptr("Jl. Tomang Raya No. 45, Jakarta Barat"), strptr("Maya Sari")},
		{"Glove & Mask Indonesia", "SUP-GMI", strptr("hello@gmi.co.id"), strptr("+62-21-5550105"), strptr("Jl. Panjang No. 88, Jakarta Barat"), strptr("Deni K")},
		{"Nusantara Equipments", "SUP-NEQ", strptr("contact@neq.id"), strptr("+62-21-5550106"), strptr("Jl. TB Simatupang No. 120, Jakarta Selatan"), strptr("Sonia Putra")},
		{"Farmasi Maju Jaya", "SUP-FMJ", strptr("support@fmajujaya.id"), strptr("+62-21-5550107"), strptr("Jl. Raya Bekasi KM 21, Bekasi"), strptr("Bayu W")},
		{"LabTech Solutions", "SUP-LABTECH", strptr("admin@labtech.co.id"), strptr("+62-21-5550108"), strptr("Jl. Cikini Raya No. 7, Jakarta Pusat"), strptr("Selvi Anggraini")},
	}

	for _, s := range suppliers {
		sp := models.Supplier{
			ID:            uuid.New(),
			Name:          s.Name,
			Code:          s.Code,
			Email:         s.Email,
			Phone:         s.Phone,
			Address:       s.Address,
			ContactPerson: s.ContactPerson,
		}
		if err := db.Create(&sp).Error; err != nil {
			return fmt.Errorf("failed to create supplier '%s': %w", sp.Name, err)
		}
	}
	return nil
}

// ======================================================================
// SEED SALES ORDERS
// ======================================================================
func mustLoc(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}

func randHourMinute() (int, int, int) {
	return 9 + rand.Intn(8), rand.Intn(60), rand.Intn(60)
}

func lastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

type spCustPair struct {
	SP  models.SalesPerson
	Cust models.Customer
}

func buildValidPairs(sps []models.SalesPerson, facs []models.Customer) []spCustPair {
	var pairs []spCustPair
	for _, sp := range sps {
		areaIDs := map[uuid.UUID]struct{}{}
		for _, as := range sp.Assignments {
			areaIDs[as.AreaID] = struct{}{}
		}
		if len(areaIDs) == 0 {
			continue
		}
		for _, f := range facs {
			if _, ok := areaIDs[f.AreaID]; ok {
				pairs = append(pairs, spCustPair{SP: sp, Cust: f})
			}
		}
	}
	return pairs
}

func SeedSalesOrders(db *gorm.DB) error {
	log.Println("Seeding sales orders")

	loc := mustLoc("Asia/Jakarta")
	today := time.Now().In(loc)

	start := time.Date(today.Year()-1, time.January, 1, 12, 0, 0, 0, loc)
	end := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 0, 0, loc)

	var salesPersons []models.SalesPerson
	if err := db.Preload("Assignments").Find(&salesPersons).Error; err != nil {
		return err
	}
	if len(salesPersons) == 0 {
		return fmt.Errorf("no sales persons found, seed sales persons first")
	}

	var customers []models.Customer
	if err := db.Find(&customers).Error; err != nil {
		return err
	}
	if len(customers) == 0 {
		return fmt.Errorf("no customers found, seed customers first")
	}

	var items []models.Item
	if err := db.Find(&items).Error; err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("no items found, seed items first")
	}

	pairs := buildValidPairs(salesPersons, customers)
	if len(pairs) == 0 {
		return fmt.Errorf("no valid SalesPerson–Customer pairs (area mismatch)")
	}

	statusOptions := []string{"Draft", "Confirmed", "Shipped", "Delivered", "Closed"}
	paymentOptions := []string{"Unpaid", "Partial", "Paid"}
	termOptions := []string{"Full", "DP", "Tempo"}

	statusIdx, payIdx, termIdx := 0, 0, 0
	pairIdx, itemIdx := 0, 0

	var existingCount int64
	if err := db.Model(&models.SalesOrder{}).Count(&existingCount).Error; err != nil {
		return err
	}
	soCounter := int(existingCount) + 1
	const ordersPerDay = 1 

	for year, month := start.Year(), start.Month(); !time.Date(year, month, 1, 0, 0, 0, 0, loc).After(end); {
		lastDay := lastDayOfMonth(year, month, loc)
		if year == today.Year() && month == today.Month() {
			lastDay = today.Day()
		}

		for day := 1; day <= lastDay; day++ {
			hh, mm, _ := randHourMinute()
			soDate := time.Date(year, month, day, hh, mm, rand.Intn(60), 0, loc)
			createdAt := soDate

			for k := 0; k < ordersPerDay; k++ {
				term := termOptions[termIdx%len(termOptions)]
				soStatus := statusOptions[statusIdx%len(statusOptions)]
				payStatus := paymentOptions[payIdx%len(paymentOptions)]
				termIdx++
				statusIdx++
				payIdx++

				p := pairs[pairIdx%len(pairs)]
				pairIdx++

				estArrival := soDate.AddDate(0, 0, 3+rand.Intn(10)) 
				dueDate := soDate.AddDate(0, 0, 14+rand.Intn(21)) 

				if err := db.Transaction(func(tx *gorm.DB) error {
					nItems := 3 + rand.Intn(4)
					totalAmount := 0
					chosenItems := make([]models.SalesOrderItem, 0, nItems)
					for j := 0; j < nItems; j++ {
						it := items[itemIdx%len(items)]
						itemIdx++

						qty := 1 + rand.Intn(5)
						unitPrice := 50_000 + rand.Intn(300_000)
						lineTotal := qty * unitPrice
						totalAmount += lineTotal

						chosenItems = append(chosenItems, models.SalesOrderItem{
							ID:         uuid.New(),
							ItemID:     it.ID,
							UoMID:      it.UoMID,
							Quantity:   qty,
							UnitPrice:  unitPrice,
							TotalPrice: lineTotal,
							CreatedAt:  createdAt,
							UpdatedAt:  createdAt,
						})
					}
					if totalAmount <= 0 {
						totalAmount = 1
					}

					dpAmount := 0
					if term == "DP" {
						dpAmount = int(float64(totalAmount) * 0.3) 
						if dpAmount <= 0 {
							dpAmount = 1
						}
					}

					paidAmount := 0
					switch payStatus {
					case "Unpaid":
						paidAmount = 0
					case "Partial":
						if term == "DP" {
							paidAmount = dpAmount
						} else {
							paidAmount = int(float64(totalAmount) * 0.5)
						}
						if paidAmount <= 0 {
							paidAmount = 1
						}
						if paidAmount >= totalAmount {
							paidAmount = totalAmount - 1
						}
					case "Paid":
						paidAmount = totalAmount
					}

					soID := uuid.New()
					so := models.SalesOrder{
						ID:               soID,
						SONumber:         fmt.Sprintf("SO-%06d", soCounter),
						SalesPersonID:    p.SP.ID,
						CustomerID:       p.Cust.ID,
						SODate:           soDate,
						EstimatedArrival: &estArrival,
						TermOfPayment:    term,
						SOStatus:         soStatus,
						PaymentStatus:    payStatus,
						TotalAmount:      totalAmount,
						PaidAmount:       paidAmount,
						DPAmount:         dpAmount,
						DueDate:          &dueDate,
						Notes:            fmt.Sprintf("Seed %s -> %s (%s/%s/%s)", p.SP.Name, p.Cust.Name, term, soStatus, payStatus),
						CreatedAt:        createdAt,
						UpdatedAt:        createdAt,
					}
					if err := tx.Create(&so).Error; err != nil {
						return err
					}

					for idx := range chosenItems {
						chosenItems[idx].SalesOrderID = soID
					}
					if err := tx.Create(&chosenItems).Error; err != nil {
						return err
					}

					if paidAmount > 0 {
						payDate := soDate.AddDate(0, 0, 1+rand.Intn(10))
						if payDate.After(end) {
							payDate = end 
						}
						payment := models.Payment{
							ID:              uuid.New(),
							OrderType:       "SO",
							SalesOrderID:    &soID,
							PaymentType:     term, 
							Amount:          paidAmount,
							PaymentDate:     payDate,
							PaymentMethod:   "Transfer",
							ReferenceNumber: fmt.Sprintf("PAY-%06d", soCounter),
							Notes:           "Seed payment",
							CreatedAt:       createdAt,
							UpdatedAt:       createdAt,
						}
						if err := tx.Create(&payment).Error; err != nil {
							return err
						}
					}

					return nil
				}); err != nil {
					return err
				}

				soCounter++
			}
		}

		if month == time.December {
			year++
			month = time.January
		} else {
			month++
		}
		if time.Date(year, month, 1, 0, 0, 0, 0, loc).After(end) {
			break
		}
	}

	log.Println("Seeding sales orders completed")
	return nil
}

//
// USER SEEDER
//

func SeedUsers(db *gorm.DB) error {
	log.Println("Seeding users...")

	var developerRole, superadminRole, salesRole models.Role

	if err := db.First(&developerRole, "name = ?", "DEVELOPER").Error; err != nil {
		return fmt.Errorf("failed to find role 'DEVELOPER': %w", err)
	}
	if err := db.First(&superadminRole, "name = ?", "SUPERADMIN").Error; err != nil {
		return fmt.Errorf("failed to find role 'SUPERADMIN': %w", err)
	}
	if err := db.First(&salesRole, "name = ?", "SALES").Error; err != nil {
		return fmt.Errorf("failed to find role 'SALES': %w", err)
	}

	hashedPassword, err := helpers.HashPassword("password")
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	users := []models.User{
		{Username: "developer", Password: hashedPassword, Name: "Developer", Email: "dev@example.com", RoleID: &developerRole.ID},
		{Username: "superadmin", Password: hashedPassword, Name: "Superadmin", Email: "superadmin@example.com", RoleID: &superadminRole.ID},
		{Username: "adit", Password: hashedPassword, Name: "Adit", Email: "adit@example.com", RoleID: &superadminRole.ID},
		{Username: "sales", Password: hashedPassword, Name: "Sales", Email: "sales@example.com", RoleID: &salesRole.ID},
	}

	for _, user := range users {
		user.ID = uuid.New()
		if err := db.FirstOrCreate(&user, models.User{Username: user.Username}).Error; err != nil {
			return fmt.Errorf("failed to seed user '%s': %w", user.Username, err)
		}
	}
	return nil
}
