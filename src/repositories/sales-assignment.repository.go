package repositories

import (
	"errors"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesAssignmentRepository interface {
	FindAll(salesPersonID uuid.UUID) ([]models.SalesAssignment, error)
	FindBySalesAndAreaID(salesPersonID uuid.UUID, areaID uuid.UUID) (*models.SalesAssignment, error)
	Insert(roleModule *models.SalesAssignment) (*models.SalesAssignment, error)
	Update(roleModule *models.SalesAssignment) (*models.SalesAssignment, error)
}

type SalesAssignmentRepositoryImpl struct {
	DB *gorm.DB
}

func NewSalesAssignmentRepository(db *gorm.DB) *SalesAssignmentRepositoryImpl {
	return &SalesAssignmentRepositoryImpl{DB: db}
}

func (r *SalesAssignmentRepositoryImpl) FindAll(salesPersonID uuid.UUID) ([]models.SalesAssignment, error) {
	var salesPersons []models.SalesAssignment

	err := r.DB.
		Preload("SalesPerson").
		Preload("Area").
		Where("sales_person_id = ?", salesPersonID).
		Find(&salesPersons).Error

	if err != nil {
		return nil, HandleDatabaseError(err, "sales_assignment")
	}

	return salesPersons, nil
}

func (r *SalesAssignmentRepositoryImpl) FindBySalesAndAreaID(salesPersonID uuid.UUID, areaID uuid.UUID) (*models.SalesAssignment, error) {
	var roleModule models.SalesAssignment
	err := r.DB.
	Preload("SalesPerson").
	Preload("Area").
	Where("sales_person_id = ? AND area_id = ?", salesPersonID, areaID).
	First(&roleModule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, HandleDatabaseError(err, "sales_assignment")
	}
	return &roleModule, nil
}

func (r *SalesAssignmentRepositoryImpl) Insert(roleModule *models.SalesAssignment) (*models.SalesAssignment, error) {
	if err := r.DB.Create(&roleModule).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_assignment")
	}
	return roleModule, nil
}

func (r *SalesAssignmentRepositoryImpl) Update(roleModule *models.SalesAssignment) (*models.SalesAssignment, error) {
	if err := r.DB.Save(&roleModule).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_assignment")
	}
	return roleModule, nil
}