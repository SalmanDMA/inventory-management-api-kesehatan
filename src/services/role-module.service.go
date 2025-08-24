package services

import (
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RoleModuleService struct {
	RoleModuleRepository repositories.RoleModuleRepository
	ModuleRepository     repositories.ModuleRepository
}

func NewRoleModuleService(
	roleModuleRepository repositories.RoleModuleRepository,
	moduleRepository repositories.ModuleRepository,
) *RoleModuleService {
	return &RoleModuleService{
		RoleModuleRepository: roleModuleRepository,
		ModuleRepository:     moduleRepository,
	}
}

func (s *RoleModuleService) GetAllRoleModule(roleID uuid.UUID) ([]models.ResponseGetRoleModule, error) {
	// read-only, tx boleh nil
	roleModules, err := s.RoleModuleRepository.FindAll(nil, roleID)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetRoleModule, 0, len(roleModules))
	for _, rm := range roleModules {
		resp = append(resp, models.ResponseGetRoleModule{
			ID:       rm.ID,
			RoleID:   rm.RoleID,
			Role:     rm.Role,
			ModuleID: rm.ModuleID,
			Module:   rm.Module,
			Checked:  rm.Checked,
		})
	}
	return resp, nil
}

func (s *RoleModuleService) CreateOrUpdateRoleModule(
	roleID uuid.UUID,
	req *models.RoleModuleRequest,
	ctx *fiber.Ctx,
	userInfo *models.User,
) (*models.RoleModule, error) {
	_ = ctx
	_ = userInfo

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// cek existing mapping dalam 1 tx
	existing, err := s.RoleModuleRepository.FindByRoleAndModule(tx, roleID, req.ModuleID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	var result *models.RoleModule
	if existing == nil {
		newRM := &models.RoleModule{
			ID:       uuid.New(),
			RoleID:   &roleID,
			ModuleID: &req.ModuleID,
			Checked:  req.Checked,
		}
		result, err = s.RoleModuleRepository.Insert(tx, newRM)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		existing.Checked = req.Checked
		result, err = s.RoleModuleRepository.Update(tx, existing)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return result, nil
}
