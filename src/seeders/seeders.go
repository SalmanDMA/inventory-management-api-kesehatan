package seeders

import (
	"fmt"
	"log"

	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//
// ROLE SEEDER
//

func SeedRoles(db *gorm.DB) error {
	roles := []models.Role{
		{ID: uuid.New(), Name: "SUPERADMIN", Alias: "SA", Color: "#f00f00", Description: "Akun Super Admin"},
		{ID: uuid.New(), Name: "DEVELOPER", Alias: "DEV", Color: "#000000", Description: "Akun Developer"},
	}

	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			return fmt.Errorf("failed to seed role '%s': %w", role.Name, err)
		}
	}
	return nil
}

//
// MODULE SEEDER
//

func SeedModules(db *gorm.DB) error {
	// Seed Module Types
	moduleTypes := []models.ModuleType{
		{ID: uuid.New(), Name: "Menu Directory", Description: "Directory menu", Icon: "mdi:folder"},
		{ID: uuid.New(), Name: "Route Menu", Description: "Route-based menu", Icon: "mdi:routes"},
		{ID: uuid.New(), Name: "Route Hidden", Description: "Hidden routes not displayed in frontend", Icon: "mdi:eye-off"},
		{ID: uuid.New(), Name: "Service API", Description: "CRUD operations and API endpoints", Icon: "mdi:server"},
	}

	for _, mt := range moduleTypes {
		db.Where("name = ?", mt.Name).FirstOrCreate(&mt)
	}

	// Get ModuleType IDs
	moduleTypeMap := make(map[string]uuid.UUID)
	for _, name := range []string{"Menu Directory", "Route Menu", "Service API", "Route Hidden"} {
		var mt models.ModuleType
		if err := db.Where("name = ?", name).First(&mt).Error; err != nil {
			return fmt.Errorf("failed to fetch module type '%s': %w", name, err)
		}
		moduleTypeMap[name] = mt.ID
	}

	// Create parent: Access Control
	var accessControlModule models.Module
	db.Where("name = ?", "Access Control").FirstOrCreate(&accessControlModule, models.Module{
		Name:         "Access Control",
		Route:        "/dashboard/settings",
		Icon:         "mdi:shield",
		ModuleTypeID: moduleTypeMap["Menu Directory"],
		Description:  "Access Control Page",
	})

	// Create parent: Master Data
	var masterDataModule models.Module
	db.Where("name = ?", "Master Data").FirstOrCreate(&masterDataModule, models.Module{
		Name:         "Master Data",
		Icon:         "mdi:database",
		ModuleTypeID: moduleTypeMap["Menu Directory"],
	})

	// Define modules
	modules := []models.Module{
		{
			Name:         "Dashboard",
			Route:        "/dashboard",
			Icon:         "mdi:home",
			ModuleTypeID: moduleTypeMap["Menu Directory"],
			Description:  "Main Dashboard",
		},
		{
			Name:         "Settings",
			Route:        "/dashboard/settings",
			Icon:         "mdi:settings",
			ModuleTypeID: moduleTypeMap["Menu Directory"],
			Description:  "Settings Page",
		},
		{
			Name:         "Users",
			Route:        "/dashboard/users",
			Icon:         "mdi:account",
			ModuleTypeID: moduleTypeMap["Route Menu"],
			ParentID:     &accessControlModule.ID,
			Description:  "User Management Page",
		},
		{
			Name:         "Roles",
			Route:        "/dashboard/roles",
			Icon:         "mdi:account-group",
			ModuleTypeID: moduleTypeMap["Route Menu"],
			ParentID:     &accessControlModule.ID,
			Description:  "Role Management Page",
		},
		{
			Name:         "Modules",
			Route:        "/dashboard/modules",
			Icon:         "mdi:layers-triple",
			ModuleTypeID: moduleTypeMap["Route Menu"],
			ParentID:     &accessControlModule.ID,
			Description:  "Module Management Page",
		},
		{
			Name:         "Items",
			Route:        "/dashboard/items",
			Icon:         "mdi:inboxes",
			ModuleTypeID: moduleTypeMap["Route Menu"],
			ParentID:     &masterDataModule.ID,
		},
		{
			Name:         "Categories",
			Route:        "/dashboard/categories",
			Icon:         "mdi:category",
			ModuleTypeID: moduleTypeMap["Route Menu"],
			ParentID:     &masterDataModule.ID,
		},
	}

	for _, m := range modules {
		var existing models.Module
		db.Where("name = ?", m.Name).FirstOrCreate(&existing, m)
	}

	return nil
}

//
// ROLE-MODULE SEEDER
//

func SeedRoleModules(db *gorm.DB) error {
	var role models.Role
	if err := db.First(&role, "name = ?", "DEVELOPER").Error; err != nil {
		return fmt.Errorf("failed to find role 'DEVELOPER': %w", err)
	}

	var modules []models.Module
	if err := db.Find(&modules).Error; err != nil {
		return fmt.Errorf("failed to get modules: %w", err)
	}

	for _, module := range modules {
		var existing models.RoleModule
		if err := db.Where("role_id = ? AND module_id = ?", role.ID, module.ID).First(&existing).Error; err == nil {
			continue
		}

		roleModule := models.RoleModule{
			ID:       uuid.New(),
			RoleID:   &role.ID,
			ModuleID: &module.ID,
			Checked:  true,
		}
		if err := db.Create(&roleModule).Error; err != nil {
			return fmt.Errorf("failed to create role-module for module '%s': %w", module.Name, err)
		}
	}

	return nil
}


//
// USER SEEDER
//

func SeedUsers(db *gorm.DB) error {
	var (
		developerRole   models.Role
		superadminRole  models.Role
	)

	if err := db.First(&developerRole, "name = ?", "DEVELOPER").Error; err != nil {
		return fmt.Errorf("failed to find role 'DEVELOPER': %w", err)
	}

	if err := db.First(&superadminRole, "name = ?", "SUPERADMIN").Error; err != nil {
		return fmt.Errorf("failed to find role 'SUPERADMIN': %w", err)
	}

	// Hash password
	hashedPassword, err := helpers.HashPassword("password")
	if err != nil {
		log.Printf("error hashing password: %v", err)
		return err
	}

	users := []models.User{
		{
			ID:       uuid.New(),
			Username: "devuser",
			Password: hashedPassword,
			Name:     "Developer",
			Email:    "dev@example.com",
			RoleID:   &developerRole.ID,
		},
		{
			ID:       uuid.New(),
			Username: "superadmin",
			Password: hashedPassword,
			Name:     "Superadmin",
			Email:    "superadmin@example.com",
			RoleID:   &superadminRole.ID,
		},
	}

	for _, user := range users {
		if err := db.FirstOrCreate(&user, models.User{Username: user.Username}).Error; err != nil {
			return fmt.Errorf("failed to seed user '%s': %w", user.Username, err)
		}
	}
	return nil
}
