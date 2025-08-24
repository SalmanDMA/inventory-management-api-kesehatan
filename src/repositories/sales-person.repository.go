package repositories

import (
	"errors"
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type SalesPersonRepository interface {
	FindAll(tx *gorm.DB) ([]models.SalesPerson, error)
	FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.SalesPerson, int64, error)
	FindByEmail(tx *gorm.DB, email string) (*models.SalesPerson, error)
	FindById(tx *gorm.DB, salesPersonId string, includeTrashed bool) (*models.SalesPerson, error)
	Insert(tx *gorm.DB, salesPerson *models.SalesPerson) (*models.SalesPerson, error)
	Update(tx *gorm.DB, salesPerson *models.SalesPerson) (*models.SalesPerson, error)
	Delete(tx *gorm.DB, salesPersonId string, isHardDelete bool) error
	Restore(tx *gorm.DB, salesPersonID string) (*models.SalesPerson, error)
}

// ==============================
// Implementation
// ==============================

type SalesPersonRepositoryImpl struct {
	DB *gorm.DB
}

func NewSalesPersonRepository(db *gorm.DB) *SalesPersonRepositoryImpl {
	return &SalesPersonRepositoryImpl{DB: db}
}

func (r *SalesPersonRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *SalesPersonRepositoryImpl) FindAll(tx *gorm.DB) ([]models.SalesPerson, error) {
	var salesPersons []models.SalesPerson
	if err := r.useDB(tx).
		Unscoped().
		Preload("Assignments").
		Find(&salesPersons).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}
	return salesPersons, nil
}

func (r *SalesPersonRepositoryImpl) FindAllPaginated(tx *gorm.DB, req *models.PaginationRequest) ([]models.SalesPerson, int64, error) {
	var (
		salesPersons []models.SalesPerson
		totalCount   int64
	)

	db := r.useDB(tx)

	spStmt := &gorm.Statement{DB: db}
	if err := spStmt.Parse(&models.SalesPerson{}); err != nil {
		return nil, 0, err
	}
	spTable := spStmt.Schema.Table

	saStmt := &gorm.Statement{DB: db}
	if err := saStmt.Parse(&models.SalesAssignment{}); err != nil {
		return nil, 0, err
	}
	saTable := saStmt.Schema.Table

	areaStmt := &gorm.Statement{DB: db}
	if err := areaStmt.Parse(&models.Area{}); err != nil {
		return nil, 0, err
	}
	areaTable := areaStmt.Schema.Table

	query := db.
		Unscoped().
		Model(&models.SalesPerson{}).
		Preload("Assignments", "deleted_at IS NULL").
		Preload("Assignments.Area")

	switch strings.ToLower(req.Status) {
	case "active":
		query = query.Where(fmt.Sprintf("%s.deleted_at IS NULL", spTable))
	case "deleted":
		query = query.Where(fmt.Sprintf("%s.deleted_at IS NOT NULL", spTable))
	case "all":
		// no filter
	default:
		query = query.Where(fmt.Sprintf("%s.deleted_at IS NULL", spTable))
	}

	needDistinct := false

	if req.AreaID != "" {
		if areaUUID, err := uuid.Parse(req.AreaID); err == nil {
			needDistinct = true
			query = query.
				Joins("LEFT JOIN " + saTable + " sa_area ON sa_area.salesperson_id = " + spTable + ".id AND sa_area.deleted_at IS NULL").
				Where("sa_area.area_id = ?", areaUUID)
		}
	}

	if s := strings.TrimSpace(req.Search); s != "" {
		needDistinct = true
		p := "%" + strings.ToLower(s) + "%"

		query = query.
			Joins("LEFT JOIN " + saTable + " sa_search ON sa_search.salesperson_id = " + spTable + ".id AND sa_search.deleted_at IS NULL").
			Joins("LEFT JOIN " + areaTable + " a ON a.id = sa_search.area_id").
			Where(`
				LOWER(`+spTable+`.name) LIKE ? OR
				LOWER(COALESCE(`+spTable+`.email, '')) LIKE ? OR
				LOWER(COALESCE(`+spTable+`.phone, '')) LIKE ? OR
				LOWER(COALESCE(a.name, '')) LIKE ?
			`, p, p, p, p)
	}

	countQ := query
	if needDistinct {
		countQ = countQ.Distinct(spTable + ".id")
	}
	if err := countQ.Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "sales_person")
	}

	offset := (req.Page - 1) * req.Limit
	listQ := query.
		Offset(offset).
		Limit(req.Limit).
		Order(spTable + ".created_at DESC")

	if needDistinct {
		listQ = listQ.Select("DISTINCT " + spTable + ".*")
	}

	if err := listQ.Find(&salesPersons).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "sales_person")
	}

	return salesPersons, totalCount, nil
}

func (r *SalesPersonRepositoryImpl) FindByEmail(tx *gorm.DB, email string) (*models.SalesPerson, error) {
	var sp models.SalesPerson
	err := r.useDB(tx).
		Preload("Assignments").
		Where("email = ?", email).
		First(&sp).Error

	if err == nil {
		return &sp, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSalesPersonNotFound
	}
	return nil, HandleDatabaseError(err, "sales_person")
}

func (r *SalesPersonRepositoryImpl) FindById(tx *gorm.DB, salesPersonId string, includeTrashed bool) (*models.SalesPerson, error) {
	var sp models.SalesPerson
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.
		Preload("Assignments").
		First(&sp, "id = ?", salesPersonId).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}
	return &sp, nil
}

// ---------- Mutations ----------

func (r *SalesPersonRepositoryImpl) Insert(tx *gorm.DB, salesPerson *models.SalesPerson) (*models.SalesPerson, error) {
	if salesPerson.ID == uuid.Nil {
		return nil, fmt.Errorf("sales person ID cannot be empty")
	}
	if err := r.useDB(tx).Create(salesPerson).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}
	return salesPerson, nil
}

func (r *SalesPersonRepositoryImpl) Update(tx *gorm.DB, salesPerson *models.SalesPerson) (*models.SalesPerson, error) {
	if salesPerson.ID == uuid.Nil {
		return nil, fmt.Errorf("sales person ID cannot be empty")
	}
	if err := r.useDB(tx).Save(salesPerson).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}
	return salesPerson, nil
}

func (r *SalesPersonRepositoryImpl) Delete(tx *gorm.DB, salesPersonId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var sp models.SalesPerson
	if err := db.Unscoped().First(&sp, "id = ?", salesPersonId).Error; err != nil {
		return HandleDatabaseError(err, "sales_person")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&sp).Error; err != nil {
			return HandleDatabaseError(err, "sales_person")
		}
	} else {
		if err := db.Delete(&sp).Error; err != nil {
			return HandleDatabaseError(err, "sales_person")
		}
	}
	return nil
}

func (r *SalesPersonRepositoryImpl) Restore(tx *gorm.DB, salesPersonID string) (*models.SalesPerson, error) {
	db := r.useDB(tx)

	if err := db.Unscoped().
		Model(&models.SalesPerson{}).
		Where("id = ?", salesPersonID).
		Update("deleted_at", nil).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}

	var restored models.SalesPerson
	if err := db.
		Preload("Assignments").
		First(&restored, "id = ?", salesPersonID).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}
	return &restored, nil
}
