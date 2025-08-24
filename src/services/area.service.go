package services

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AreaService struct {
	AreaRepository repositories.AreaRepository
}

func NewAreaService(areaRepo repositories.AreaRepository) *AreaService {
	return &AreaService{
		AreaRepository: areaRepo,
	}
}

// ==============================
// Reads (tanpa transaction)
// ==============================

func (s *AreaService) GetAllAreas(userInfo *models.User) ([]models.ResponseGetArea, error) {
	_ = userInfo

	areas, err := s.AreaRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResponseGetArea, 0, len(areas))
	for _, a := range areas {
		out = append(out, models.ResponseGetArea{
			ID:        a.ID,
			Name:      a.Name,
			Code:      a.Code,
			Color:     a.Color,
			Latitude:  a.Latitude,
			Longitude: a.Longitude,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			DeletedAt: a.DeletedAt,
		})
	}
	return out, nil
}

func (s *AreaService) GetAllAreasPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.AreaPaginatedResponse, error) {
	_ = userInfo

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

	areas, totalCount, err := s.AreaRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	data := make([]models.ResponseGetArea, 0, len(areas))
	for _, a := range areas {
		data = append(data, models.ResponseGetArea{
			ID:        a.ID,
			Name:      a.Name,
			Code:      a.Code,
			Color:     a.Color,
			Latitude:  a.Latitude,
			Longitude: a.Longitude,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			DeletedAt: a.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.AreaPaginatedResponse{
		Data: data,
		Pagination: models.PaginationResponse{
			CurrentPage:  req.Page,
			PerPage:      req.Limit,
			TotalPages:   totalPages,
			TotalRecords: totalCount,
			HasNext:      req.Page < totalPages,
			HasPrev:      req.Page > 1,
		},
	}, nil
}

func (s *AreaService) GetAreaByID(areaId string) (*models.ResponseGetArea, error) {
	a, err := s.AreaRepository.FindById(nil, areaId, false)
	if err != nil {
		return nil, err
	}

	return &models.ResponseGetArea{
		ID:        a.ID,
		Name:      a.Name,
		Code:      a.Code,
		Color:     a.Color,
		Latitude:  a.Latitude,
		Longitude: a.Longitude,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		DeletedAt: a.DeletedAt,
	}, nil
}

// ==============================
// Mutations (pakai transaction)
// ==============================

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

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Uniqueness in-tx
	if _, err := s.AreaRepository.FindByName(tx, name); err == nil {
		tx.Rollback()
		return nil, errors.New("area name already exists")
	} else if !errors.Is(err, repositories.ErrAreaNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("check name failed: %w", err)
	}
	if _, err := s.AreaRepository.FindByCode(tx, code); err == nil {
		tx.Rollback()
		return nil, errors.New("area code already exists")
	} else if !errors.Is(err, repositories.ErrAreaNotFound) {
		tx.Rollback()
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

	created, err := s.AreaRepository.Insert(tx, newArea)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("area name/code already exists")
		}
		return nil, fmt.Errorf("insert area failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return created, nil
}

func (s *AreaService) UpdateArea(areaID string, upd *models.AreaCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Area, error) {
	_ = ctx
	_ = userInfo

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	area, err := s.AreaRepository.FindById(tx, areaID, false)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, repositories.ErrAreaNotFound) {
			return nil, errors.New("area not found")
		}
		return nil, fmt.Errorf("find area failed: %w", err)
	}

	if strings.TrimSpace(upd.Code) != "" {
		newCode := norm(upd.Code)
		if newCode != area.Code {
			if ex, err := s.AreaRepository.FindByCode(tx, newCode); err == nil && ex.ID != area.ID {
				tx.Rollback()
				return nil, errors.New("area code already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrAreaNotFound) {
				tx.Rollback()
				return nil, fmt.Errorf("check code failed: %w", err)
			}
			area.Code = newCode
		}
	}

	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if newName != area.Name {
			if ex, err := s.AreaRepository.FindByName(tx, newName); err == nil && ex.ID != area.ID {
				tx.Rollback()
				return nil, errors.New("area name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrAreaNotFound) {
				tx.Rollback()
				return nil, fmt.Errorf("check name failed: %w", err)
			}
			area.Name = newName
		}
	}

	if strings.TrimSpace(upd.Color) != "" {
		newColor := norm(upd.Color)
		if newColor != "" && !repositories.IsHexColor(newColor) {
			tx.Rollback()
			return nil, errors.New("color must be a valid hex (e.g. #1a2b3c)")
		}
		area.Color = newColor
	}

	if upd.Latitude != nil {
		if *upd.Latitude < -90 || *upd.Latitude > 90 {
			tx.Rollback()
			return nil, errors.New("latitude out of range")
		}
		area.Latitude = upd.Latitude
	}
	if upd.Longitude != nil {
		if *upd.Longitude < -180 || *upd.Longitude > 180 {
			tx.Rollback()
			return nil, errors.New("longitude out of range")
		}
		area.Longitude = upd.Longitude
	}

	updated, err := s.AreaRepository.Update(tx, area)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("area name/code already exists")
		}
		return nil, fmt.Errorf("update area failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return updated, nil
}

func (s *AreaService) DeleteAreas(in *models.AreaIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return errors.New("areaIds cannot be empty")
	}
	isHard := in.IsHardDelete == "hardDelete"

	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for area %v: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		if _, err := s.AreaRepository.FindById(tx, id.String(), true); err != nil {
			tx.Rollback()
			if err == repositories.ErrAreaNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Area not found: %v\n", id)
				continue
			}
			log.Printf("Error finding area %v: %v\n", id, err)
			return errors.New("error finding area")
		}

		if err := s.AreaRepository.Delete(tx, id.String(), isHard); err != nil {
			tx.Rollback()
			if isHard {
				log.Printf("Error hard deleting area %v: %v\n", id, err)
				return errors.New("error hard deleting area")
			}
			log.Printf("Error soft deleting area %v: %v\n", id, err)
			return errors.New("error soft deleting area")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for area %v: %v\n", id, err)
			return errors.New("error committing delete")
		}
	}
	return nil
}

func (s *AreaService) RestoreAreas(in *models.AreaRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Area, error) {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return nil, errors.New("areaIds cannot be empty")
	}

	var restored []models.Area
	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for area restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		res, err := s.AreaRepository.Restore(tx, id.String())
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrAreaNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Area not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error restoring area %v: %v\n", id, err)
			return nil, errors.New("error restoring area")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing area restore %v: %v\n", id, err)
			return nil, errors.New("error committing area restore")
		}

		a, ferr := s.AreaRepository.FindById(nil, res.ID.String(), true)
		if ferr != nil {
			log.Printf("Error fetching restored area %v: %v\n", id, ferr)
			restored = append(restored, *res)
			continue
		}
		restored = append(restored, *a)
	}
	return restored, nil
}
