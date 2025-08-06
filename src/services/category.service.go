package services

import (
	"errors"
	"fmt"
	"log"

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

func (service *CategoryService) CreateCategory(categoryRequest *models.CategoryCreateRequest, ctx *fiber.Ctx, userInfo *models.User) ( *models.Category, error) {
	if _, err := service.CategoryRepository.FindByName(categoryRequest.Name); err == nil {
		return nil, errors.New("category already exists") 
	} else if err != repositories.ErrCategoryNotFound {
		 return nil, errors.New("error checking category: " + err.Error())
	}

	newCategory := &models.Category{
		ID: 									uuid.New(),
		Name:        categoryRequest.Name,
		Color:       categoryRequest.Color,
		Description: categoryRequest.Description,
	}

	fmt.Println(newCategory, "newCategory")

	category, err := service.CategoryRepository.Insert(newCategory)

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (service *CategoryService) UpdateCategory(categoryID string, categoryUpdate *models.CategoryUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Category, error) {
	categoryExists, err := service.CategoryRepository.FindById(categoryID, false)
	if err != nil {
		return nil, err
	}
	if categoryExists == nil {
		return nil, errors.New("category not found")
	}

	if categoryUpdate.Name != "" {
		categoryExists.Name = categoryUpdate.Name
	}

	if categoryUpdate.Color != "" {
		categoryExists.Color = categoryUpdate.Color
	}

	if categoryUpdate.Description != "" {
		categoryExists.Description = categoryUpdate.Description
	}

	updateCategory , err := service.CategoryRepository.Update(categoryExists)
	if err != nil {
		return nil, err
	}
	
	return updateCategory, nil
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