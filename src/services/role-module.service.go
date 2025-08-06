package services

import (
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RoleModuleService struct {
	RoleModuleRepository repositories.RoleModuleRepository
	ModuleRepository     repositories.ModuleRepository
}

func NewRoleModuleService(roleModuleRepository repositories.RoleModuleRepository, moduleRepository repositories.ModuleRepository) *RoleModuleService {
	return &RoleModuleService{
		RoleModuleRepository: roleModuleRepository, 
		ModuleRepository: moduleRepository,
	}
}

func (s *RoleModuleService) GetAllRoleModule(roleID uuid.UUID) ([]models.ResponseGetRoleModule, error) {
	roleModules, err := s.RoleModuleRepository.FindAll(roleID)

	if err != nil {
		return nil, err
	}

	roleModulesResponse := []models.ResponseGetRoleModule{}

	for _, roleModule := range roleModules {
		roleModulesResponse = append(roleModulesResponse, models.ResponseGetRoleModule{
			ID:        roleModule.ID,
			RoleID:    roleModule.RoleID,
			Role:      roleModule.Role,
			ModuleID:  roleModule.ModuleID,
			Module:    roleModule.Module,
			Checked:   roleModule.Checked,
		})
	}

	return roleModulesResponse, nil
}

func (s *RoleModuleService) CreateOrUpdateRoleModule(roleId uuid.UUID, roleModuleRequest *models.RoleModuleRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.RoleModule, error) {
	existingRoleModule, err := s.RoleModuleRepository.FindByRoleAndModule(roleId, roleModuleRequest.ModuleID)
	if err != nil {
		return nil, err
	}

	var result *models.RoleModule

	if existingRoleModule == nil {
		newRoleModule := &models.RoleModule{
			ID:       uuid.New(),
			RoleID:   &roleId,
			ModuleID: &roleModuleRequest.ModuleID,
			Checked:  roleModuleRequest.Checked,
		}
		result, err = s.RoleModuleRepository.Insert(newRoleModule)
		if err != nil {
			return nil, err
		}
	} else {
		existingRoleModule.Checked = roleModuleRequest.Checked
		result, err = s.RoleModuleRepository.Update(existingRoleModule)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}


