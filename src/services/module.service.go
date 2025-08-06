package services

import (
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ModuleService struct {
	ModuleRepository repositories.ModuleRepository
}

func NewModuleService(moduleRepository repositories.ModuleRepository) *ModuleService {
	return &ModuleService{
		ModuleRepository: moduleRepository,
	}
}

func (service *ModuleService) GetAllModules() ([]models.ResponseGetModule, error) {
	modules, err := service.ModuleRepository.FindAll()

	if err != nil {
		return nil, err
	}

	modulesResponse := []models.ResponseGetModule{}

	for _, module := range modules {
		modulesResponse = append(modulesResponse, models.ResponseGetModule{
			ID:          module.ID,
			Name:       module.Name,
			ParentID:    module.ParentID,
			Parent:      module.Parent,
			ModuleTypeID: module.ModuleTypeID,
			ModuleType:  module.ModuleType,
			Route:       module.Route,
			Path:       module.Path,
			Icon:        module.Icon,
			Children: 		module.Children,
			Description: module.Description,
			RoleModules: module.RoleModules,
			DeletedAt:   module.DeletedAt,
		})
	}

	return modulesResponse, nil
}

func (service *ModuleService) CreateModule(moduleRequest *models.ModuleCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Module, error) {
	var parentID *int
	var parentName string

	if moduleRequest.ParentID != nil {
		parentID = moduleRequest.ParentID
		parentModule, err := service.ModuleRepository.FindById(*parentID, false)
		if err != nil {
			return nil, err
		}
		parentID = &parentModule.ID
		parentName = parentModule.Name

		if parentName == "Root" {
			parentID = nil
		}
	}

	route, err := helpers.FormatRoute(moduleRequest.Route)
	if err != nil {
		return nil, err
	}

	newModule := &models.Module{
		Name:       helpers.CapitalizeTitle(moduleRequest.Name),
		ParentID:    parentID,
		ModuleTypeID: moduleRequest.ModuleTypeID,
		Route:       route,
		Path:       moduleRequest.Path,
		Icon:      moduleRequest.Icon,
		Description: moduleRequest.Description,
	}

	createdModule, err := service.ModuleRepository.Insert(newModule)
	if err != nil {
		return nil, err
	}
	return createdModule, nil
}

func (service *ModuleService) UpdateModule(moduleId int, moduleRequest *models.ModuleUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Module, error) {
	module, err := service.ModuleRepository.FindById(moduleId, true)
	if err != nil {
					return nil, err
	}

	var parentID *int
	var parentName string

	fmt.Println(moduleRequest, "moduleRequest")

	if moduleRequest.ParentID != nil {
					parentID = moduleRequest.ParentID
					parentModule, err := service.ModuleRepository.FindById(*parentID, false)
					if err != nil {
									return nil, err
					}
					parentID = &parentModule.ID
					parentName = parentModule.Name

					if parentName == "Root" {
									parentID = nil
					}
	}

	if moduleRequest.ModuleTypeID != uuid.Nil {
					module.ModuleTypeID = moduleRequest.ModuleTypeID
	}
	
	route, err := helpers.FormatRoute(moduleRequest.Route)
	if err != nil {
					return nil, err
	}

	if moduleRequest.Name != "" {
					module.Name = helpers.CapitalizeTitle(moduleRequest.Name)
	}

	if moduleRequest.ParentID != nil {
					module.ParentID = parentID
					module.Parent = nil
	}
	
	if moduleRequest.Route != "" {
					module.Route = route
	}

	if moduleRequest.Path != "" {
					module.Path = moduleRequest.Path
	}

	if moduleRequest.Icon != "" {
					module.Icon = moduleRequest.Icon
	}

	if moduleRequest.Description != "" {
					module.Description = moduleRequest.Description
	}

	fmt.Println(module, "module")

	updatedModule, err := service.ModuleRepository.Update(module)
	if err != nil {
					return nil, err
	}

	return updatedModule, nil
}

func (service *ModuleService) DeleteModule(moduleRequest *models.ModuleIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {

	if len(moduleRequest.IDs) == 0 {
					return fmt.Errorf("moduleIds cannot be empty")
	}

	allModules, err := service.ModuleRepository.FindAll()
	if err != nil {
					return fmt.Errorf("failed to retrieve all modules: %w", err)
	}

	modulesToDelete := make(map[int]bool)
	for _, moduleId := range moduleRequest.IDs {
					modulesToDelete[moduleId] = true
	}

	var addModulesWithParentID func(parentID int)
	addModulesWithParentID = func(parentID int) {
					for _, module := range allModules {
									if module.ParentID != nil && *module.ParentID == parentID {
													if !modulesToDelete[module.ID] {
																	modulesToDelete[module.ID] = true
																	addModulesWithParentID(module.ID)
													}
									}
					}
	}

	for _, module := range allModules {
					if module.ParentID != nil && modulesToDelete[module.ID] {
									parentID := *module.ParentID
									if !modulesToDelete[parentID] {
													module.ParentID = nil
													if _, err := service.ModuleRepository.Update(&module); err != nil {
																	return err
													}
									}
					}
	}

	for moduleId := range modulesToDelete {
					addModulesWithParentID(moduleId)
	}

	for moduleId := range modulesToDelete {
					_, err := service.ModuleRepository.FindById(moduleId, false)

					if err != nil {
								return err
					}

					if moduleRequest.IsHardDelete == "hardDelete" {
									if err := service.ModuleRepository.Delete(moduleId, true); err != nil {
													return err
									}
					} else {
									if err := service.ModuleRepository.Delete(moduleId, false); err != nil {
													return err
									}
					}
	}

	return nil
}



func (service *ModuleService) RestoreModule(moduleRequest *models.ModuleRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Module, error) {
	var restoredModules []models.Module

	if len(moduleRequest.IDs) == 0 {
					return nil, fmt.Errorf("moduleIds cannot be empty")
	}

	allModules, err := service.ModuleRepository.FindAll()
	if err != nil {
					return nil, fmt.Errorf("failed to retrieve all modules: %w", err)
	}

	modulesToRestore := make(map[int]bool)
	for _, moduleId := range moduleRequest.IDs {
					modulesToRestore[moduleId] = true
	}

	var addModulesWithParentID func(parentID int)
	addModulesWithParentID = func(parentID int) {
					for _, module := range allModules {
									if module.ParentID != nil && *module.ParentID == parentID {
													if !modulesToRestore[module.ID] {
																	modulesToRestore[module.ID] = true
																	addModulesWithParentID(module.ID)
													}
									}
					}
	}

	for _, module := range allModules {
					if module.ParentID != nil && modulesToRestore[module.ID] {
									parentID := *module.ParentID
									if !modulesToRestore[parentID] {
													module.ParentID = nil
													if _, err := service.ModuleRepository.Update(&module); err != nil {
																	return nil, err
													}
									}
					}
	}

	for moduleId := range modulesToRestore {
					addModulesWithParentID(moduleId)
	}

	for moduleId := range modulesToRestore {
					module := &models.Module{
									ID: moduleId,
					}

					restoredModule, err := service.ModuleRepository.Restore(module, moduleId)
					if err != nil {
									if err == repositories.ErrModuleNotFound {
													continue
									}
									return nil, err
					}

					restoredModules = append(restoredModules, *restoredModule)
	}

	return restoredModules, nil
}


