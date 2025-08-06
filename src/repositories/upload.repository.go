package repositories

import (
	"fmt"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadRepository interface {
	GetFilesToDelete(time time.Time) ([]models.Upload, error)
	FindAll() ([]models.Upload, error)
	FindById(uploadId string, isSoftDelete bool) (*models.Upload, error)
	Insert(upload *models.Upload) (*models.Upload, error)
	Update(upload *models.Upload) (*models.Upload, error)
	Delete(uploadId string, isHardDelete bool) error
}

type UploadRepositoryImpl struct{
	DB *gorm.DB
}

func NewUploadRepository(db *gorm.DB) *UploadRepositoryImpl {
	return &UploadRepositoryImpl{DB: db}
}

func (r *UploadRepositoryImpl) GetFilesToDelete(time time.Time) ([]models.Upload, error) {
	var uploads []models.Upload
	if err := r.DB.
	 Unscoped().
		Where("created_at < ? AND deleted_at IS NOT NULL", time).
		Find(&uploads).Error; err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}
	return uploads, nil
}

func (r *UploadRepositoryImpl) FindAll() ([]models.Upload, error) {
	var uploads []models.Upload
	if err := r.DB.Find(&uploads).Error; err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}
	return uploads, nil
}

func (r *UploadRepositoryImpl) FindById(uploadId string, isSoftDelete bool) (*models.Upload, error) {
	var upload models.Upload
		db := r.DB

	if !isSoftDelete {
		db = db.Unscoped()
	}

	err := db.
		First(&upload, "id = ?", uploadId).Error

	if err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}

	return &upload, nil
}

func (r *UploadRepositoryImpl) Insert(upload *models.Upload) (*models.Upload, error) {

	if upload.ID == uuid.Nil {
		return nil, fmt.Errorf("upload ID cannot be empty")
	}

	if err := r.DB.Create(&upload).Error; err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}
	return upload, nil
}

func (r *UploadRepositoryImpl) Update(upload *models.Upload) (*models.Upload, error) {

	if upload.ID == uuid.Nil {
		return nil, fmt.Errorf("upload ID cannot be empty")
	}

	if err := r.DB.Save(&upload).Error; err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}
	return upload, nil
}

func (r *UploadRepositoryImpl) Delete(uploadId string, isHardDelete bool) error {
	var upload *models.Upload
	
	if err := r.DB.Unscoped().First(&upload, "id = ?", uploadId).Error; err != nil {
		return HandleDatabaseError(err, "upload")
	}

	if isHardDelete {
		if err := r.DB.Unscoped().Delete(&upload).Error; err != nil {
			return HandleDatabaseError(err, "upload")
		}
	} else {
		if err := r.DB.Delete(&upload).Error; err != nil {
			return HandleDatabaseError(err, "upload")
		}
	}
	return nil
}
