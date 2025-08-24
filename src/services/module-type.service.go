package services

import (
	"errors"
	"log"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModuleTypeService struct {
	ModuleTypeRepository repositories.ModuleTypeRepository
}

func NewModuleTypeService(moduleTypeRepository repositories.ModuleTypeRepository) *ModuleTypeService {
	return &ModuleTypeService{
		ModuleTypeRepository: moduleTypeRepository,
	}
}

// ==============================
// Reads (no tx required)
// ==============================

func (s *ModuleTypeService) GetAllModuleTypes() ([]models.ResponseGetModuleType, error) {
	moduleTypes, err := s.ModuleTypeRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetModuleType, 0, len(moduleTypes))
	for _, mt := range moduleTypes {
		resp = append(resp, models.ResponseGetModuleType{
			ID:          mt.ID,
			Icon:        mt.Icon,
			Name:        mt.Name,
			Description: mt.Description,
			DeletedAt:   mt.DeletedAt,
		})
	}
	return resp, nil
}

func (s *ModuleTypeService) GetModuleTypeByID(moduleTypeId string) (*models.ResponseGetModuleType, error) {
	mt, err := s.ModuleTypeRepository.FindById(nil, moduleTypeId, false)
	if err != nil {
		return nil, err
	}
	return &models.ResponseGetModuleType{
		ID:          mt.ID,
		Name:        mt.Name,
		Description: mt.Description,
		Icon:        mt.Icon,
		DeletedAt:   mt.DeletedAt,
	}, nil
}

// ==============================
// Mutations (transaction-aware; configs.DB.Begin())
// ==============================

func (s *ModuleTypeService) CreateModuleType(req *models.ModuleTypeCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.ModuleType, error) {
	_ = ctx; _ = userInfo

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	newMT := &models.ModuleType{
		ID:          uuid.New(),
		Name:        req.Name,
		Icon:        req.Icon,
		Description: req.Description,
	}

	created, err := s.ModuleTypeRepository.Insert(tx, newMT)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return created, nil
}

func (s *ModuleTypeService) UpdateModuleType(moduleTypeId string, req *models.ModuleTypeUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.ModuleType, error) {
	_ = ctx; _ = userInfo

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	mt, err := s.ModuleTypeRepository.FindById(tx, moduleTypeId, false)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if req.Name != "" {
		mt.Name = req.Name
	}
	if req.Icon != "" {
		mt.Icon = req.Icon
	}
	if req.Description != "" {
		mt.Description = req.Description
	}

	updated, err := s.ModuleTypeRepository.Update(tx, mt)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *ModuleTypeService) DeleteModuleTypes(in *models.ModuleTypeIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return errors.New("moduleTypeIds cannot be empty")
	}
	isHard := in.IsHardDelete == "hardDelete"

	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for moduleType %v: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		if _, err := s.ModuleTypeRepository.FindById(tx, id.String(), true); err != nil {
			tx.Rollback()
			if err == repositories.ErrModuleTypeNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("ModuleType not found: %v\n", id)
				continue
			}
			log.Printf("Error finding moduleType %v: %v\n", id, err)
			return errors.New("error finding moduleType")
		}

		if err := s.ModuleTypeRepository.Delete(tx, id.String(), isHard); err != nil {
			tx.Rollback()
			log.Printf("Error deleting moduleType %v: %v\n", id, err)
			return errors.New("error deleting moduleType")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for moduleType %v: %v\n", id, err)
			return errors.New("error committing delete")
		}
	}
	return nil
}

func (s *ModuleTypeService) RestoreModuleTypes(in *models.ModuleTypeRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.ModuleType, error) {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return nil, errors.New("moduleTypeIds cannot be empty")
	}

	var restored []models.ModuleType
	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for moduleType restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		res, err := s.ModuleTypeRepository.Restore(tx, id.String())
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrModuleTypeNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("ModuleType not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error restoring moduleType %v: %v\n", id, err)
			return nil, errors.New("error restoring moduleType")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing moduleType restore %v: %v\n", id, err)
			return nil, errors.New("error committing moduleType restore")
		}

		mt, ferr := s.ModuleTypeRepository.FindById(nil, res.ID.String(), true)
		if ferr != nil {
			log.Printf("Error fetching restored moduleType %v: %v\n", id, ferr)
			continue
		}
		restored = append(restored, *mt)
	}
	return restored, nil
}

