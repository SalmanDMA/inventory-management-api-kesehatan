package services

import (
	"errors"
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

func (service *RoleService) CreateRole(roleRequest *models.RoleCreateRequest, ctx *fiber.Ctx, userInfo *models.User) ( *models.Role, error) {
	if _, err := service.RoleRepository.FindByName(roleRequest.Name); err == nil {
		return nil, errors.New("role already exists") 
	} else if err != repositories.ErrRoleNotFound {
		 return nil, errors.New("error checking role: " + err.Error())
	}

	newRole := &models.Role{
		ID: 									uuid.New(),
		Name:        roleRequest.Name,
		Alias:       roleRequest.Alias,
		Color:       roleRequest.Color,
		Description: roleRequest.Description,
	}

	role, err := service.RoleRepository.Insert(newRole)

	if err != nil {
		return nil, err
	}

	return role, nil
}

func (service *RoleService) UpdateRole(roleID string, roleUpdate *models.RoleUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Role, error) {
	roleExists, err := service.RoleRepository.FindById(roleID, false)
	if err != nil {
		return nil, err
	}
	if roleExists == nil {
		return nil, errors.New("role not found")
	}

	if roleUpdate.Name != "" {
		roleExists.Name = roleUpdate.Name
	}

	if roleUpdate.Alias != "" {
		roleExists.Alias = roleUpdate.Alias
	}

	if roleUpdate.Color != "" {
		roleExists.Color = roleUpdate.Color
	}

	if roleUpdate.Description != "" {
		roleExists.Description = roleUpdate.Description
	}

	updateRole , err := service.RoleRepository.Update(roleExists)
	if err != nil {
		return nil, err
	}

	return updateRole, nil
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