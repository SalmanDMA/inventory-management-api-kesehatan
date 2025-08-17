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

type FacilityService struct {
	FacilityRepository repositories.FacilityRepository
}

func NewFacilityService(facilityRepo repositories.FacilityRepository) *FacilityService {
	return &FacilityService{
		FacilityRepository: facilityRepo,
	}
}

func (service *FacilityService) GetAllFacilities(userInfo *models.User) ([]models.ResponseGetFacility, error) {
	facilities, err := service.FacilityRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var facilitiesResponse []models.ResponseGetFacility
	for _, facility := range facilities {
		facilitiesResponse = append(facilitiesResponse, models.ResponseGetFacility{
			ID:          facility.ID,
			Name:        facility.Name,
			Code:        facility.Code,
   FacilityTypeID: facility.FacilityTypeID,
			AreaID:      facility.AreaID,
			Address:     facility.Address,
			Phone:       facility.Phone,
			Email:       facility.Email,
		 Latitude:    facility.Latitude,
			Longitude:   facility.Longitude,
			FacilityType: facility.FacilityType,
			Area:         facility.Area,
			CreatedAt: 	facility.CreatedAt,
			UpdatedAt: 	facility.UpdatedAt,
			DeletedAt: 	facility.DeletedAt,
		})
	}

	return facilitiesResponse, nil
}

func (service *FacilityService) GetAllFacilitiesPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.FacilityPaginatedResponse, error) {
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

	facilities, totalCount, err := service.FacilityRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	facilitiesResponse := []models.ResponseGetFacility{}
	for _, facility := range facilities {
		facilitiesResponse = append(facilitiesResponse, models.ResponseGetFacility{
			ID:          facility.ID,
			Name:        facility.Name,
			Code:        facility.Code,
   FacilityTypeID: facility.FacilityTypeID,
			AreaID:      facility.AreaID,
			Address:     facility.Address,
			Phone:       facility.Phone,
			Email:       facility.Email,
		 Latitude:    facility.Latitude,
			Longitude:   facility.Longitude,
			FacilityType: facility.FacilityType,
			Area:         facility.Area,
			CreatedAt: 	facility.CreatedAt,
			UpdatedAt: 	facility.UpdatedAt,
			DeletedAt: 	facility.DeletedAt,
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

	return &models.FacilityPaginatedResponse{
		Data:       facilitiesResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *FacilityService) GetFacilityByID(facilityId string) (*models.ResponseGetFacility, error) {
	facility, err := service.FacilityRepository.FindById(facilityId, false)

	if err != nil {
		return nil, err
	}

	return &models.ResponseGetFacility{
		 ID:          facility.ID,
			Name:        facility.Name,
			Code:        facility.Code,
   FacilityTypeID: facility.FacilityTypeID,
			AreaID:      facility.AreaID,
			Address:     facility.Address,
			Phone:       facility.Phone,
			Email:       facility.Email,
		 Latitude:    facility.Latitude,
			Longitude:   facility.Longitude,
			FacilityType: facility.FacilityType,
			Area:         facility.Area,
			CreatedAt: 	facility.CreatedAt,
			UpdatedAt: 	facility.UpdatedAt,
			DeletedAt: 	facility.DeletedAt,
	}, nil
}

func (s *FacilityService) CreateFacility(req *models.FacilityCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Facility, error) {
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
	if req.FacilityTypeID == uuid.Nil {
		return nil, errors.New("facility_type_id is required")
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

	if _, err := s.FacilityRepository.FindByName(name); err == nil {
		return nil, errors.New("facility name already exists")
	} else if !errors.Is(err, repositories.ErrFacilityNotFound) {
		return nil, fmt.Errorf("check name failed: %w", err)
	}
	if _, err := s.FacilityRepository.FindByCode(code); err == nil {
		return nil, errors.New("facility code already exists")
	} else if !errors.Is(err, repositories.ErrFacilityNotFound) {
		return nil, fmt.Errorf("check code failed: %w", err)
	}

	newFac := &models.Facility{
		ID:             uuid.New(),
		Name:           name,
		Code:           code,
		FacilityTypeID: req.FacilityTypeID,
		AreaID:         req.AreaID,
		Address:        req.Address,
		Phone:          req.Phone,
		Email:          req.Email,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
	}

	created, err := s.FacilityRepository.Insert(newFac)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("facility name/code already exists")
		}
		return nil, fmt.Errorf("insert facility failed: %w", err)
	}
	return created, nil
}

func (s *FacilityService) UpdateFacility(idStr string, upd *models.FacilityCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Facility, error) {
	_ = ctx
	_ = userInfo

	existing, err := s.FacilityRepository.FindById(idStr, false)
	if err != nil {
		if errors.Is(err, repositories.ErrFacilityNotFound) {
			return nil, errors.New("facility not found")
		}
		return nil, fmt.Errorf("find facility failed: %w", err)
	}

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	if strings.TrimSpace(upd.Code) != "" {
		newCode := norm(upd.Code)
		if len(newCode) > 30 {
			return nil, errors.New("code exceeds max length")
		}
		if newCode != existing.Code {
			if ex, err := s.FacilityRepository.FindByCode(newCode); err == nil && ex.ID != existing.ID {
				return nil, errors.New("facility code already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrFacilityNotFound) {
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
			if ex, err := s.FacilityRepository.FindByName(newName); err == nil && ex.ID != existing.ID {
				return nil, errors.New("facility name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrFacilityNotFound) {
				return nil, fmt.Errorf("check name failed: %w", err)
			}
			existing.Name = newName
		}
	}

	if upd.FacilityTypeID != uuid.Nil {
		existing.FacilityTypeID = upd.FacilityTypeID
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
			return nil, errors.New("invalid email")
		} else {
			existing.Email = &e
		}
	}
	if upd.Latitude != nil {
		if *upd.Latitude < -90 || *upd.Latitude > 90 {
			return nil, errors.New("latitude out of range")
		}
		existing.Latitude = upd.Latitude
	}
	if upd.Longitude != nil {
		if *upd.Longitude < -180 || *upd.Longitude > 180 {
			return nil, errors.New("longitude out of range")
		}
		existing.Longitude = upd.Longitude
	}

	updated, err := s.FacilityRepository.Update(existing)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("facility name/code already exists")
		}
		return nil, fmt.Errorf("update facility failed: %w", err)
	}
	return updated, nil
}

func (service *FacilityService) DeleteFacilities(facilityRequest *models.FacilityIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, facilityId := range facilityRequest.IDs {
		_, err := service.FacilityRepository.FindById(facilityId.String(), false)
		if err != nil {
			if err == repositories.ErrFacilityNotFound {
				log.Printf("Facility not found: %v\n", facilityId)
				continue
			}
			log.Printf("Error finding facility %v: %v\n", facilityId, err)
			return errors.New("error finding facility")
		}

		if facilityRequest.IsHardDelete == "hardDelete" {
			if err := service.FacilityRepository.Delete(facilityId.String(), true); err != nil {
				log.Printf("Error hard deleting facility %v: %v\n", facilityId, err)
				return errors.New("error hard deleting facility")
			}
		} else {
			if err := service.FacilityRepository.Delete(facilityId.String(), false); err != nil {
				log.Printf("Error soft deleting facility %v: %v\n", facilityId, err)
				return errors.New("error soft deleting facility")
			}
		}
	}

	return nil
}

func (service *FacilityService) RestoreFacilities(facility *models.FacilityRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Facility, error) {
	var restoredFacilities []models.Facility

	for _, facilityId := range facility.IDs {
		facility := &models.Facility{ID: facilityId}

		restoredFacility, err := service.FacilityRepository.Restore(facility, facilityId.String())
		if err != nil {
			if err == repositories.ErrFacilityNotFound {
				log.Printf("Facility not found: %v\n", facilityId)
				continue
			}
			log.Printf("Error restoring facility %v: %v\n", facilityId, err)
			return nil, errors.New("error restoring facility")
		}

		restoredFacilities = append(restoredFacilities, *restoredFacility)
	}

	return restoredFacilities, nil
}