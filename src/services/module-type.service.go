package services

import (
	"errors"
	"log"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ModuleTypeService struct {
	ModuleTypeRepository repositories.ModuleTypeRepository
}

func NewModuleTypeService(moduleTypeRepository repositories.ModuleTypeRepository) *ModuleTypeService {
	return &ModuleTypeService{ModuleTypeRepository: moduleTypeRepository}
}

func (service *ModuleTypeService) GetAllModuleTypes() ([]models.ResponseGetModuleType, error) {
	moduleTypes, err := service.ModuleTypeRepository.FindAll()

	if err != nil {
		return nil, err
	}

	moduleTypesReponse := []models.ResponseGetModuleType{}

	for _, moduleType := range moduleTypes {
		moduleTypesReponse = append(moduleTypesReponse, models.ResponseGetModuleType{
			ID:          moduleType.ID,
			Icon:        moduleType.Icon,
			Name:        moduleType.Name,
			Description: moduleType.Description,
			DeletedAt:   moduleType.DeletedAt,
		})
	}

	return moduleTypesReponse, nil
}

func (service *ModuleTypeService) GetModuleTypeByID(moduleTypeId string) (*models.ResponseGetModuleType, error) {
	moduleType, err := service.ModuleTypeRepository.FindById(moduleTypeId, false)

	if err != nil {
		return nil, err
	}

	return &models.ResponseGetModuleType{
		 ID:          moduleType.ID,
			Name:        moduleType.Name,
			Description: moduleType.Description,
			Icon:        moduleType.Icon,
			DeletedAt:   moduleType.DeletedAt,
	}, nil
}



func (service *ModuleTypeService) CreateModuleType(moduleType *models.ModuleTypeCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.ModuleType, error) {
	newModuleType := &models.ModuleType{
		ID:          uuid.New(),
		Name:        moduleType.Name,
		Icon:        moduleType.Icon,
		Description: moduleType.Description,
	}

	createdModuleType, err := service.ModuleTypeRepository.Insert(newModuleType)

	if err != nil {
		return nil, err
	}

	return createdModuleType, nil
}

func (service *ModuleTypeService) UpdateModuleType(moduleTypeId string, moduleTypeRequest *models.ModuleTypeUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.ModuleType, error) {
	moduleType, err := service.ModuleTypeRepository.FindById(moduleTypeId, false)

	if err != nil {
		return nil, err
	}

	if moduleTypeRequest.Name != "" {
		moduleType.Name = moduleTypeRequest.Name
	}

	if moduleTypeRequest.Icon != "" {
		moduleType.Icon = moduleTypeRequest.Icon
	}

	if moduleTypeRequest.Description != "" {
		moduleType.Description = moduleTypeRequest.Description
	}

	updatedModuleType, err := service.ModuleTypeRepository.Update(moduleType)

	if err != nil {
		return nil, err
	}

	return updatedModuleType, nil
}


func (service *ModuleTypeService) DeleteModuleTypes(moduleTypeRequest *models.ModuleTypeIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, moduleTypeId := range moduleTypeRequest.IDs {
		_, err := service.ModuleTypeRepository.FindById(moduleTypeId.String(), false)
		if err != nil {
			if err == repositories.ErrModuleTypeNotFound {
				log.Printf("ModuleType not found: %v\n", moduleTypeId)
				continue
			}
			log.Printf("Error finding moduleType %v: %v\n", moduleTypeId, err)
			return errors.New("error finding usemoduleTyper")
		}

		if moduleTypeRequest.IsHardDelete == "hardDelete" {
			if err := service.ModuleTypeRepository.Delete(moduleTypeId.String(), true); err != nil {
				log.Printf("Error hard deleting moduleType %v: %v\n", moduleTypeId, err)
				return errors.New("error hard deleting moduleType")
			}
		} else {
			if err := service.ModuleTypeRepository.Delete(moduleTypeId.String(), false); err != nil {
				log.Printf("Error soft deleting moduleType %v: %v\n", moduleTypeId, err)
				return errors.New("error soft deleting moduleType")
			}
		}
	}

	return nil
}

func (service *ModuleTypeService) RestoreModuleTypes(moduleType *models.ModuleTypeRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.ModuleType, error) {
	var restoredModuleTypes []models.ModuleType

	for _, moduleTypeId := range moduleType.IDs {
		moduleType := &models.ModuleType{ID: moduleTypeId}

		restoredModuleType, err := service.ModuleTypeRepository.Restore(moduleType, moduleTypeId.String())
		if err != nil {
			if err == repositories.ErrModuleTypeNotFound {
				log.Printf("ModuleType not found: %v\n", moduleTypeId)
				continue
			}
			log.Printf("Error restoring moduleType %v: %v\n", moduleTypeId, err)
			return nil, errors.New("error restoring moduleType")
		}

		restoredModuleTypes = append(restoredModuleTypes, *restoredModuleType)
	}

	return restoredModuleTypes, nil
}