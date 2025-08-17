package services

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type FacilityTypeService struct {
	FacilityTypeRepository repositories.FacilityTypeRepository
}

func NewFacilityTypeService(facilityTypeRepo repositories.FacilityTypeRepository) *FacilityTypeService {
	return &FacilityTypeService{
		FacilityTypeRepository: facilityTypeRepo,
	}
}

func (service *FacilityTypeService) GetAllFacilityTypes(userInfo *models.User) ([]models.ResponseGetFacilityType, error) {
	facilityTypes, err := service.FacilityTypeRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var facilityTypesResponse []models.ResponseGetFacilityType
	for _, facilityType := range facilityTypes {
		facilityTypesResponse = append(facilityTypesResponse, models.ResponseGetFacilityType{
			ID:          facilityType.ID,
			Name:        facilityType.Name,
		 Color:       facilityType.Color,
			Description: facilityType.Description,
			CreatedAt: 	facilityType.CreatedAt,
			UpdatedAt: 	facilityType.UpdatedAt,
			DeletedAt: 	facilityType.DeletedAt,
		})
	}

	return facilityTypesResponse, nil
}

func (service *FacilityTypeService) GetAllFacilityTypesPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.FacilityTypePaginatedResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100 
	}
	if req.Status == "" {
		req.Status = "active"
	}

	facilityTypes, totalCount, err := service.FacilityTypeRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	facilityTypesResponse := []models.ResponseGetFacilityType{}
	for _, facilityType := range facilityTypes {
		facilityTypesResponse = append(facilityTypesResponse, models.ResponseGetFacilityType{
			ID:          facilityType.ID,
			Name:        facilityType.Name,
		 Color:       facilityType.Color,
			Description: facilityType.Description,
			CreatedAt: 	facilityType.CreatedAt,
			UpdatedAt: 	facilityType.UpdatedAt,
			DeletedAt: 	facilityType.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	paginationResponse := models.PaginationResponse{
		CurrentPage:  req.Page,
		PerPage:      req.Limit,
		TotalPages:   totalPages,
		TotalRecords: totalCount,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
	}

	return &models.FacilityTypePaginatedResponse{
		Data:       facilityTypesResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *FacilityTypeService) GetFacilityTypeByID(facilityTypeId string) (*models.ResponseGetFacilityType, error) {
	facilityType, err := service.FacilityTypeRepository.FindById(facilityTypeId, false)

	if err != nil {
		return nil, err
	}

	return &models.ResponseGetFacilityType{
		 ID:          facilityType.ID,
			Name:        facilityType.Name,
		 Color:       facilityType.Color,
			Description: facilityType.Description,
			CreatedAt: 	facilityType.CreatedAt,
			UpdatedAt: 	facilityType.UpdatedAt,
			DeletedAt: 	facilityType.DeletedAt,
	}, nil
}

func (s *FacilityTypeService) CreateFacilityType(req *models.FacilityTypeCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.FacilityType, error) {
	_ = ctx
	_ = userInfo

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	name := norm(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	if len(name) > 100 {
		return nil, errors.New("name exceeds max length")
	}

	if _, err := s.FacilityTypeRepository.FindByName(name); err == nil {
		return nil, errors.New("facility type name already exists")
	} else if !errors.Is(err, repositories.ErrFacilityTypeNotFound) {
		return nil, fmt.Errorf("check name failed: %w", err)
	}

	ft := &models.FacilityType{
		ID:          uuid.New(),
		Name:        name, 
		Color:       norm(req.Color),
		Description: req.Description, 
	}

	created, err := s.FacilityTypeRepository.Insert(ft)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("facility type name already exists")
		}
		return nil, fmt.Errorf("insert facility type failed: %w", err)
	}
	return created, nil
}

func (s *FacilityTypeService) UpdateFacilityType(idStr string, upd *models.FacilityTypeCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.FacilityType, error) {
	_ = ctx
	_ = userInfo

	ft, err := s.FacilityTypeRepository.FindById(idStr, false)
	if err != nil {
		if errors.Is(err, repositories.ErrFacilityTypeNotFound) {
			return nil, errors.New("facility type not found")
		}
		return nil, fmt.Errorf("find facility type failed: %w", err)
	}

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if len(newName) > 100 {
			return nil, errors.New("name exceeds max length")
		}
		if newName != ft.Name {
			if ex, err := s.FacilityTypeRepository.FindByName(newName); err == nil && ex.ID != ft.ID {
				return nil, errors.New("facility type name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrFacilityTypeNotFound) {
				return nil, fmt.Errorf("check name failed: %w", err)
			}
			ft.Name = newName
		}
	}

	if strings.TrimSpace(upd.Color) != "" {
		ft.Color = norm(upd.Color)
	}

	if upd.Description != nil {
		ft.Description = upd.Description
	}

	updated, err := s.FacilityTypeRepository.Update(ft)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("facility type name already exists")
		}
		return nil, fmt.Errorf("update facility type failed: %w", err)
	}
	return updated, nil
}

func (service *FacilityTypeService) DeleteFacilityTypes(facilityTypeRequest *models.FacilityTypeIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, facilityTypeId := range facilityTypeRequest.IDs {
		_, err := service.FacilityTypeRepository.FindById(facilityTypeId.String(), false)
		if err != nil {
			if err == repositories.ErrFacilityTypeNotFound {
				log.Printf("FacilityType not found: %v\n", facilityTypeId)
				continue
			}
			log.Printf("Error finding facility type %v: %v\n", facilityTypeId, err)
			return errors.New("error finding facility type")
		}

		if facilityTypeRequest.IsHardDelete == "hardDelete" {
			if err := service.FacilityTypeRepository.Delete(facilityTypeId.String(), true); err != nil {
				log.Printf("Error hard deleting facility type %v: %v\n", facilityTypeId, err)
				return errors.New("error hard deleting facility type")
			}
		} else {
			if err := service.FacilityTypeRepository.Delete(facilityTypeId.String(), false); err != nil {
				log.Printf("Error soft deleting facility type %v: %v\n", facilityTypeId, err)
				return errors.New("error soft deleting facility type")
			}
		}
	}

	return nil
}

func (service *FacilityTypeService) RestoreFacilityTypes(facilityType *models.FacilityTypeRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.FacilityType, error) {
	var restoredFacilityTypes []models.FacilityType

	for _, facilityTypeId := range facilityType.IDs {
		facilityType := &models.FacilityType{ID: facilityTypeId}

		restoredFacilityType, err := service.FacilityTypeRepository.Restore(facilityType, facilityTypeId.String())
		if err != nil {
			if err == repositories.ErrFacilityTypeNotFound {
				log.Printf("FacilityType not found: %v\n", facilityTypeId)
				continue
			}
			log.Printf("Error restoring facility type %v: %v\n", facilityTypeId, err)
			return nil, errors.New("error restoring facility type")
		}

		restoredFacilityTypes = append(restoredFacilityTypes, *restoredFacilityType)
	}

	return restoredFacilityTypes, nil
}