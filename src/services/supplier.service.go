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

type SupplierService struct {
	SupplierRepository repositories.SupplierRepository
}

func NewSupplierService(supplierRepo repositories.SupplierRepository) *SupplierService {
	return &SupplierService{
		SupplierRepository: supplierRepo,
	}
}

func (service *SupplierService) GetAllSuppliers() ([]models.ResponseGetSupplier, error) {
	suppliers, err := service.SupplierRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var suppliersResponse []models.ResponseGetSupplier
	for _, supplier := range suppliers {
		suppliersResponse = append(suppliersResponse, models.ResponseGetSupplier{
			ID:            supplier.ID,
			Name:          supplier.Name,
			Code:          supplier.Code,
			Email:         supplier.Email,
			Phone:         supplier.Phone,
			Address:       supplier.Address,
			ContactPerson: supplier.ContactPerson,
			CreatedAt:     supplier.CreatedAt,
			UpdatedAt:     supplier.UpdatedAt,
			DeletedAt:     supplier.DeletedAt,
		})
	}

	return suppliersResponse, nil
}

func (service *SupplierService) GetAllSuppliersPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.SupplierPaginatedResponse, error) {
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

	suppliers, totalCount, err := service.SupplierRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	suppliersResponse := []models.ResponseGetSupplier{}
	for _, supplier := range suppliers {
		suppliersResponse = append(suppliersResponse, models.ResponseGetSupplier{
			ID:            supplier.ID,
			Name:          supplier.Name,
			Code:          supplier.Code,
			Email:         supplier.Email,
			Phone:         supplier.Phone,
			Address:       supplier.Address,
			ContactPerson: supplier.ContactPerson,
			CreatedAt:     supplier.CreatedAt,
			UpdatedAt:     supplier.UpdatedAt,
			DeletedAt:     supplier.DeletedAt,
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

	return &models.SupplierPaginatedResponse{
		Data:       suppliersResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *SupplierService) GetSupplierByID(supplierId string) (*models.ResponseGetSupplier, error) {
	supplier, err := service.SupplierRepository.FindById(supplierId, false)
	if err != nil {
		return nil, err
	}

	return &models.ResponseGetSupplier{
		ID:            supplier.ID,
		Name:          supplier.Name,
		Code:          supplier.Code,
		Email:         supplier.Email,
		Phone:         supplier.Phone,
		Address:       supplier.Address,
		ContactPerson: supplier.ContactPerson,
		CreatedAt:     supplier.CreatedAt,
		UpdatedAt:     supplier.UpdatedAt,
		DeletedAt:     supplier.DeletedAt,
	}, nil
}

func (service *SupplierService) CreateSupplier(req *models.SupplierCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Supplier, error) {
	_ = ctx
	_ = userInfo

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	name := norm(req.Name)
	code := norm(req.Code)

	if name == "" || code == "" {
		return nil, errors.New("name and code are required")
	}

	if len(name) > 150 || len(code) > 30 {
		return nil, errors.New("name/code exceeds max length")
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email == "" {
			req.Email = nil
		} else if !repositories.IsValidEmail(email) {
			return nil, errors.New("invalid email")
		} else {
			req.Email = &email
		}
	}
	if req.Phone != nil {
		p := strings.TrimSpace(*req.Phone)
		if p == "" { 
			req.Phone = nil
		} else {
			req.Phone = &p
		}
	}
	if req.Address != nil {
		a := strings.TrimSpace(*req.Address)
		if a == "" {
			req.Address = nil
		} else {
			req.Address = &a
		}
	}
	if req.ContactPerson != nil {
		cp := strings.TrimSpace(*req.ContactPerson)
		if cp == "" {
			req.ContactPerson = nil
		} else {
			req.ContactPerson = &cp
		}
	}

	if _, err := service.SupplierRepository.FindByName(name); err == nil {
		return nil, errors.New("supplier with this name already exists")
	} else if err != repositories.ErrSupplierNotFound {
		return nil, errors.New("error checking supplier: " + err.Error())
	}

	if _, err := service.SupplierRepository.FindByCode(code); err == nil {
		return nil, errors.New("supplier with this code already exists")
	} else if err != repositories.ErrSupplierNotFound {
		return nil, errors.New("error checking supplier: " + err.Error())
	}

	newSupplier := &models.Supplier{
		ID:            uuid.New(),
		Name:          req.Name,
		Code:          req.Code,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		ContactPerson: req.ContactPerson,
	}

	result, err := service.SupplierRepository.Insert(newSupplier)
	if err != nil {
		return nil, fmt.Errorf("error creating supplier: %w", err)
	}

	return result, nil
}

func (service *SupplierService) UpdateSupplier(idStr string, upd *models.SupplierUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Supplier, error) {
	_ = ctx
	_ = userInfo

	existing, err := service.SupplierRepository.FindById(idStr, false)
	if err != nil {
		if errors.Is(err, repositories.ErrSupplierNotFound) {
			return nil, errors.New("supplier not found")
		}
		return nil, fmt.Errorf("find supplier failed: %w", err)
	}

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	if strings.TrimSpace(upd.Code) != "" {
		newCode := norm(upd.Code)
		if len(newCode) > 30 {
			return nil, errors.New("code exceeds max length")
		}
		if newCode != existing.Code {
			if ex, err := service.SupplierRepository.FindByCode(newCode); err == nil && ex.ID != existing.ID {
				return nil, errors.New("supplier code already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrSalesPersonNotFound) {
				return nil, fmt.Errorf("check code failed: %w", err)
			}
			existing.Code = newCode
		}
	}

	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if len(newName) > 150 {
			return nil, errors.New("name exceeds max length")
		}
		if newName != existing.Name {
			if ex, err := service.SupplierRepository.FindByName(newName); err == nil && ex.ID != existing.ID {
				return nil, errors.New("supplier name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrSalesPersonNotFound) {
				return nil, fmt.Errorf("check name failed: %w", err)
			}
			existing.Name = newName
		}
	}

	if upd.Address != nil {
		a := strings.TrimSpace(*upd.Address)
		if a == "" {
			existing.Address = nil
		} else {
			existing.Address = &a
		}
	}
	if upd.Phone != nil {
		p := strings.TrimSpace(*upd.Phone)
		if p == "" {
			existing.Phone = nil
		} else {
			existing.Phone = &p
		}
	}
	if upd.Email != nil {
		e := strings.TrimSpace(*upd.Email)
		if e == "" {
			existing.Email = nil
		} else if !repositories.IsValidEmail(e) {
			return nil, errors.New("invalid email")
		} else {
			existing.Email = &e
		}
	}

	if upd.ContactPerson != nil {
		cp := strings.TrimSpace(*upd.ContactPerson)
		if cp == "" {
			existing.ContactPerson = nil
		} else {
			existing.ContactPerson = &cp
		}
	}

	result, err := service.SupplierRepository.Update(existing)
	if err != nil {
		return nil, fmt.Errorf("error updating supplier: %w", err)
	}

	return result, nil
}

func (service *SupplierService) DeleteSuppliers(supplierRequest *models.SupplierIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, supplierId := range supplierRequest.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for supplier %v: %v\n", supplierId, tx.Error)
			return errors.New("error beginning transaction")
		}

		_, err := service.SupplierRepository.FindById(supplierId.String(), true)
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrSupplierNotFound {
				log.Printf("Supplier not found: %v\n", supplierId)
				continue
			}
			log.Printf("Error finding supplier %v: %v\n", supplierId, err)
			return errors.New("error finding supplier")
		}

		if supplierRequest.IsHardDelete == "hardDelete" {
				if err := tx.Unscoped().Delete(&models.Supplier{}, "id = ?", supplierId).Error; err != nil {
					tx.Rollback()
					log.Printf("Error hard deleting supplier %v: %v\n", supplierId, err)
					return errors.New("error hard deleting supplier")
				}

				if err := tx.Commit().Error; err != nil {
					log.Printf("Error committing hard delete for supplier %v: %v\n", supplierId, err)
					return errors.New("error committing hard delete")
				}
		} else {
			if err := tx.Delete(&models.Supplier{}, "id = ?", supplierId).Error; err != nil {
				tx.Rollback()
				log.Printf("Error soft deleting supplier %v: %v\n", supplierId, err)
				return errors.New("error soft deleting supplier")
			}

			if err := tx.Commit().Error; err != nil {
				log.Printf("Error committing soft delete for supplier %v: %v\n", supplierId, err)
				return errors.New("error committing soft delete")
			}
		}
	}

	return nil
}

func (service *SupplierService) RestoreSuppliers(supplierRequest *models.SupplierRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Supplier, error) {
	var restoredSuppliers []models.Supplier

	for _, supplierId := range supplierRequest.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for supplier restore %v: %v\n", supplierId, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		result := tx.Model(&models.Supplier{}).Unscoped().Where("id = ?", supplierId).Update("deleted_at", nil)
		if result.Error != nil {
			tx.Rollback()
			log.Printf("Error restoring supplier %v: %v\n", supplierId, result.Error)
			return nil, errors.New("error restoring supplier")
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			log.Printf("Supplier not found for restore: %v\n", supplierId)
			continue
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing supplier restore %v: %v\n", supplierId, err)
			return nil, errors.New("error committing supplier restore")
		}

		restoredSupplier, err := service.SupplierRepository.FindById(supplierId.String(), true)
		if err != nil {
			log.Printf("Error fetching restored supplier %v: %v\n", supplierId, err)
			continue
		}

		restoredSuppliers = append(restoredSuppliers, *restoredSupplier)
	}

	return restoredSuppliers, nil
}