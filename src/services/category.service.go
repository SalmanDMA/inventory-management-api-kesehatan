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

type CategoryService struct {
	CategoryRepository repositories.CategoryRepository
}

func NewCategoryService(categoryRepo repositories.CategoryRepository) *CategoryService {
	return &CategoryService{
		CategoryRepository: categoryRepo,
	}
}

// ==============================
// Reads (tanpa transaction)
// ==============================

func (s *CategoryService) GetAllCategories() ([]models.ResponseGetCategory, error) {
	categories, err := s.CategoryRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResponseGetCategory, 0, len(categories))
	for _, c := range categories {
		out = append(out, models.ResponseGetCategory{
			ID:          c.ID,
			Name:        c.Name,
			Color:       c.Color,
			Description: c.Description,
			Items:       c.Items,
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
			DeletedAt:   c.DeletedAt,
		})
	}
	return out, nil
}

func (s *CategoryService) GetAllCategoriesPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.CategoryPaginatedResponse, error) {
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

	list, totalCount, err := s.CategoryRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	data := make([]models.ResponseGetCategory, 0, len(list))
	for _, c := range list {
		data = append(data, models.ResponseGetCategory{
			ID:          c.ID,
			Name:        c.Name,
			Color:       c.Color,
			Description: c.Description,
			Items:       c.Items,
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
			DeletedAt:   c.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.CategoryPaginatedResponse{
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

func (s *CategoryService) GetCategoryByID(categoryId string) (*models.ResponseGetCategory, error) {
	c, err := s.CategoryRepository.FindById(nil, categoryId, false)
	if err != nil {
		return nil, err
	}

	return &models.ResponseGetCategory{
		ID:          c.ID,
		Name:        c.Name,
		Color:       c.Color,
		Description: c.Description,
	}, nil
}

// ==============================
// Mutations
// ==============================

func (s *CategoryService) CreateCategory(req *models.CategoryCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Category, error) {
	_ = ctx
	_ = userInfo

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	name := norm(req.Name)
	color := norm(req.Color)
	desc := req.Description

	if name == "" {
		return nil, errors.New("name is required")
	}
	if len(name) > 100 {
		return nil, errors.New("name exceeds max length")
	}
	if color != "" && !repositories.IsHexColor(color) {
		return nil, errors.New("color must be a valid hex (e.g. #1a2b3c)")
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

	// uniqueness check in-tx
	if _, err := s.CategoryRepository.FindByName(tx, name); err == nil {
		tx.Rollback()
		return nil, errors.New("category already exists")
	} else if !errors.Is(err, repositories.ErrCategoryNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("check category failed: %w", err)
	}

	cat := &models.Category{
		ID:          uuid.New(),
		Name:        name,
		Color:       color,
		Description: desc,
	}

	created, err := s.CategoryRepository.Insert(tx, cat)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("category already exists")
		}
		return nil, fmt.Errorf("insert category failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return created, nil
}

func (s *CategoryService) UpdateCategory(categoryID string, upd *models.CategoryUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Category, error) {
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

	existing, err := s.CategoryRepository.FindById(tx, categoryID, false)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, repositories.ErrCategoryNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("find category failed: %w", err)
	}

	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if len(newName) > 100 {
			tx.Rollback()
			return nil, errors.New("name exceeds max length")
		}
		if newName != existing.Name {
			if ex, err := s.CategoryRepository.FindByName(tx, newName); err == nil && ex.ID != existing.ID {
				tx.Rollback()
				return nil, errors.New("category already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrCategoryNotFound) {
				tx.Rollback()
				return nil, fmt.Errorf("check category failed: %w", err)
			}
			existing.Name = newName
		}
	}

	if strings.TrimSpace(upd.Color) != "" {
		newColor := norm(upd.Color)
		if newColor != "" && !repositories.IsHexColor(newColor) {
			tx.Rollback()
			return nil, errors.New("color must be a valid hex (e.g. #1a2b3c)")
		}
		existing.Color = newColor
	}

	if upd.Description != "" {
		existing.Description = upd.Description
	}

	updated, err := s.CategoryRepository.Update(tx, existing)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("category already exists")
		}
		return nil, fmt.Errorf("update category failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return updated, nil
}

func (s *CategoryService) DeleteCategories(in *models.CategoryIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return errors.New("categoryIds cannot be empty")
	}
	isHard := in.IsHardDelete == "hardDelete"

	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for category %v: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		if _, err := s.CategoryRepository.FindById(tx, id.String(), true); err != nil {
			tx.Rollback()
			if err == repositories.ErrCategoryNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Category not found: %v\n", id)
				continue
			}
			log.Printf("Error finding category %v: %v\n", id, err)
			return errors.New("error finding category")
		}

		if err := s.CategoryRepository.Delete(tx, id.String(), isHard); err != nil {
			tx.Rollback()
			if isHard {
				log.Printf("Error hard deleting category %v: %v\n", id, err)
				return errors.New("error hard deleting category")
			}
			log.Printf("Error soft deleting category %v: %v\n", id, err)
			return errors.New("error soft deleting category")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for category %v: %v\n", id, err)
			return errors.New("error committing delete")
		}
	}
	return nil
}

func (s *CategoryService) RestoreCategories(in *models.CategoryRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Category, error) {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return nil, errors.New("categoryIds cannot be empty")
	}

	var restored []models.Category
	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for category restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		res, err := s.CategoryRepository.Restore(tx, id.String())
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrCategoryNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Category not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error restoring category %v: %v\n", id, err)
			return nil, errors.New("error restoring category")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing category restore %v: %v\n", id, err)
			return nil, errors.New("error committing category restore")
		}

		cat, ferr := s.CategoryRepository.FindById(nil, res.ID.String(), true)
		if ferr != nil {
			log.Printf("Error fetching restored category %v: %v\n", id, ferr)
			restored = append(restored, *res) 
			continue
		}
		restored = append(restored, *cat)
	}
	return restored, nil
}
