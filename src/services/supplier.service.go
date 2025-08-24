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

func (s *SupplierService) GetAllSuppliers() ([]models.ResponseGetSupplier, error) {
	suppliers, err := s.SupplierRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResponseGetSupplier, 0, len(suppliers))
	for _, sp := range suppliers {
		out = append(out, models.ResponseGetSupplier{
			ID:            sp.ID,
			Name:          sp.Name,
			Code:          sp.Code,
			Email:         sp.Email,
			Phone:         sp.Phone,
			Address:       sp.Address,
			ContactPerson: sp.ContactPerson,
			CreatedAt:     sp.CreatedAt,
			UpdatedAt:     sp.UpdatedAt,
			DeletedAt:     sp.DeletedAt,
		})
	}
	return out, nil
}

func (s *SupplierService) GetAllSuppliersPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.SupplierPaginatedResponse, error) {
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

	rows, total, err := s.SupplierRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	data := make([]models.ResponseGetSupplier, 0, len(rows))
	for _, sp := range rows {
		data = append(data, models.ResponseGetSupplier{
			ID:            sp.ID,
			Name:          sp.Name,
			Code:          sp.Code,
			Email:         sp.Email,
			Phone:         sp.Phone,
			Address:       sp.Address,
			ContactPerson: sp.ContactPerson,
			CreatedAt:     sp.CreatedAt,
			UpdatedAt:     sp.UpdatedAt,
			DeletedAt:     sp.DeletedAt,
		})
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.SupplierPaginatedResponse{
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

func (s *SupplierService) GetSupplierByID(supplierId string) (*models.ResponseGetSupplier, error) {
	sp, err := s.SupplierRepository.FindById(nil, supplierId, false)
	if err != nil {
		return nil, err
	}

	return &models.ResponseGetSupplier{
		ID:            sp.ID,
		Name:          sp.Name,
		Code:          sp.Code,
		Email:         sp.Email,
		Phone:         sp.Phone,
		Address:       sp.Address,
		ContactPerson: sp.ContactPerson,
		CreatedAt:     sp.CreatedAt,
		UpdatedAt:     sp.UpdatedAt,
		DeletedAt:     sp.DeletedAt,
	}, nil
}

func (s *SupplierService) CreateSupplier(req *models.SupplierCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Supplier, error) {
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

	// normalisasi optional fields
	if req.Email != nil {
		e := strings.TrimSpace(*req.Email)
		if e == "" {
			req.Email = nil
		} else if !repositories.IsValidEmail(e) {
			return nil, errors.New("invalid email")
		} else {
			req.Email = &e
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

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Uniqueness check dalam tx
	if _, err := s.SupplierRepository.FindByName(tx, name); err == nil {
		tx.Rollback()
		return nil, errors.New("supplier with this name already exists")
	} else if !errors.Is(err, repositories.ErrSupplierNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("check supplier name failed: %w", err)
	}

	if _, err := s.SupplierRepository.FindByCode(tx, code); err == nil {
		tx.Rollback()
		return nil, errors.New("supplier with this code already exists")
	} else if !errors.Is(err, repositories.ErrSupplierNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("check supplier code failed: %w", err)
	}

	entity := &models.Supplier{
		ID:            uuid.New(),
		Name:          name, // simpan versi normalized untuk unik
		Code:          code,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		ContactPerson: req.ContactPerson,
	}

	inserted, err := s.SupplierRepository.Insert(tx, entity)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("supplier already exists")
		}
		return nil, fmt.Errorf("error creating supplier: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// fetch ulang (di luar tx) kalau butuh preload
	out, err := s.SupplierRepository.FindById(nil, inserted.ID.String(), false)
	if err != nil {
		return inserted, nil
	}
	return out, nil
}

func (s *SupplierService) UpdateSupplier(idStr string, upd *models.SupplierUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Supplier, error) {
	_ = ctx
	_ = userInfo

	// ambil dulu di luar tx
	cur, err := s.SupplierRepository.FindById(nil, idStr, false)
	if err != nil {
		if errors.Is(err, repositories.ErrSupplierNotFound) {
			return nil, errors.New("supplier not found")
		}
		return nil, fmt.Errorf("find supplier failed: %w", err)
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

	// code
	if strings.TrimSpace(upd.Code) != "" {
		newCode := norm(upd.Code)
		if len(newCode) > 30 {
			tx.Rollback()
			return nil, errors.New("code exceeds max length")
		}
		if newCode != cur.Code {
			if ex, err := s.SupplierRepository.FindByCode(tx, newCode); err == nil && ex.ID != cur.ID {
				tx.Rollback()
				return nil, errors.New("supplier code already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrSupplierNotFound) {
				tx.Rollback()
				return nil, fmt.Errorf("check code failed: %w", err)
			}
			cur.Code = newCode
		}
	}

	// name
	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if len(newName) > 150 {
			tx.Rollback()
			return nil, errors.New("name exceeds max length")
		}
		if newName != cur.Name {
			if ex, err := s.SupplierRepository.FindByName(tx, newName); err == nil && ex.ID != cur.ID {
				tx.Rollback()
				return nil, errors.New("supplier name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrSupplierNotFound) {
				tx.Rollback()
				return nil, fmt.Errorf("check name failed: %w", err)
			}
			cur.Name = newName
		}
	}

	// optional fields
	if upd.Address != nil {
		a := strings.TrimSpace(*upd.Address)
		if a == "" {
			cur.Address = nil
		} else {
			cur.Address = &a
		}
	}
	if upd.Phone != nil {
		p := strings.TrimSpace(*upd.Phone)
		if p == "" {
			cur.Phone = nil
		} else {
			cur.Phone = &p
		}
	}
	if upd.Email != nil {
		e := strings.TrimSpace(*upd.Email)
		if e == "" {
			cur.Email = nil
		} else if !repositories.IsValidEmail(e) {
			tx.Rollback()
			return nil, errors.New("invalid email")
		} else {
			cur.Email = &e
		}
	}
	if upd.ContactPerson != nil {
		cp := strings.TrimSpace(*upd.ContactPerson)
		if cp == "" {
			cur.ContactPerson = nil
		} else {
			cur.ContactPerson = &cp
		}
	}

	updated, err := s.SupplierRepository.Update(tx, cur)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("supplier already exists")
		}
		return nil, fmt.Errorf("error updating supplier: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	out, err := s.SupplierRepository.FindById(nil, updated.ID.String(), false)
	if err != nil {
		return updated, nil
	}
	return out, nil
}

func (s *SupplierService) DeleteSuppliers(req *models.SupplierIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx
	_ = userInfo

	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for supplier %v: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		if _, err := s.SupplierRepository.FindById(tx, id.String(), true); err != nil {
			tx.Rollback()
			if errors.Is(err, repositories.ErrSupplierNotFound) {
				log.Printf("Supplier not found: %v\n", id)
				continue
			}
			log.Printf("Error finding supplier %v: %v\n", id, err)
			return errors.New("error finding supplier")
		}

		if err := s.SupplierRepository.Delete(tx, id.String(), req.IsHardDelete == "hardDelete"); err != nil {
			tx.Rollback()
			log.Printf("Error deleting supplier %v: %v\n", id, err)
			return errors.New("error deleting supplier")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for supplier %v: %v\n", id, err)
			return errors.New("error committing delete")
		}
	}
	return nil
}

func (s *SupplierService) RestoreSuppliers(req *models.SupplierRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Supplier, error) {
	_ = ctx
	_ = userInfo

	var restored []models.Supplier
	for _, id := range req.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for supplier restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		sp := &models.Supplier{ID: id}
		if _, err := s.SupplierRepository.Restore(tx, id.String()); err != nil {
			tx.Rollback()
			if errors.Is(err, repositories.ErrSupplierNotFound) {
				log.Printf("Supplier not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error restoring supplier %v: %v\n", id, err)
			return nil, errors.New("error restoring supplier")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing supplier restore %v: %v\n", id, err)
			return nil, errors.New("error committing supplier restore")
		}

		row, err := s.SupplierRepository.FindById(nil, id.String(), true)
		if err != nil {
			restored = append(restored, *sp)
			continue
		}
		restored = append(restored, *row)
	}
	return restored, nil
}
