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
)

type UoMService struct {
	UoMRepository repositories.UoMRepository
}

func NewUoMService(uomRepo repositories.UoMRepository) *UoMService {
	return &UoMService{
		UoMRepository: uomRepo,
	}
}

func (s *UoMService) GetAllUoMs(userInfo *models.User) ([]models.ResponseGetUoM, error) {
	_ = userInfo
	uoms, err := s.UoMRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResponseGetUoM, 0, len(uoms))
	for _, u := range uoms {
		out = append(out, models.ResponseGetUoM{
			ID:          u.ID,
			Name:        u.Name,
			Color:       u.Color,
			Description: u.Description,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
			DeletedAt:   u.DeletedAt,
		})
	}
	return out, nil
}

func (s *UoMService) GetAllUoMsPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.UoMPaginatedResponse, error) {
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

	uoms, total, err := s.UoMRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	data := make([]models.ResponseGetUoM, 0, len(uoms))
	for _, u := range uoms {
		data = append(data, models.ResponseGetUoM{
			ID:          u.ID,
			Name:        u.Name,
			Color:       u.Color,
			Description: u.Description,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
			DeletedAt:   u.DeletedAt,
		})
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.UoMPaginatedResponse{
		Data: data,
		Pagination: models.PaginationResponse{
			CurrentPage:  req.Page,
			PerPage:      req.Limit,
			TotalPages:   totalPages,
			TotalRecords: total,
			HasNext:      req.Page < totalPages,
			HasPrev:      req.Page > 1,
		},
	}, nil
}

func (s *UoMService) GetUoMByID(uomId string) (*models.ResponseGetUoM, error) {
	uom, err := s.UoMRepository.FindById(nil, uomId, false)
	if err != nil {
		return nil, err
	}

	return &models.ResponseGetUoM{
		ID:          uom.ID,
		Name:        uom.Name,
		Color:       uom.Color,
		Description: uom.Description,
		CreatedAt:   uom.CreatedAt,
		UpdatedAt:   uom.UpdatedAt,
		DeletedAt:   uom.DeletedAt,
	}, nil
}

func (s *UoMService) CreateUoM(req *models.UoMCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.UoM, error) {
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

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if _, err := s.UoMRepository.FindByName(tx, name); err == nil {
		tx.Rollback()
		return nil, errors.New("uom name already exists")
	} else if !errors.Is(err, repositories.ErrUoMNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("check name failed: %w", err)
	}

	entity := &models.UoM{
		ID:          uuid.New(),
		Name:        name,
		Color:       norm(req.Color),
		Description: req.Description,
	}

	created, err := s.UoMRepository.Insert(tx, entity)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("uom name already exists")
		}
		return nil, fmt.Errorf("insert uom failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// fetch ulang (di luar tx) jika perlu preload
	out, err := s.UoMRepository.FindById(nil, created.ID.String(), false)
	if err != nil {
		// fallback ke entity yg sudah dibuat
		return created, nil
	}
	return out, nil
}

func (s *UoMService) UpdateUoM(idStr string, upd *models.UoMCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.UoM, error) {
	_ = ctx
	_ = userInfo

	// baca dulu di luar tx
	cur, err := s.UoMRepository.FindById(nil, idStr, false)
	if err != nil {
		if errors.Is(err, repositories.ErrUoMNotFound) {
			return nil, errors.New("uom not found")
		}
		return nil, fmt.Errorf("find uom failed: %w", err)
	}

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// validasi & update name
	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if len(newName) > 100 {
			tx.Rollback()
			return nil, errors.New("name exceeds max length")
		}
		if newName != cur.Name {
			if ex, err := s.UoMRepository.FindByName(tx, newName); err == nil && ex.ID != cur.ID {
				tx.Rollback()
				return nil, errors.New("uom name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrUoMNotFound) {
				tx.Rollback()
				return nil, fmt.Errorf("check name failed: %w", err)
			}
			cur.Name = newName
		}
	}

	if strings.TrimSpace(upd.Color) != "" {
		cur.Color = norm(upd.Color)
	}
	if upd.Description != nil {
		cur.Description = upd.Description
	}

	updated, err := s.UoMRepository.Update(tx, cur)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("uom name already exists")
		}
		return nil, fmt.Errorf("update uom failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	out, err := s.UoMRepository.FindById(nil, updated.ID.String(), false)
	if err != nil {
		return updated, nil
	}
	return out, nil
}

func (s *UoMService) DeleteUoMs(req *models.UoMIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx
	_ = userInfo

	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("begin tx failed for %s: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		if _, err := s.UoMRepository.FindById(tx, id.String(), true); err != nil {
			tx.Rollback()
			if errors.Is(err, repositories.ErrUoMNotFound) {
				log.Printf("UoM not found: %v\n", id)
				continue
			}
			log.Printf("Error finding uom %v: %v\n", id, err)
			return errors.New("error finding uom")
		}

		if err := s.UoMRepository.Delete(tx, id.String(), req.IsHardDelete == "hardDelete"); err != nil {
			tx.Rollback()
			log.Printf("Error deleting uom %v: %v\n", id, err)
			return errors.New("error deleting uom")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("commit tx failed for %s: %v\n", id, err)
			return errors.New("error committing transaction")
		}
	}
	return nil
}

func (s *UoMService) RestoreUoMs(req *models.UoMRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.UoM, error) {
	_ = ctx
	_ = userInfo

	var restored []models.UoM
	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("begin tx failed for restore %s: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		u := &models.UoM{ID: id}
		if _, err := s.UoMRepository.Restore(tx, id.String()); err != nil {
			tx.Rollback()
			if errors.Is(err, repositories.ErrUoMNotFound) {
				log.Printf("UoM not found: %v\n", id)
				continue
			}
			log.Printf("Error restoring uom %v: %v\n", id, err)
			return nil, errors.New("error restoring uom")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("commit restore failed for %s: %v\n", id, err)
			return nil, errors.New("error committing user restore")
		}

		row, err := s.UoMRepository.FindById(nil, id.String(), false)
		if err != nil {
			restored = append(restored, *u)
			continue
		}
		restored = append(restored, *row)
	}
	return restored, nil
}
