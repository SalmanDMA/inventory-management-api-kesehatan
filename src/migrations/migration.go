package migrations

import (
	"fmt"
	"log"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/seeders"
)

func RunMigration() {
	err := configs.DB.AutoMigrate(
		&models.Upload{},
		&models.User{},
		&models.Role{},
		&models.ModuleType{},
		&models.Module{},
		&models.RoleModule{},
		&models.Category{},
		&models.Item{},
		&models.ItemHistory{},
	)
	
	var count int64

	configs.DB.Model(&models.Role{}).Count(&count)
	if count == 0 {
		if err := seeders.SeedRoles(configs.DB); err != nil {
			fmt.Println("Seeding roles failed:", err)
		} else {
			fmt.Println("Seeding roles successful")
		}
	} else {
		fmt.Println("Roles are already seeded")
	}

	configs.DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		if err := seeders.SeedUsers(configs.DB); err != nil {
			fmt.Println("Seeding users failed:", err)
		} else {
			fmt.Println("Seeding users successful")
		}
	} else {
		fmt.Println("Users are already seeded")
	}

	configs.DB.Model((&models.Module{})).Count(&count)
	if count == 0 {
		if err := seeders.SeedModules(configs.DB); err != nil {
			fmt.Println("Seeding modules failed:", err)
		} else {
			fmt.Println("Seeding modules successful")
		}
	} else {
		fmt.Println("Modules are already seeded")
	}

	configs.DB.Model((&models.RoleModule{})).Count(&count)
	if count == 0 {
		if err := seeders.SeedRoleModules(configs.DB); err != nil {
			fmt.Println("Seeding role module failed:", err)
		} else {
			fmt.Println("Seeding role module successful")
		}
	} else {
		fmt.Println("Role module are already seeded")
	}

	if err != nil {
		log.Println(err)
	}

	fmt.Println("Database Migrated")
}