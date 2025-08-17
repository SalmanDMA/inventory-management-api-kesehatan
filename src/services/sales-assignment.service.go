package services

import (
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SalesAssignmentService struct {
	SalesAssignmentRepository repositories.SalesAssignmentRepository
	AreaRepository     repositories.AreaRepository
}

func NewSalesAssignmentService(salesAssignmentRepository repositories.SalesAssignmentRepository, areaRepository repositories.AreaRepository) *SalesAssignmentService {
	return &SalesAssignmentService{
		SalesAssignmentRepository: salesAssignmentRepository, 
		AreaRepository: areaRepository,
	}
}

func (s *SalesAssignmentService) GetAllSalesAssignment(salesPersonID uuid.UUID) ([]models.ResponseGetSalesAssignment, error) {
	salesAssignments, err := s.SalesAssignmentRepository.FindAll(salesPersonID)

	if err != nil {
		return nil, err
	}

	salesAssignmentsResponse := []models.ResponseGetSalesAssignment{}

	for _, salesAssignment := range salesAssignments {
		salesAssignmentsResponse = append(salesAssignmentsResponse, models.ResponseGetSalesAssignment{
			ID:        salesAssignment.ID,
			SalesPersonID: salesAssignment.SalesPersonID,
		 AreaID:        salesAssignment.AreaID,
			Area:          salesAssignment.Area,
			SalesPerson:   salesAssignment.SalesPerson,
			Checked:   salesAssignment.Checked,
		})
	}

	return salesAssignmentsResponse, nil
}

func (s *SalesAssignmentService) CreateOrUpdateSalesAssignment(salesPersonID uuid.UUID, salesAssignmentRequest *models.SalesAssignmentRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.SalesAssignment, error) {
	existingSalesAssignment, err := s.SalesAssignmentRepository.FindBySalesAndAreaID(salesPersonID, salesAssignmentRequest.AreaID)
	if err != nil {
		return nil, err
	}

	var result *models.SalesAssignment

	if existingSalesAssignment == nil {
		newSalesAssignment := &models.SalesAssignment{
			ID:       uuid.New(),
			SalesPersonID:   salesPersonID,
			AreaID:   salesAssignmentRequest.AreaID,
			Checked:  salesAssignmentRequest.Checked,
		}
		result, err = s.SalesAssignmentRepository.Insert(newSalesAssignment)
		if err != nil {
			return nil, err
		}
	} else {
		existingSalesAssignment.Checked = salesAssignmentRequest.Checked
		result, err = s.SalesAssignmentRepository.Update(existingSalesAssignment)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}


