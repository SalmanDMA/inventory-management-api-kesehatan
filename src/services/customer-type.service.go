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

type CustomerTypeService struct {
	CustomerTypeRepository repositories.CustomerTypeRepository
}

func NewCustomerTypeService(customerTypeRepo repositories.CustomerTypeRepository) *CustomerTypeService {
	return &CustomerTypeService{
		CustomerTypeRepository: customerTypeRepo,
	}
}

// ==============================
// Reads (tanpa transaction)
// ==============================

func (s *CustomerTypeService) GetAllCustomerTypes(userInfo *models.User) ([]models.ResponseGetCustomerType, error) {
	_ = userInfo

	types, err := s.CustomerTypeRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResponseGetCustomerType, 0, len(types))
	for _, ft := range types {
		out = append(out, models.ResponseGetCustomerType{
			ID:          ft.ID,
			Name:        ft.Name,
			Color:       ft.Color,
			Description: ft.Description,
			CreatedAt:   ft.CreatedAt,
			UpdatedAt:   ft.UpdatedAt,
			DeletedAt:   ft.DeletedAt,
		})
	}
	return out, nil
}

func (s *CustomerTypeService) GetAllCustomerTypesPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.CustomerTypePaginatedResponse, error) {
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

	list, totalCount, err := s.CustomerTypeRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	data := make([]models.ResponseGetCustomerType, 0, len(list))
	for _, ft := range list {
		data = append(data, models.ResponseGetCustomerType{
			ID:          ft.ID,
			Name:        ft.Name,
			Color:       ft.Color,
			Description: ft.Description,
			CreatedAt:   ft.CreatedAt,
			UpdatedAt:   ft.UpdatedAt,
			DeletedAt:   ft.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.CustomerTypePaginatedResponse{
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

func (s *CustomerTypeService) GetCustomerTypeByID(id string) (*models.ResponseGetCustomerType, error) {
	ft, err := s.CustomerTypeRepository.FindById(nil, id, false)
	if err != nil {
		return nil, err
	}
	return &models.ResponseGetCustomerType{
		ID:          ft.ID,
		Name:        ft.Name,
		Color:       ft.Color,
		Description: ft.Description,
		CreatedAt:   ft.CreatedAt,
		UpdatedAt:   ft.UpdatedAt,
		DeletedAt:   ft.DeletedAt,
	}, nil
}

// ==============================
// Mutations
// ==============================

func (s *CustomerTypeService) CreateCustomerType(req *models.CustomerTypeCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.CustomerType, error) {
	_ = ctx; _ = userInfo

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	name := norm(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	if len(name) > 100 {
		return nil, errors.New("name exceeds max length")
	}
	color := norm(req.Color)

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Uniqueness check in-tx
	if _, err := s.CustomerTypeRepository.FindByName(tx, name); err == nil {
		tx.Rollback()
		return nil, errors.New("customer type name already exists")
	} else if err != nil && err != repositories.ErrCustomerTypeNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("check name failed: %w", err)
	}

	ft := &models.CustomerType{
		ID:          uuid.New(),
		Name:        name,
		Color:       color,
		Description: req.Description,
	}

	created, err := s.CustomerTypeRepository.Insert(tx, ft)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("customer type name already exists")
		}
		return nil, fmt.Errorf("insert customer type failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return created, nil
}

func (s *CustomerTypeService) UpdateCustomerType(idStr string, upd *models.CustomerTypeCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.CustomerType, error) {
	_ = ctx; _ = userInfo
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

	ft, err := s.CustomerTypeRepository.FindById(tx, idStr, false)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, repositories.ErrCustomerTypeNotFound) {
			return nil, errors.New("customer type not found")
		}
		return nil, fmt.Errorf("find customer type failed: %w", err)
	}

	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if len(newName) > 100 {
			tx.Rollback()
			return nil, errors.New("name exceeds max length")
		}
		if newName != ft.Name {
			if ex, err := s.CustomerTypeRepository.FindByName(tx, newName); err == nil && ex.ID != ft.ID {
				tx.Rollback()
				return nil, errors.New("customer type name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrCustomerTypeNotFound) {
				tx.Rollback()
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

	updated, err := s.CustomerTypeRepository.Update(tx, ft)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("customer type name already exists")
		}
		return nil, fmt.Errorf("update customer type failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return updated, nil
}

func (s *CustomerTypeService) DeleteCustomerTypes(in *models.CustomerTypeIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return errors.New("customerTypeIds cannot be empty")
	}
	isHard := in.IsHardDelete == "hardDelete"

	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for customer type %v: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		if _, err := s.CustomerTypeRepository.FindById(tx, id.String(), true); err != nil {
			tx.Rollback()
			if err == repositories.ErrCustomerTypeNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("CustomerType not found: %v\n", id)
				continue
			}
			log.Printf("Error finding customer type %v: %v\n", id, err)
			return errors.New("error finding customer type")
		}

		if err := s.CustomerTypeRepository.Delete(tx, id.String(), isHard); err != nil {
			tx.Rollback()
			log.Printf("Error deleting customer type %v: %v\n", id, err)
			return errors.New("error deleting customer type")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for customer type %v: %v\n", id, err)
			return errors.New("error committing delete")
		}
	}
	return nil
}

func (s *CustomerTypeService) RestoreCustomerTypes(in *models.CustomerTypeRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.CustomerType, error) {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return nil, errors.New("customerTypeIds cannot be empty")
	}

	var restored []models.CustomerType
	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for customer type restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		res, err := s.CustomerTypeRepository.Restore(tx, id.String())
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrCustomerTypeNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("CustomerType not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error restoring customer type %v: %v\n", id, err)
			return nil, errors.New("error restoring customer type")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing customer type restore %v: %v\n", id, err)
			return nil, errors.New("error committing customer type restore")
		}

		ft, ferr := s.CustomerTypeRepository.FindById(nil, res.ID.String(), true)
		if ferr != nil {
			log.Printf("Error fetching restored customer type %v: %v\n", id, ferr)
			continue
		}
		restored = append(restored, *ft)
	}
	return restored, nil
}
