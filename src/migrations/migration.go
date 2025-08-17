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
		&models.Area{},
		&models.FacilityType{},
		&models.Facility{},
		&models.SalesPerson{},
		&models.SalesAssignment{},
		&models.Supplier{},
		&models.Payment{},
		&models.PurchaseOrder{},
		&models.PurchaseOrderItem{},
		&models.SalesOrder{},
		&models.SalesOrderItem{},
		&models.Notification{},
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

	configs.DB.Model((&models.Category{})).Count(&count)
	if count == 0 {
		if err := seeders.SeedCategories(configs.DB); err != nil {
			fmt.Println("Seeding categories failed:", err)
		} else {
			fmt.Println("Seeding categories successful")
		}
	} else {
		fmt.Println("Categories are already seeded")
	}

	configs.DB.Model((&models.Item{})).Count(&count)
	if count == 0 {
		if err := seeders.SeedItems(configs.DB); err != nil {
			fmt.Println("Seeding items failed:", err)
		} else {
			fmt.Println("Seeding items successful")
		}
	} else {
		fmt.Println("Items are already seeded")
	}
	
	configs.DB.Model((&models.Area{})).Count(&count)
	if count == 0 {
		if err := seeders.SeedAreas(configs.DB); err != nil {
			fmt.Println("Seeding areas failed:", err)
		} else {
			fmt.Println("Seeding areas successful")
		}
	} else {
		fmt.Println("Areas are already seeded")
	}

	configs.DB.Model((&models.Facility{})).Count(&count)
	if count == 0 {
		if err := seeders.SeedFacilities(configs.DB); err != nil {
			fmt.Println("Seeding facilities failed:", err)
		} else {
			fmt.Println("Seeding facilities successful")
		}
	} else {
		fmt.Println("Facilities are already seeded")
	}

	configs.DB.Model((&models.SalesPerson{})).Count(&count)
	if count == 0 {
		if err := seeders.SeedSalesPersons(configs.DB); err != nil {
			fmt.Println("Seeding sales persons failed:", err)
		} else {
			fmt.Println("Seeding sales persons successful")
		}
	} else {
		fmt.Println("Sales persons are already seeded")
	}

	configs.DB.Model((&models.Supplier{})).Count(&count)
	if count == 0 {
		if err := seeders.SeedSuppliers(configs.DB); err != nil {
			fmt.Println("Seeding suppliers failed:", err)
		} else {
			fmt.Println("Seeding suppliers successful")
		}
	} else {
		fmt.Println("Suppliers are already seeded")
	}


	if err != nil {
		log.Println(err)
	}

	fmt.Println("Database Migrated")
}