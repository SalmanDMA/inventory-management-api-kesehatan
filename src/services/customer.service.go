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

type CustomerService struct {
	CustomerRepository repositories.CustomerRepository
}

func NewCustomerService(customerRepo repositories.CustomerRepository) *CustomerService {
	return &CustomerService{
		CustomerRepository: customerRepo,
	}
}

// ==============================
// Reads (tanpa transaksi)
// ==============================

func (s *CustomerService) GetAllCustomers(userInfo *models.User) ([]models.ResponseGetCustomer, error) {
	_ = userInfo

	customers, err := s.CustomerRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetCustomer, 0, len(customers))
	for _, f := range customers {
		resp = append(resp, models.ResponseGetCustomer{
			ID:            f.ID,
			Name:          f.Name,
			Nomor:          f.Nomor,
			CustomerTypeID:f.CustomerTypeID,
			AreaID:        f.AreaID,
			Address:       f.Address,
			Phone:         f.Phone,
			Email:         f.Email,
			Latitude:      f.Latitude,
			Longitude:     f.Longitude,
			CustomerType:  f.CustomerType,
			Area:          f.Area,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
			DeletedAt:     f.DeletedAt,
		})
	}
	return resp, nil
}

func (s *CustomerService) GetAllCustomersPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.CustomerPaginatedResponse, error) {
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

	list, totalCount, err := s.CustomerRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	data := make([]models.ResponseGetCustomer, 0, len(list))
	for _, f := range list {
		data = append(data, models.ResponseGetCustomer{
			ID:            f.ID,
			Name:          f.Name,
			Nomor:          f.Nomor,
			CustomerTypeID:f.CustomerTypeID,
			AreaID:        f.AreaID,
			Address:       f.Address,
			Phone:         f.Phone,
			Email:         f.Email,
			Latitude:      f.Latitude,
			Longitude:     f.Longitude,
			CustomerType:  f.CustomerType,
			Area:          f.Area,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
			DeletedAt:     f.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.CustomerPaginatedResponse{
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

func (s *CustomerService) GetCustomerByID(customerId string) (*models.ResponseGetCustomer, error) {
	f, err := s.CustomerRepository.FindById(nil, customerId, false)
	if err != nil {
		return nil, err
	}

	return &models.ResponseGetCustomer{
		ID:            f.ID,
		Name:          f.Name,
		Nomor:          f.Nomor,
		CustomerTypeID:f.CustomerTypeID,
		AreaID:        f.AreaID,
		Address:       f.Address,
		Phone:         f.Phone,
		Email:         f.Email,
		Latitude:      f.Latitude,
		Longitude:     f.Longitude,
		CustomerType:  f.CustomerType,
		Area:          f.Area,
		CreatedAt:     f.CreatedAt,
		UpdatedAt:     f.UpdatedAt,
		DeletedAt:     f.DeletedAt,
	}, nil
}

// ==============================
// Mutations (transaction-aware via configs.DB.Begin())
// ==============================

func (s *CustomerService) CreateCustomer(req *models.CustomerCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Customer, error) {
	_ = ctx; _ = userInfo

	normalize := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	name := normalize(req.Name)
	code := normalize(req.Nomor)

	if name == "" || code == "" {
		return nil, errors.New("name and code are required")
	}
	if len(name) > 150 || len(code) > 30 {
		return nil, errors.New("name/code exceeds max length")
	}
	if req.CustomerTypeID == uuid.Nil {
		return nil, errors.New("customer_type_id is required")
	}
	if req.AreaID == uuid.Nil {
		return nil, errors.New("area_id is required")
	}
	if req.Latitude != nil && (*req.Latitude < -90 || *req.Latitude > 90) {
		return nil, errors.New("latitude out of range")
	}
	if req.Longitude != nil && (*req.Longitude < -180 || *req.Longitude > 180) {
		return nil, errors.New("longitude out of range")
	}

	// sanitize optional strings
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

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// uniqueness checks (tx-aware)
	if _, err := s.CustomerRepository.FindByName(tx, name); err == nil {
		tx.Rollback()
		return nil, errors.New("customer name already exists")
	} else if !errors.Is(err, repositories.ErrCustomerNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("check name failed: %w", err)
	}
	if _, err := s.CustomerRepository.FindByNomor(tx, code); err == nil {
		tx.Rollback()
		return nil, errors.New("customer code already exists")
	} else if !errors.Is(err, repositories.ErrCustomerNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("check code failed: %w", err)
	}

	newFac := &models.Customer{
		ID:             uuid.New(),
		Name:           name,
		Nomor:           code,
		CustomerTypeID: req.CustomerTypeID,
		AreaID:         req.AreaID,
		Address:        req.Address,
		Phone:          req.Phone,
		Email:          req.Email,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
	}

	created, err := s.CustomerRepository.Insert(tx, newFac)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("customer name/code already exists")
		}
		return nil, fmt.Errorf("insert customer failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return created, nil
}

func (s *CustomerService) UpdateCustomer(idStr string, upd *models.CustomerCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Customer, error) {
	_ = ctx; _ = userInfo

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	existing, err := s.CustomerRepository.FindById(tx, idStr, false)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, repositories.ErrCustomerNotFound) {
			return nil, errors.New("customer not found")
		}
		return nil, fmt.Errorf("find customer failed: %w", err)
	}

	normalize := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	if strings.TrimSpace(upd.Nomor) != "" {
		newNomor := normalize(upd.Nomor)
		if len(newNomor) > 30 {
			tx.Rollback()
			return nil, errors.New("code exceeds max length")
		}
		if newNomor != existing.Nomor {
			if ex, err := s.CustomerRepository.FindByNomor(tx, newNomor); err == nil && ex.ID != existing.ID {
				tx.Rollback()
				return nil, errors.New("customer code already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrCustomerNotFound) {
				tx.Rollback()
				return nil, fmt.Errorf("check code failed: %w", err)
			}
			existing.Nomor = newNomor
		}
	}

	if strings.TrimSpace(upd.Name) != "" {
		newName := normalize(upd.Name)
		if len(newName) > 150 {
			tx.Rollback()
			return nil, errors.New("name exceeds max length")
		}
		if newName != existing.Name {
			if ex, err := s.CustomerRepository.FindByName(tx, newName); err == nil && ex.ID != existing.ID {
				tx.Rollback()
				return nil, errors.New("customer name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrCustomerNotFound) {
				tx.Rollback()
				return nil, fmt.Errorf("check name failed: %w", err)
			}
			existing.Name = newName
		}
	}

	if upd.CustomerTypeID != uuid.Nil {
		existing.CustomerTypeID = upd.CustomerTypeID
	}
	if upd.AreaID != uuid.Nil {
		existing.AreaID = upd.AreaID
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
			tx.Rollback()
			return nil, errors.New("invalid email")
		} else {
			existing.Email = &e
		}
	}
	if upd.Latitude != nil {
		if *upd.Latitude < -90 || *upd.Latitude > 90 {
			tx.Rollback()
			return nil, errors.New("latitude out of range")
		}
		existing.Latitude = upd.Latitude
	}
	if upd.Longitude != nil {
		if *upd.Longitude < -180 || *upd.Longitude > 180 {
			tx.Rollback()
			return nil, errors.New("longitude out of range")
		}
		existing.Longitude = upd.Longitude
	}

	updated, err := s.CustomerRepository.Update(tx, existing)
	if err != nil {
		tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("customer name/code already exists")
		}
		return nil, fmt.Errorf("update customer failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return updated, nil
}

func (s *CustomerService) DeleteCustomers(in *models.CustomerIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx; _ = userInfo

	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for customer %v: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		_, err := s.CustomerRepository.FindById(tx, id.String(), true)
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrCustomerNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Customer not found: %v\n", id)
				continue
			}
			log.Printf("Error finding customer %v: %v\n", id, err)
			return errors.New("error finding customer")
		}

		isHard := in.IsHardDelete == "hardDelete"
		if err := s.CustomerRepository.Delete(tx, id.String(), isHard); err != nil {
			tx.Rollback()
			log.Printf("Error deleting customer %v: %v\n", id, err)
			return errors.New("error deleting customer")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for customer %v: %v\n", id, err)
			return errors.New("error committing delete")
		}
	}
	return nil
}

func (s *CustomerService) RestoreCustomers(in *models.CustomerRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Customer, error) {
	_ = ctx; _ = userInfo

	var restored []models.Customer
	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for customer restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		res, err := s.CustomerRepository.Restore(tx, id.String())
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrCustomerNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Customer not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error restoring customer %v: %v\n", id, err)
			return nil, errors.New("error restoring customer")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing customer restore %v: %v\n", id, err)
			return nil, errors.New("error committing customer restore")
		}

		restoredFac, ferr := s.CustomerRepository.FindById(nil, res.ID.String(), true)
		if ferr != nil {
			log.Printf("Error fetching restored customer %v: %v\n", id, ferr)
			continue
		}
		restored = append(restored, *restoredFac)
	}
	return restored, nil
}

