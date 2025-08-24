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
	SalesPersonRepository repositories.SalesPersonRepository
}

func NewSalesPersonService(repo repositories.SalesPersonRepository) *SalesPersonService {
	return &SalesPersonService{SalesPersonRepository: repo}
}

func (s *SalesPersonService) GetAllSalesPersons(userInfo *models.User) ([]models.ResponseGetSalesPerson, error) {
	_ = userInfo
	rows, err := s.SalesPersonRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}
	out := make([]models.ResponseGetSalesPerson, 0, len(rows))
	for _, sp := range rows {
		out = append(out, models.ResponseGetSalesPerson{
			ID:          sp.ID,
			Name:        sp.Name,
			Email:       sp.Email,
			Address:     sp.Address,
			Phone:       sp.Phone,
			HireDate:    sp.HireDate,
			NPWP: 							sp.NPWP,
			Assignments: sp.Assignments,
			CreatedAt:   sp.CreatedAt,
			UpdatedAt:   sp.UpdatedAt,
			DeletedAt:   sp.DeletedAt,
		})
	}
	return out, nil
}

func (s *SalesPersonService) GetAllSalesPersonsPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.SalesPersonPaginatedResponse, error) {
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

	rows, total, err := s.SalesPersonRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	data := make([]models.ResponseGetSalesPerson, 0, len(rows))
	for _, sp := range rows {
		data = append(data, models.ResponseGetSalesPerson{
			ID:          sp.ID,
			Name:        sp.Name,
			Email:       sp.Email,
			Address:     sp.Address,
			Phone:       sp.Phone,
			HireDate:    sp.HireDate,
			NPWP: 							sp.NPWP,
			Assignments: sp.Assignments,
			CreatedAt:   sp.CreatedAt,
			UpdatedAt:   sp.UpdatedAt,
			DeletedAt:   sp.DeletedAt,
		})
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.SalesPersonPaginatedResponse{
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

func (s *SalesPersonService) GetSalesPersonByID(id string, userInfo *models.User) (*models.ResponseGetSalesPerson, error) {
	_ = userInfo
	sp, err := s.SalesPersonRepository.FindById(nil, id, false)
	if err != nil {
		return nil, err
	}
	return &models.ResponseGetSalesPerson{
		ID:          sp.ID,
		Name:        sp.Name,
		Email:       sp.Email,
		Address:     sp.Address,
		Phone:       sp.Phone,
		HireDate:    sp.HireDate,
		NPWP: 							sp.NPWP,
		Assignments: sp.Assignments,
		CreatedAt:   sp.CreatedAt,
		UpdatedAt:   sp.UpdatedAt,
		DeletedAt:   sp.DeletedAt,
	}, nil
}

func (s *SalesPersonService) CreateSalesPerson(req *models.SalesPersonCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.SalesPerson, error) {
	_ = ctx
	_ = userInfo

	norm := func(v string) string { return strings.TrimSpace(v) }
	if norm(req.Name) == "" {
		return nil, errors.New("name is required")
	}

	var emailPtr *string
	if req.Email != nil {
		e := norm(*req.Email)
		if e != "" {
			if !repositories.IsValidEmail(e) {
				return nil, errors.New("invalid email")
			}
			emailPtr = &e
		}
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

	// uniqueness check inside tx
	if emailPtr != nil {
		if sp, err := s.SalesPersonRepository.FindByEmail(tx, *emailPtr); err == nil && sp != nil {
			_ = tx.Rollback()
			return nil, errors.New("email sales person already exists")
		} else if err != nil && !errors.Is(err, repositories.ErrSalesPersonNotFound) {
			_ = tx.Rollback()
			return nil, fmt.Errorf("error checking email: %w", err)
		}
	}

	entity := &models.SalesPerson{
		ID:       uuid.New(),
		Name:     norm(req.Name),
		Email:    emailPtr,
		Phone:    req.Phone,
		Address:  req.Address,
		HireDate: req.HireDate,
		NPWP: 			req.NPWP,
	}

	created, err := s.SalesPersonRepository.Insert(tx, entity)
	if err != nil {
		_ = tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("email sales person already exists")
		}
		return nil, fmt.Errorf("error creating sales person: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	out, err := s.SalesPersonRepository.FindById(nil, created.ID.String(), true)
	if err != nil {
		log.Printf("warning: created but fetch failed: %v", err)
		return created, nil
	}
	return out, nil
}

func (s *SalesPersonService) UpdateSalesPerson(req *models.SalesPersonUpdateRequest, id string, ctx *fiber.Ctx, userInfo *models.User) (*models.SalesPerson, error) {
	_ = ctx
	_ = userInfo

	cur, err := s.SalesPersonRepository.FindById(nil, id, true)
	if err != nil {
		if errors.Is(err, repositories.ErrSalesPersonNotFound) {
			return nil, errors.New("sales person not found")
		}
		return nil, fmt.Errorf("error finding sales person: %w", err)
	}

	norm := func(v string) string { return strings.TrimSpace(v) }

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// email
	if req.Email != nil {
		var newEmailPtr *string
		e := norm(*req.Email)
		if e != "" {
			if !repositories.IsValidEmail(e) {
				_ = tx.Rollback()
				return nil, errors.New("invalid email")
			}
			newEmailPtr = &e
		}
		curStr := ""
		if cur.Email != nil {
			curStr = *cur.Email
		}
		nextStr := ""
		if newEmailPtr != nil {
			nextStr = *newEmailPtr
		}
		if nextStr != curStr {
			if newEmailPtr != nil {
				if sp, err := s.SalesPersonRepository.FindByEmail(tx, *newEmailPtr); err == nil && sp != nil && sp.ID != cur.ID {
					_ = tx.Rollback()
					return nil, errors.New("email sales person already exists")
				} else if err != nil && !errors.Is(err, repositories.ErrSalesPersonNotFound) {
					_ = tx.Rollback()
					return nil, fmt.Errorf("error checking existing email: %w", err)
				}
			}
			cur.Email = newEmailPtr
		}
	}

	if sName := norm(req.Name); sName != "" && sName != cur.Name {
		cur.Name = sName
	}

	if req.Phone != nil {
		p := norm(*req.Phone)
		if p == "" {
			cur.Phone = nil
		} else {
			cur.Phone = &p
		}
	}
	if req.Address != nil {
		a := norm(*req.Address)
		if a == "" {
			cur.Address = nil
		} else {
			cur.Address = &a
		}
	}
	if req.HireDate != nil {
		cur.HireDate = req.HireDate
	}
	if req.NPWP != nil {
		n := norm(*req.NPWP)
		if n == "" {
			cur.NPWP = nil
		} else {
			cur.NPWP = &n
		}
	}

	updated, err := s.SalesPersonRepository.Update(tx, cur)
	if err != nil {
		_ = tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("email sales person already exists")
		}
		return nil, fmt.Errorf("error updating sales person: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	out, err := s.SalesPersonRepository.FindById(nil, updated.ID.String(), true)
	if err != nil {
		return updated, nil
	}
	return out, nil
}

func (s *SalesPersonService) DeleteSalesPersons(req *models.SalesPersonIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx
	_ = userInfo

	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for salesPerson %v: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		if _, err := s.SalesPersonRepository.FindById(tx, id.String(), false); err != nil {
			tx.Rollback()
			if errors.Is(err, repositories.ErrSalesPersonNotFound) {
				log.Printf("SalesPerson not found: %v\n", id)
				continue
			}
			log.Printf("Error finding salesPerson %v: %v\n", id, err)
			return errors.New("error finding salesPerson")
		}

		if err := s.SalesPersonRepository.Delete(tx, id.String(), req.IsHardDelete == "hardDelete"); err != nil {
			tx.Rollback()
			log.Printf("Error deleting sales person %v: %v\n", id, err)
			return errors.New("error deleting sales person")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for sales person %v: %v\n", id, err)
			return errors.New("error committing delete")
		}
	}
	return nil
}

func (s *SalesPersonService) RestoreSalesPersons(req *models.SalesPersonRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.SalesPerson, error) {
	_ = ctx
	_ = userInfo

	var restored []models.SalesPerson
	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for sales person restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		sp := &models.SalesPerson{ID: id}
		if _, err := s.SalesPersonRepository.Restore(tx, id.String()); err != nil {
			tx.Rollback()
			if errors.Is(err, repositories.ErrSalesPersonNotFound) {
				log.Printf("SalesPerson not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error restoring sales person %v: %v\n", id, err)
			return nil, errors.New("error restoring sales person")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing sales person restore %v: %v\n", id, err)
			return nil, errors.New("error committing sales person restore")
		}

		row, err := s.SalesPersonRepository.FindById(nil, id.String(), true)
		if err != nil {
			restored = append(restored, *sp)
			continue
		}
		restored = append(restored, *row)
	}
	return restored, nil
}
