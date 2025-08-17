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

type RoleService struct {
	RoleRepository repositories.RoleRepository
}

func NewRoleService(roleRepo repositories.RoleRepository) *RoleService {
	return &RoleService{
		RoleRepository: roleRepo,
	}
}

func (service *RoleService) GetAllRoles(userInfo *models.User) ([]models.ResponseGetRole, error) {
	roles, err := service.RoleRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var rolesResponse []models.ResponseGetRole
	for _, role := range roles {
		rolesResponse = append(rolesResponse, models.ResponseGetRole{
			ID:          role.ID,
			Name:        role.Name,
			Alias:       role.Alias,
			Color:       role.Color,
			Description: role.Description,
			RoleModules: role.RoleModules,
			CreatedAt: 	role.CreatedAt,
			UpdatedAt: 	role.UpdatedAt,
			DeletedAt: 	role.DeletedAt,
		})
	}

	return rolesResponse, nil
}

func (service *RoleService) GetAllRolesPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.RolePaginatedResponse, error) {
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

	roles, totalCount, err := service.RoleRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	rolesResponse := []models.ResponseGetRole{}
	for _, role := range roles {
		if strings.ToLower(userInfo.Role.Name) != "developer" && strings.ToLower(role.Name) == "developer" {
			continue
		}
		
		rolesResponse = append(rolesResponse, models.ResponseGetRole{
		ID:          role.ID,
			Name:        role.Name,
			Alias:       role.Alias,
			Color:       role.Color,
			Description: role.Description,
			RoleModules: role.RoleModules,
			CreatedAt: 	role.CreatedAt,
			UpdatedAt: 	role.UpdatedAt,
			DeletedAt: 	role.DeletedAt,
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

	return &models.RolePaginatedResponse{
		Data:       rolesResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *RoleService) GetRoleByID(roleId string) (*models.ResponseGetRole, error) {
	role, err := service.RoleRepository.FindById(roleId, false)

	if err != nil {
		return nil, err
	}

	return &models.ResponseGetRole{
		 ID:          role.ID,
			Name:        role.Name,
			Alias:       role.Alias,
			Color:       role.Color,
			Description: role.Description,
			RoleModules: role.RoleModules,
	}, nil
}

func (s *RoleService) CreateRole(req *models.RoleCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Role, error) {
	_ = ctx
	_ = userInfo

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	name  := norm(req.Name)
	alias := norm(req.Alias)
	color := norm(req.Color)

	if name == "" {
		return nil, errors.New("name is required")
	}
	if len(name) > 100 {
		return nil, errors.New("name exceeds max length")
	}
	if alias != "" && len(alias) > 50 {
		return nil, errors.New("alias exceeds max length")
	}
	if color != "" && !repositories.IsHexColor(color) {
		return nil, errors.New("color must be a valid hex (e.g. #1a2b3c)")
	}

	if _, err := s.RoleRepository.FindByName(name); err == nil {
		return nil, errors.New("role name already exists")
	} else if !errors.Is(err, repositories.ErrRoleNotFound) {
		return nil, fmt.Errorf("check name failed: %w", err)
	}
	if alias != "" {
		if _, err := s.RoleRepository.FindByAlias(alias); err == nil {
			return nil, errors.New("role alias already exists")
		} else if !errors.Is(err, repositories.ErrRoleNotFound) {
			return nil, fmt.Errorf("check alias failed: %w", err)
		}
	}

	role := &models.Role{
		ID:          uuid.New(),
		Name:        name,   
		Alias:       alias,  
		Color:       color,  
		Description: req.Description,
	}

	created, err := s.RoleRepository.Insert(role)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("role name/alias already exists")
		}
		return nil, fmt.Errorf("insert role failed: %w", err)
	}
	return created, nil
}

func (s *RoleService) UpdateRole(roleID string, upd *models.RoleUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Role, error) {
	_ = ctx
	_ = userInfo

	existing, err := s.RoleRepository.FindById(roleID, false)
	if err != nil {
		if errors.Is(err, repositories.ErrRoleNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("find role failed: %w", err)
	}

	norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

	if strings.TrimSpace(upd.Name) != "" {
		newName := norm(upd.Name)
		if len(newName) > 100 {
			return nil, errors.New("name exceeds max length")
		}
		if newName != existing.Name {
			if ex, err := s.RoleRepository.FindByName(newName); err == nil && ex.ID != existing.ID {
				return nil, errors.New("role name already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrRoleNotFound) {
				return nil, fmt.Errorf("check name failed: %w", err)
			}
			existing.Name = newName
		}
	}

	if strings.TrimSpace(upd.Alias) != "" {
		newAlias := norm(upd.Alias)
		if len(newAlias) > 50 {
			return nil, errors.New("alias exceeds max length")
		}
		if newAlias != existing.Alias {
			if ex, err := s.RoleRepository.FindByAlias(newAlias); err == nil && ex.ID != existing.ID {
				return nil, errors.New("role alias already exists")
			} else if err != nil && !errors.Is(err, repositories.ErrRoleNotFound) {
				return nil, fmt.Errorf("check alias failed: %w", err)
			}
			existing.Alias = newAlias
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

	updated, err := s.RoleRepository.Update(existing)
	if err != nil {
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("role name/alias already exists")
		}
		return nil, fmt.Errorf("update role failed: %w", err)
	}
	return updated, nil
}

func (service *RoleService) DeleteRoles(roleRequest *models.RoleIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, roleId := range roleRequest.IDs {
		_, err := service.RoleRepository.FindById(roleId.String(), false)
		if err != nil {
			if err == repositories.ErrRoleNotFound {
				log.Printf("Role not found: %v\n", roleId)
				continue
			}
			log.Printf("Error finding role %v: %v\n", roleId, err)
			return errors.New("error finding useroler")
		}

		if roleRequest.IsHardDelete == "hardDelete" {
			if err := service.RoleRepository.Delete(roleId.String(), true); err != nil {
				log.Printf("Error hard deleting role %v: %v\n", roleId, err)
				return errors.New("error hard deleting role")
			}
		} else {
			if err := service.RoleRepository.Delete(roleId.String(), false); err != nil {
				log.Printf("Error soft deleting role %v: %v\n", roleId, err)
				return errors.New("error soft deleting role")
			}
		}
	}

	return nil
}

func (service *RoleService) RestoreRoles(role *models.RoleRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Role, error) {
	var restoredRoles []models.Role

	for _, roleId := range role.IDs {
		role := &models.Role{ID: roleId}

		restoredRole, err := service.RoleRepository.Restore(role, roleId.String())
		if err != nil {
			if err == repositories.ErrRoleNotFound {
				log.Printf("Role not found: %v\n", roleId)
				continue
			}
			log.Printf("Error restoring role %v: %v\n", roleId, err)
			return nil, errors.New("error restoring role")
		}

		restoredRoles = append(restoredRoles, *restoredRole)
	}

	return restoredRoles, nil
}