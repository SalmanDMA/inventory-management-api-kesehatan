package repositories

import (
	"errors"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type SalesAssignmentRepository interface {
	FindAll(tx *gorm.DB, salesPersonID uuid.UUID) ([]models.SalesAssignment, error)
	FindBySalesAndAreaID(tx *gorm.DB, salesPersonID uuid.UUID, areaID uuid.UUID) (*models.SalesAssignment, error)
	Insert(tx *gorm.DB, assignment *models.SalesAssignment) (*models.SalesAssignment, error)
	Update(tx *gorm.DB, assignment *models.SalesAssignment) (*models.SalesAssignment, error)
}

// ==============================
// Implementation
// ==============================

type SalesAssignmentRepositoryImpl struct {
	DB *gorm.DB
}

func NewSalesAssignmentRepository(db *gorm.DB) *SalesAssignmentRepositoryImpl {
	return &SalesAssignmentRepositoryImpl{DB: db}
}

func (r *SalesAssignmentRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *SalesAssignmentRepositoryImpl) FindAll(tx *gorm.DB, salesPersonID uuid.UUID) ([]models.SalesAssignment, error) {
	var assignments []models.SalesAssignment

	if err := r.useDB(tx).
		Preload("SalesPerson").
		Preload("Area").
		Where("sales_person_id = ?", salesPersonID).
		Find(&assignments).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_assignment")
	}
	return assignments, nil
}

func (r *SalesAssignmentRepositoryImpl) FindBySalesAndAreaID(tx *gorm.DB, salesPersonID uuid.UUID, areaID uuid.UUID) (*models.SalesAssignment, error) {
	var assignment models.SalesAssignment
	err := r.useDB(tx).
		Preload("SalesPerson").
		Preload("Area").
		Where("sales_person_id = ? AND area_id = ?", salesPersonID, areaID).
		First(&assignment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, HandleDatabaseError(err, "sales_assignment")
	}
	return &assignment, nil
}

// ---------- Mutations ----------

func (r *SalesAssignmentRepositoryImpl) Insert(tx *gorm.DB, assignment *models.SalesAssignment) (*models.SalesAssignment, error) {
	if err := r.useDB(tx).Create(assignment).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_assignment")
	}
	return assignment, nil
}

func (r *SalesAssignmentRepositoryImpl) Update(tx *gorm.DB, assignment *models.SalesAssignment) (*models.SalesAssignment, error) {
	if err := r.useDB(tx).Save(assignment).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_assignment")
	}
	return assignment, nil
}
