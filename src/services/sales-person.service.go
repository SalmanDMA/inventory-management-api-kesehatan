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

type SalesPersonService struct {
	SalesPersonRepository   repositories.SalesPersonRepository
}

func NewSalesPersonService(salesPersonRepo repositories.SalesPersonRepository) *SalesPersonService {
	return &SalesPersonService{
		SalesPersonRepository:   salesPersonRepo,
	}
}

func (service *SalesPersonService) GetAllSalesPersons(userInfo *models.User) ([]models.ResponseGetSalesPerson, error) {
	salesPersons, err := service.SalesPersonRepository.FindAll()
	if err != nil {
		return nil, err
	}

	salesPersonsResponse := []models.ResponseGetSalesPerson{}
	for _, salesPerson := range salesPersons {
		salesPersonsResponse = append(salesPersonsResponse, models.ResponseGetSalesPerson{
			ID:          salesPerson.ID,
			Name:        salesPerson.Name,
			Email:       salesPerson.Email,
			Address:     salesPerson.Address,
			Phone:       salesPerson.Phone,
			HireDate: 			salesPerson.HireDate,
			Assignments: salesPerson.Assignments,
			CreatedAt:   salesPerson.CreatedAt,
			UpdatedAt:   salesPerson.UpdatedAt,
			DeletedAt:   salesPerson.DeletedAt,
		})
	}

	return salesPersonsResponse, nil
}

func (service *SalesPersonService) GetAllSalesPersonsPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.SalesPersonPaginatedResponse, error) {
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

	salesPersons, totalCount, err := service.SalesPersonRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	salesPersonsResponse := []models.ResponseGetSalesPerson{}
	for _, salesPerson := range salesPersons {
		salesPersonsResponse = append(salesPersonsResponse, models.ResponseGetSalesPerson{
			ID:          salesPerson.ID,
			Name:        salesPerson.Name,
			Email:       salesPerson.Email,
			Address:     salesPerson.Address,
			Phone:       salesPerson.Phone,
			HireDate: 			salesPerson.HireDate,
			Assignments: salesPerson.Assignments,
			CreatedAt:   salesPerson.CreatedAt,
			UpdatedAt:   salesPerson.UpdatedAt,
			DeletedAt:   salesPerson.DeletedAt,
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

	return &models.SalesPersonPaginatedResponse{
		Data:       salesPersonsResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *SalesPersonService) GetSalesPersonByID(id string, userInfo *models.User) (*models.ResponseGetSalesPerson, error) {
	salesPerson, err := service.SalesPersonRepository.FindById(id, false)
	if err != nil {
		return nil, err
	}

	return &models.ResponseGetSalesPerson{
		ID:          salesPerson.ID,
		Name:        salesPerson.Name,
		Email:       salesPerson.Email,
		Address:     salesPerson.Address,
		Phone:       salesPerson.Phone,
		HireDate: 			salesPerson.HireDate,
		Assignments: salesPerson.Assignments,
		CreatedAt:   salesPerson.CreatedAt,
		UpdatedAt:   salesPerson.UpdatedAt,
		DeletedAt:   salesPerson.DeletedAt,
	}, nil
}

func (service *SalesPersonService) CreateSalesPerson(
	salesPersonRequest *models.SalesPersonCreateRequest,
	ctx *fiber.Ctx,
	userInfo *models.User,
) (*models.SalesPerson, error) {
	_ = ctx
	_ = userInfo

	norm := func(v string) string { return strings.TrimSpace(v) }

	if norm(salesPersonRequest.Name) == "" {
		return nil, errors.New("name is required")
	}

	var emailPtr *string
	if salesPersonRequest.Email != nil {
		e := norm(*salesPersonRequest.Email)
		if e == "" {
			emailPtr = nil
		} else {
			if !repositories.IsValidEmail(e) {
				return nil, errors.New("invalid email")
			}
			emailPtr = &e
		}
	}

	if emailPtr != nil {
		if sp, err := service.SalesPersonRepository.FindByEmail(*emailPtr); err == nil && sp != nil {
			return nil, errors.New("email sales person already exists")
		} else if err != nil && err != repositories.ErrSalesPersonNotFound {
			return nil, fmt.Errorf("error checking email: %w", err)
		}
	}

	newSP := &models.SalesPerson{
		ID:       uuid.New(),
		Name:     norm(salesPersonRequest.Name),
		Email:    emailPtr,                  
		Phone:    salesPersonRequest.Phone,  
		Address:  salesPersonRequest.Address,
		HireDate: salesPersonRequest.HireDate,
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	if err := tx.Create(newSP).Error; err != nil {
		_ = tx.Rollback()
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, errors.New("email sales person already exists")
		}
		return nil, fmt.Errorf("error creating sales person: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	res, err := service.SalesPersonRepository.FindById(newSP.ID.String(), true)
	if err != nil {
		log.Printf("warning: created but fetch failed: %v", err)
		return newSP, nil
	}
	return res, nil
}

func (service *SalesPersonService) UpdateSalesPerson(
	salesPersonRequest *models.SalesPersonUpdateRequest,
	salesPersonId string,
	ctx *fiber.Ctx,
	userInfo *models.User,
) (*models.SalesPerson, error) {
	_ = ctx
	_ = userInfo

	norm := func(v string) string { return strings.TrimSpace(v) }

	existing, err := service.SalesPersonRepository.FindById(salesPersonId, true)
	if err != nil {
		return nil, fmt.Errorf("error finding sales person: %w", err)
	}

	if salesPersonRequest.Email != nil {
		var newEmailPtr *string
		e := norm(*salesPersonRequest.Email)
		if e == "" {
			newEmailPtr = nil
		} else {
			if !repositories.IsValidEmail(e) {
				return nil, errors.New("invalid email")
			}
			newEmailPtr = &e
		}

		cur := ""
		if existing.Email != nil {
			cur = *existing.Email
		}
		next := ""
		if newEmailPtr != nil {
			next = *newEmailPtr
		}

		if next != cur {
			if newEmailPtr != nil {
				if sp, err := service.SalesPersonRepository.FindByEmail(*newEmailPtr); err == nil && sp != nil && sp.ID != existing.ID {
					return nil, errors.New("email sales person already exists")
				} else if err != nil && err != repositories.ErrSalesPersonNotFound {
					return nil, fmt.Errorf("error checking existing email: %w", err)
				}
			}
			existing.Email = newEmailPtr 
		}
	}

	if s := norm(salesPersonRequest.Name); s != "" && s != existing.Name {
		existing.Name = s
	}

	if salesPersonRequest.Phone != nil {
		p := norm(*salesPersonRequest.Phone)
		if p == "" {
			existing.Phone = nil
		} else {
			existing.Phone = &p
		}
	}

	if salesPersonRequest.Address != nil {
		a := norm(*salesPersonRequest.Address)
		if a == "" {
			existing.Address = nil
		} else {
			existing.Address = &a
		}
	}

	if salesPersonRequest.HireDate != nil {
		existing.HireDate = salesPersonRequest.HireDate
	}

	updated, err := service.SalesPersonRepository.Update(existing)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, errors.New("email sales person already exists")
		}
		return nil, fmt.Errorf("error updating sales person: %w", err)
	}
	return updated, nil
}

func (service *SalesPersonService) DeleteSalesPersons(salesPersonRequest *models.SalesPersonIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, salesPersonId := range salesPersonRequest.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for salesPerson %v: %v\n", salesPersonId, tx.Error)
			return errors.New("error beginning transaction")
		}

		_, err := service.SalesPersonRepository.FindById(salesPersonId.String(), false)
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrSalesPersonNotFound {
				log.Printf("SalesPerson not found: %v\n", salesPersonId)
				continue
			}
			log.Printf("Error finding salesPerson %v: %v\n", salesPersonId, err)
			return errors.New("error finding salesPerson")
		}

		if salesPersonRequest.IsHardDelete == "hardDelete" {
			if err := tx.Unscoped().Delete(&models.SalesPerson{}, "id = ?", salesPersonId).Error; err != nil {
				tx.Rollback()
				log.Printf("Error hard deleting sales person %v: %v\n", salesPersonId, err)
				return errors.New("error hard deleting sales person")
			}

			if err := tx.Commit().Error; err != nil {
				log.Printf("Error committing hard delete for sales person %v: %v\n", salesPersonId, err)
				return errors.New("error committing hard delete")
			}
		} else {
			if err := tx.Delete(&models.SalesPerson{}, "id = ?", salesPersonId).Error; err != nil {
				tx.Rollback()
				log.Printf("Error soft deleting sales person %v: %v\n", salesPersonId, err)
				return errors.New("error soft deleting sales person")
			}

			if err := tx.Commit().Error; err != nil {
				log.Printf("Error committing soft delete for sales person %v: %v\n", salesPersonId, err)
				return errors.New("error committing soft delete")
			}
		}
	}

	return nil
}

func (service *SalesPersonService) RestoreSalesPersons(salesPersonRequest *models.SalesPersonRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.SalesPerson, error) {
	var restoredSalesPersons []models.SalesPerson

	for _, salesPersonId := range salesPersonRequest.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for sales person restore %v: %v\n", salesPersonId, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		result := tx.Model(&models.SalesPerson{}).Unscoped().Where("id = ?", salesPersonId).Update("deleted_at", nil)
		if result.Error != nil {
			tx.Rollback()
			log.Printf("Error restoring sales person %v: %v\n", salesPersonId, result.Error)
			return nil, errors.New("error restoring sales person")
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			log.Printf("SalesPerson not found for restore: %v\n", salesPersonId)
			continue
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing sales person restore %v: %v\n", salesPersonId, err)
			return nil, errors.New("error committing sales person restore")
		}

		restoredSalesPerson, err := service.SalesPersonRepository.FindById(salesPersonId.String(), true)
		if err != nil {
			log.Printf("Error fetching restored sales person %v: %v\n", salesPersonId, err)
			continue
		}

		restoredSalesPersons = append(restoredSalesPersons, *restoredSalesPerson)
	}

	return restoredSalesPersons, nil
}