package services

import (
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SalesAssignmentService struct {
	SalesAssignmentRepository repositories.SalesAssignmentRepository
	AreaRepository            repositories.AreaRepository
}

func NewSalesAssignmentService(
	saRepo repositories.SalesAssignmentRepository,
	areaRepo repositories.AreaRepository,
) *SalesAssignmentService {
	return &SalesAssignmentService{
		SalesAssignmentRepository: saRepo,
		AreaRepository:            areaRepo,
	}
}

func (s *SalesAssignmentService) GetAllSalesAssignment(salesPersonID uuid.UUID) ([]models.ResponseGetSalesAssignment, error) {
	rows, err := s.SalesAssignmentRepository.FindAll(nil, salesPersonID)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResponseGetSalesAssignment, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.ResponseGetSalesAssignment{
			ID:            r.ID,
			SalesPersonID: r.SalesPersonID,
			AreaID:        r.AreaID,
			Area:          r.Area,
			SalesPerson:   r.SalesPerson,
			Checked:       r.Checked,
		})
	}
	return out, nil
}

func (s *SalesAssignmentService) CreateOrUpdateSalesAssignment(
	salesPersonID uuid.UUID,
	req *models.SalesAssignmentRequest,
	ctx *fiber.Ctx,
	userInfo *models.User,
) (*models.SalesAssignment, error) {
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

	// optional: validasi area exist
	if _, err := s.AreaRepository.FindById(tx, req.AreaID.String(), false); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("area not found: %w", err)
	}

	existing, err := s.SalesAssignmentRepository.FindBySalesAndAreaID(tx, salesPersonID, req.AreaID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	var result *models.SalesAssignment
	if existing == nil {
		row := &models.SalesAssignment{
			ID:            uuid.New(),
			SalesPersonID: salesPersonID,
			AreaID:        req.AreaID,
			Checked:       req.Checked,
		}
		result, err = s.SalesAssignmentRepository.Insert(tx, row)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		existing.Checked = req.Checked
		result, err = s.SalesAssignmentRepository.Update(tx, existing)
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
