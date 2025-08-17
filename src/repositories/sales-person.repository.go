package repositories

import (
	"errors"
	"fmt"
	"strings"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesPersonRepository interface {
	FindAll() ([]models.SalesPerson, error)
	FindAllPaginated(req *models.PaginationRequest) ([]models.SalesPerson, int64, error)
	FindByEmail(email string) (*models.SalesPerson, error)
	FindById(salesPersonId string, isSoftDelete bool) (*models.SalesPerson, error)
	Insert(salesPerson *models.SalesPerson) (*models.SalesPerson, error)
	Update(salesPerson *models.SalesPerson) (*models.SalesPerson, error)
	Delete(salesPersonId string, isHardDelete bool) error
	Restore(salesPerson *models.SalesPerson, salesPersonId string) (*models.SalesPerson, error)
}

type SalesPersonRepositoryImpl struct {
	DB *gorm.DB
}

func NewSalesPersonRepository(db *gorm.DB) *SalesPersonRepositoryImpl {
	return &SalesPersonRepositoryImpl{DB: db}
}

func (r *SalesPersonRepositoryImpl) FindAll() ([]models.SalesPerson, error) {
	var salesPersons []models.SalesPerson
	if err := r.DB.Unscoped().
	 Preload("Assignments").
		Find(&salesPersons).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}
	return salesPersons, nil
}

func (r *SalesPersonRepositoryImpl) FindAllPaginated(req *models.PaginationRequest) ([]models.SalesPerson, int64, error) {
	var (
		salesPersons []models.SalesPerson
		totalCount   int64
	)

	spStmt := &gorm.Statement{DB: r.DB}
	if err := spStmt.Parse(&models.SalesPerson{}); err != nil {
		return nil, 0, err
	}
	spTable := spStmt.Schema.Table

	saStmt := &gorm.Statement{DB: r.DB}
	if err := saStmt.Parse(&models.SalesAssignment{}); err != nil {
		return nil, 0, err
	}
	saTable := saStmt.Schema.Table

	areaStmt := &gorm.Statement{DB: r.DB}
	if err := areaStmt.Parse(&models.Area{}); err != nil {
		return nil, 0, err
	}
	areaTable := areaStmt.Schema.Table 

	query := r.DB.
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

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
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

func (r *SalesPersonRepositoryImpl) FindByEmail(email string) (*models.SalesPerson, error) {
    var sp models.SalesPerson
    err := r.DB.
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

func (r *SalesPersonRepositoryImpl) FindById(salesPersonId string, isSoftDelete bool) (*models.SalesPerson, error) {
	var salesPerson *models.SalesPerson
	db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	err := db.
	 Preload("Assignments").
		First(&salesPerson, "id = ?", salesPersonId).Error

	if err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}

	return salesPerson, nil
}

func (r *SalesPersonRepositoryImpl) Insert(salesPerson *models.SalesPerson) (*models.SalesPerson, error) {

	if salesPerson.ID == uuid.Nil {
		return nil, fmt.Errorf("sales person ID cannot be empty")
	}

	if err := r.DB.Create(&salesPerson).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}
	return salesPerson, nil
}

func (r *SalesPersonRepositoryImpl) Update(salesPerson *models.SalesPerson) (*models.SalesPerson, error) {

	if salesPerson.ID == uuid.Nil {
		return nil, fmt.Errorf("sales person ID cannot be empty")
	}

	if err := r.DB.Save(&salesPerson).Error; err != nil {
		return nil, HandleDatabaseError(err, "sales_person")
	}
	return salesPerson, nil
}

func (r *SalesPersonRepositoryImpl) Delete(salesPersonId string, isHardDelete bool) error {
	var salesPerson *models.SalesPerson

	if err := r.DB.Unscoped().First(&salesPerson, "id = ?", salesPersonId).Error; err != nil {
		return HandleDatabaseError(err, "sales_person")
	}

	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&salesPerson).Error; err != nil {
			return HandleDatabaseError(err, "sales_person")
		}
	} else {
		if err := r.DB.Delete(&salesPerson).Error; err != nil {
			return HandleDatabaseError(err, "sales_person")
		}
	}
	return nil
}

func (r *SalesPersonRepositoryImpl) Restore(salesPerson *models.SalesPerson, salesPersonID string) (*models.SalesPerson, error) {
	if err := r.DB.Unscoped().Model(salesPerson).Where("id = ?", salesPersonID).Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}

	var restoredSalesPerson *models.SalesPerson
	if err := r.DB.Unscoped().First(&restoredSalesPerson, "id = ?", salesPersonID).Error; err != nil {
		return nil, err
	}
	
	return restoredSalesPerson, nil
}


