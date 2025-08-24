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

type RoleService struct {
	RoleRepository repositories.RoleRepository
}

func NewRoleService(roleRepo repositories.RoleRepository) *RoleService {
	return &RoleService{RoleRepository: roleRepo}
}

func (service *RoleService) GetAllRoles(userInfo *models.User) ([]models.ResponseGetRole, error) {
	roles, err := service.RoleRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResponseGetRole, 0, len(roles))
	for _, r := range roles {
		out = append(out, models.ResponseGetRole{
			ID:          r.ID,
			Name:        r.Name,
			Alias:       r.Alias,
			Color:       r.Color,
			Description: r.Description,
			RoleModules: r.RoleModules,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
			DeletedAt:   r.DeletedAt,
		})
	}
	return out, nil
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

	roles, totalCount, err := service.RoleRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetRole, 0, len(roles))
	for _, r := range roles {
		if strings.ToLower(userInfo.Role.Name) != "developer" && strings.ToLower(r.Name) == "developer" {
			continue
		}
		resp = append(resp, models.ResponseGetRole{
			ID:          r.ID,
			Name:        r.Name,
			Alias:       r.Alias,
			Color:       r.Color,
			Description: r.Description,
			RoleModules: r.RoleModules,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
			DeletedAt:   r.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	return &models.RolePaginatedResponse{
		Data: resp,
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

func (service *RoleService) GetRoleByID(roleId string) (*models.ResponseGetRole, error) {
	role, err := service.RoleRepository.FindById(nil, roleId, false)
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
	name := norm(req.Name)
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

	// cek duplikasi (read boleh tanpa tx)
	if _, err := s.RoleRepository.FindByName(nil, name); err == nil {
		return nil, errors.New("role name already exists")
	} else if !errors.Is(err, repositories.ErrRoleNotFound) {
		return nil, fmt.Errorf("check name failed: %w", err)
	}
	if alias != "" {
		if _, err := s.RoleRepository.FindByAlias(nil, alias); err == nil {
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

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	created, err := s.RoleRepository.Insert(tx, role)
	if err != nil {
		_ = tx.Rollback()
		if repositories.IsUniqueViolation(err) {
			return nil, errors.New("role name/alias already exists")
		}
		return nil, fmt.Errorf("insert role failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return created, nil
}

func (s *RoleService) UpdateRole(roleID string, upd *models.RoleUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Role, error) {
	_ = ctx
	_ = userInfo

	existing, err := s.RoleRepository.FindById(nil, roleID, false)
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
			if ex, err := s.RoleRepository.FindByName(nil, newName); err == nil && ex.ID != existing.ID {
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
			if ex, err := s.RoleRepository.FindByAlias(nil, newAlias); err == nil && ex.ID != existing.ID {
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

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	updated, err := s.RoleRepository.Update(tx, existing)
	if err != nil {
		_ = tx.Rollback()
		if repositories.IsUniqueViolation(err) {
		 return nil, errors.New("role name/alias already exists")
		}
		return nil, fmt.Errorf("update role failed: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return updated, nil
}

func (service *RoleService) DeleteRoles(roleRequest *models.RoleIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx
	_ = userInfo

	for _, roleId := range roleRequest.IDs {
		// pastikan ada
		if _, err := service.RoleRepository.FindById(nil, roleId.String(), true); err != nil {
			if err == repositories.ErrRoleNotFound {
				log.Printf("Role not found: %v\n", roleId)
				continue
			}
			log.Printf("Error finding role %v: %v\n", roleId, err)
			return errors.New("error finding role")
		}

		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for role %v: %v\n", roleId, tx.Error)
			return errors.New("error beginning transaction")
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					_ = tx.Rollback()
				}
			}()
			if err := service.RoleRepository.Delete(tx, roleId.String(), roleRequest.IsHardDelete == "hardDelete"); err != nil {
				_ = tx.Rollback()
				log.Printf("Error deleting role %v: %v\n", roleId, err)
				panic("rollback")
			}
			if err := tx.Commit().Error; err != nil {
				log.Printf("Error committing delete for role %v: %v\n", roleId, err)
				panic("rollback")
			}
		}()
	}
	return nil
}

func (service *RoleService) RestoreRoles(req *models.RoleRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Role, error) {
	_ = ctx
	_ = userInfo

	var restored []models.Role
	for _, id := range req.IDs {
		if _, err := service.RoleRepository.FindById(nil, id.String(), true); err != nil {
			if err == repositories.ErrRoleNotFound {
				log.Printf("Role not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error finding role %v: %v\n", id, err)
			return nil, errors.New("error finding role")
		}

		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for role restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}
		var r *models.Role
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					_ = tx.Rollback()
				}
			}()
			role := &models.Role{ID: id}
			var err error
			r, err = service.RoleRepository.Restore(tx, role.ID.String())
			if err != nil {
				_ = tx.Rollback()
				log.Printf("Error restoring role %v: %v\n", id, err)
				panic("rollback")
			}
			if err := tx.Commit().Error; err != nil {
				log.Printf("Error committing role restore %v: %v\n", id, err)
				panic("rollback")
			}
		}()
		if r != nil {
			restored = append(restored, *r)
		}
	}
	return restored, nil
}
