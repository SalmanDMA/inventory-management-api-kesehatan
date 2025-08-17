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

type AreaService struct {
	AreaRepository repositories.AreaRepository
}

func NewAreaService(areaRepo repositories.AreaRepository) *AreaService {
	return &AreaService{
		AreaRepository: areaRepo,
	}
}

func (service *AreaService) GetAllAreas(userInfo *models.User) ([]models.ResponseGetArea, error) {
	areas, err := service.AreaRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var areasResponse []models.ResponseGetArea
	for _, area := range areas {
		areasResponse = append(areasResponse, models.ResponseGetArea{
			ID:          area.ID,
			Name:        area.Name,
			Code:        area.Code,
			Color:       area.Color,
		 Latitude:    area.Latitude,
			Longitude:   area.Longitude,
			CreatedAt: 	area.CreatedAt,
			UpdatedAt: 	area.UpdatedAt,
			DeletedAt: 	area.DeletedAt,
		})
	}

	return areasResponse, nil
}

func (service *AreaService) GetAllAreasPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.AreaPaginatedResponse, error) {
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

	areas, totalCount, err := service.AreaRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	areasResponse := []models.ResponseGetArea{}
	for _, area := range areas {
		areasResponse = append(areasResponse, models.ResponseGetArea{
			ID:          area.ID,
			Name:        area.Name,
			Code:        area.Code,
			Color:       area.Color,
		 Latitude:    area.Latitude,
			Longitude:   area.Longitude,
			CreatedAt: 	area.CreatedAt,
			UpdatedAt: 	area.UpdatedAt,
			DeletedAt: 	area.DeletedAt,
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

	return &models.AreaPaginatedResponse{
		Data:       areasResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *AreaService) GetAreaByID(areaId string) (*models.ResponseGetArea, error) {
	area, err := service.AreaRepository.FindById(areaId, false)

	if err != nil {
		return nil, err
	}

	return &models.ResponseGetArea{
		 ID:          area.ID,
			Name:        area.Name,
			Code:        area.Code,
			Color:       area.Color,
		 Latitude:    area.Latitude,
			Longitude:   area.Longitude,
			CreatedAt: 	area.CreatedAt,
			UpdatedAt: 	area.UpdatedAt,
			DeletedAt: 	area.DeletedAt,
	}, nil
}

func (s *AreaService) CreateArea(req *models.AreaCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Area, error) {
	_ = ctx
	_ = userInfo

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
	name := norm(req.Name)
	code := norm(req.Code)
	color := norm(req.Color)

	if name == "" || code == "" {
		return nil, errors.New("name and code are required")
	}
	if len(name) > 100 || len(code) > 20 {
		return nil, errors.New("name/code exceeds max length")
	}
	if color != "" && !repositories.IsHexColor(color) {
		return nil, errors.New("color must be a valid hex (e.g. #1a2b3c)")
	}
	if req.Latitude != nil && (*req.Latitude < -90 || *req.Latitude > 90) {
		return nil, errors.New("latitude out of range")
	}
	if req.Longitude != nil && (*req.Longitude < -180 || *req.Longitude > 180) {
		return nil, errors.New("longitude out of range")
	}

	if _, err := s.AreaRepository.FindByName(name); err == nil {
		return nil, errors.New("area name already exists")
	} else if !errors.Is(err, repositories.ErrAreaNotFound) {
		return nil, fmt.Errorf("check name failed: %w", err)
	}
	if _, err := s.AreaRepository.FindByCode(code); err == nil {
		return nil, errors.New("area code already exists")
	} else if !errors.Is(err, repositories.ErrAreaNotFound) {
		return nil, fmt.Errorf("check code failed: %w", err)
	}

	newArea := &models.Area{
		ID:        uuid.New(),
		Name:      name, 
		Code:      code,
		Color:     color,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	area, err := s.AreaRepository.Insert(newArea)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("area name/code already exists")
		}
		return nil, fmt.Errorf("insert area failed: %w", err)
	}
	return area, nil
}

func (s *AreaService) UpdateArea(areaID string, upd *models.AreaCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Area, error) {
	_ = ctx
	_ = userInfo

	area, err := s.AreaRepository.FindById(areaID, false)
	if err != nil {
		if errors.Is(err, repositories.ErrAreaNotFound) {
			return nil, errors.New("area not found")
		}
		return nil, fmt.Errorf("find area failed: %w", err)
	}

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	if strings.TrimSpace(upd.Code) != "" {
		newCode := norm(upd.Code)
		if newCode != area.Code {
			if ex, err := s.AreaRepository.FindByCode(newCode); err == nil && ex.ID != area.ID {
				return nil, errors.New("area code already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrAreaNotFound) {
				return nil, fmt.Errorf("check code failed: %w", err)
			}
			area.Code = newCode
		}
	}

	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if newName != area.Name {
			if ex, err := s.AreaRepository.FindByName(newName); err == nil && ex.ID != area.ID {
				return nil, errors.New("area name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrAreaNotFound) {
				return nil, fmt.Errorf("check name failed: %w", err)
			}
			area.Name = newName
		}
	}

	if strings.TrimSpace(upd.Color) != "" {
		newColor := norm(upd.Color)
		if newColor != "" && !repositories.IsHexColor(newColor) {
			return nil, errors.New("color must be a valid hex (e.g. #1a2b3c)")
		}
		area.Color = newColor
	}

	if upd.Latitude != nil {
		if *upd.Latitude < -90 || *upd.Latitude > 90 {
			return nil, errors.New("latitude out of range")
		}
		area.Latitude = upd.Latitude
	}
	if upd.Longitude != nil {
		if *upd.Longitude < -180 || *upd.Longitude > 180 {
			return nil, errors.New("longitude out of range")
		}
		area.Longitude = upd.Longitude
	}

	updated, err := s.AreaRepository.Update(area)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("area name/code already exists")
		}
		return nil, fmt.Errorf("update area failed: %w", err)
	}
	return updated, nil
}

func (service *AreaService) DeleteAreas(areaRequest *models.AreaIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, areaId := range areaRequest.IDs {
		_, err := service.AreaRepository.FindById(areaId.String(), false)
		if err != nil {
			if err == repositories.ErrAreaNotFound {
				log.Printf("Area not found: %v\n", areaId)
				continue
			}
			log.Printf("Error finding area %v: %v\n", areaId, err)
			return errors.New("error finding area")
		}

		if areaRequest.IsHardDelete == "hardDelete" {
			if err := service.AreaRepository.Delete(areaId.String(), true); err != nil {
				log.Printf("Error hard deleting area %v: %v\n", areaId, err)
				return errors.New("error hard deleting area")
			}
		} else {
			if err := service.AreaRepository.Delete(areaId.String(), false); err != nil {
				log.Printf("Error soft deleting area %v: %v\n", areaId, err)
				return errors.New("error soft deleting area")
			}
		}
	}

	return nil
}

func (service *AreaService) RestoreAreas(area *models.AreaRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Area, error) {
	var restoredAreas []models.Area

	for _, areaId := range area.IDs {
		area := &models.Area{ID: areaId}

		restoredArea, err := service.AreaRepository.Restore(area, areaId.String())
		if err != nil {
			if err == repositories.ErrAreaNotFound {
				log.Printf("Area not found: %v\n", areaId)
				continue
			}
			log.Printf("Error restoring area %v: %v\n", areaId, err)
			return nil, errors.New("error restoring area")
		}

		restoredAreas = append(restoredAreas, *restoredArea)
	}

	return restoredAreas, nil
}