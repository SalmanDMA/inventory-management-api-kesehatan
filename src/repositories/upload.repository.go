package repositories

import (
	"fmt"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type UploadRepository interface {
	GetFilesToDelete(tx *gorm.DB, t time.Time) ([]models.Upload, error)
	FindAll(tx *gorm.DB) ([]models.Upload, error)
	FindById(tx *gorm.DB, uploadId string, includeTrashed bool) (*models.Upload, error)
	Insert(tx *gorm.DB, upload *models.Upload) (*models.Upload, error)
	Update(tx *gorm.DB, upload *models.Upload) (*models.Upload, error)
	Delete(tx *gorm.DB, uploadId string, isHardDelete bool) error
}

// ==============================
// Implementation
// ==============================

type UploadRepositoryImpl struct {
	DB *gorm.DB
}

func NewUploadRepository(db *gorm.DB) *UploadRepositoryImpl {
	return &UploadRepositoryImpl{DB: db}
}

func (r *UploadRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ---------- Reads ----------

func (r *UploadRepositoryImpl) GetFilesToDelete(tx *gorm.DB, t time.Time) ([]models.Upload, error) {
	var uploads []models.Upload
	if err := r.useDB(tx).
		Unscoped().
		Where("created_at < ? AND deleted_at IS NOT NULL", t).
		Find(&uploads).Error; err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}
	return uploads, nil
}

func (r *UploadRepositoryImpl) FindAll(tx *gorm.DB) ([]models.Upload, error) {
	var uploads []models.Upload
	if err := r.useDB(tx).Find(&uploads).Error; err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}
	return uploads, nil
}

func (r *UploadRepositoryImpl) FindById(tx *gorm.DB, uploadId string, includeTrashed bool) (*models.Upload, error) {
	var upload models.Upload
	db := r.useDB(tx)
	if includeTrashed {
		db = db.Unscoped()
	}

	if err := db.First(&upload, "id = ?", uploadId).Error; err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}
	return &upload, nil
}

// ---------- Mutations ----------

func (r *UploadRepositoryImpl) Insert(tx *gorm.DB, upload *models.Upload) (*models.Upload, error) {
	if upload.ID == uuid.Nil {
		return nil, fmt.Errorf("upload ID cannot be empty")
	}

	if err := r.useDB(tx).Create(upload).Error; err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}
	return upload, nil
}

func (r *UploadRepositoryImpl) Update(tx *gorm.DB, upload *models.Upload) (*models.Upload, error) {
	if upload.ID == uuid.Nil {
		return nil, fmt.Errorf("upload ID cannot be empty")
	}

	if err := r.useDB(tx).Save(upload).Error; err != nil {
		return nil, HandleDatabaseError(err, "upload")
	}
	return upload, nil
}

func (r *UploadRepositoryImpl) Delete(tx *gorm.DB, uploadId string, isHardDelete bool) error {
	db := r.useDB(tx)

	var upload models.Upload
	if err := db.Unscoped().First(&upload, "id = ?", uploadId).Error; err != nil {
		return HandleDatabaseError(err, "upload")
	}

	if isHardDelete {
		if err := db.Unscoped().Delete(&upload).Error; err != nil {
			return HandleDatabaseError(err, "upload")
		}
	} else {
		if err := db.Delete(&upload).Error; err != nil {
			return HandleDatabaseError(err, "upload")
		}
	}
	return nil
}
