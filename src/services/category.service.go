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

type CategoryService struct {
	CategoryRepository repositories.CategoryRepository
}

func NewCategoryService(categoryRepo repositories.CategoryRepository) *CategoryService {
	return &CategoryService{
		CategoryRepository: categoryRepo,
	}
}

func (service *CategoryService) GetAllCategories() ([]models.ResponseGetCategory, error) {
	categories, err := service.CategoryRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var categoriesResponse []models.ResponseGetCategory
	for _, category := range categories {
		categoriesResponse = append(categoriesResponse, models.ResponseGetCategory{
			ID:          category.ID,
			Name:        category.Name,
			Color:       category.Color,
			Description: category.Description,
			Items:       category.Items,
			CreatedAt: 	category.CreatedAt,
			UpdatedAt: 	category.UpdatedAt,
			DeletedAt: 	category.DeletedAt,
		})
	}

	return categoriesResponse, nil
}

func (service *CategoryService) GetAllCategoriesPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.CategoryPaginatedResponse, error) {
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

	categories, totalCount, err := service.CategoryRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	categoriesResponse := []models.ResponseGetCategory{}
	for _, category := range categories {
		categoriesResponse = append(categoriesResponse, models.ResponseGetCategory{
			ID:          category.ID,
			Name:        category.Name,
			Color:       category.Color,
			Description: category.Description,
			Items:       category.Items,
			CreatedAt: 	category.CreatedAt,
			UpdatedAt: 	category.UpdatedAt,
			DeletedAt: 	category.DeletedAt,
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

	return &models.CategoryPaginatedResponse{
		Data:       categoriesResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *CategoryService) GetCategoryByID(categoryId string) (*models.ResponseGetCategory, error) {
	category, err := service.CategoryRepository.FindById(categoryId, false)

	if err != nil {
		return nil, err
	}

	return &models.ResponseGetCategory{
		 ID:          category.ID,
			Name:        category.Name,
			Color:       category.Color,
			Description: category.Description,
	}, nil
}

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

	if _, err := s.CategoryRepository.FindByName(name); err == nil {
		return nil, errors.New("category already exists")
	} else if !errors.Is(err, repositories.ErrCategoryNotFound) {
		return nil, fmt.Errorf("check category failed: %w", err)
	}

	cat := &models.Category{
		ID:          uuid.New(),
		Name:        name,   
		Color:       color,  
		Description: desc,
	}

	created, err := s.CategoryRepository.Insert(cat)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("category already exists")
		}
		return nil, fmt.Errorf("insert category failed: %w", err)
	}
	return created, nil
}

func (s *CategoryService) UpdateCategory(categoryID string, upd *models.CategoryUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Category, error) {
	_ = ctx
	_ = userInfo

	existing, err := s.CategoryRepository.FindById(categoryID, false)
	if err != nil {
		if errors.Is(err, repositories.ErrCategoryNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("find category failed: %w", err)
	}

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if len(newName) > 100 {
			return nil, errors.New("name exceeds max length")
		}
		if newName != existing.Name {
			if ex, err := s.CategoryRepository.FindByName(newName); err == nil && ex.ID != existing.ID {
				return nil, errors.New("category already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrCategoryNotFound) {
				return nil, fmt.Errorf("check category failed: %w", err)
			}
			existing.Name = newName
		}
	}

	if strings.TrimSpace(upd.Color) != "" {
		newColor := norm(upd.Color)
		if newColor != "" && !repositories.IsHexColor(newColor) {
			return nil, errors.New("color must be a valid hex (e.g. #1a2b3c)")
		}
		existing.Color = newColor
	}

	if upd.Description != "" {
		existing.Description = upd.Description
	}

	updated, err := s.CategoryRepository.Update(existing)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("category already exists")
		}
		return nil, fmt.Errorf("update category failed: %w", err)
	}
	return updated, nil
}

func (service *CategoryService) DeleteCategories(categoryRequest *models.CategoryIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, categoryId := range categoryRequest.IDs {
		_, err := service.CategoryRepository.FindById(categoryId.String(), false)
		if err != nil {
			if err == repositories.ErrCategoryNotFound {
				log.Printf("Category not found: %v\n", categoryId)
				continue
			}
			log.Printf("Error finding category %v: %v\n", categoryId, err)
			return errors.New("error finding usecategoryr")
		}

		if categoryRequest.IsHardDelete == "hardDelete" {
			if err := service.CategoryRepository.Delete(categoryId.String(), true); err != nil {
				log.Printf("Error hard deleting category %v: %v\n", categoryId, err)
				return errors.New("error hard deleting category")
			}
		} else {
			if err := service.CategoryRepository.Delete(categoryId.String(), false); err != nil {
				log.Printf("Error soft deleting category %v: %v\n", categoryId, err)
				return errors.New("error soft deleting category")
			}
		}
	}

	return nil
}

func (service *CategoryService) RestoreCategories(category *models.CategoryRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Category, error) {
	var restoredCategories []models.Category

	for _, categoryId := range category.IDs {
		category := &models.Category{ID: categoryId}

		restoredCategory, err := service.CategoryRepository.Restore(category, categoryId.String())
		if err != nil {
			if err == repositories.ErrCategoryNotFound {
				log.Printf("Category not found: %v\n", categoryId)
				continue
			}
			log.Printf("Error restoring category %v: %v\n", categoryId, err)
			return nil, errors.New("error restoring category")
		}

		restoredCategories = append(restoredCategories, *restoredCategory)
	}

	return restoredCategories, nil
}