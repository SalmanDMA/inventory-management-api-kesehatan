package services

import (
	"errors"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/google/uuid"
)

type UploadService struct {
	UploadRepository repositories.UploadRepository
}

func NewUploadService(uploadRepo repositories.UploadRepository) *UploadService {
	return &UploadService{
		UploadRepository: uploadRepo,
	}
}

func (s *UploadService) GetAllUploads() ([]models.ResponseGetUpload, error) {
	uploads, err := s.UploadRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResponseGetUpload, 0, len(uploads))
	for _, u := range uploads {
		out = append(out, models.ResponseGetUpload{
			ID:             u.ID,
			Filename:       u.Filename,
			FilenameOrigin: u.FilenameOrigin,
			Category:       u.Category,
			Path:           u.Path,
			Type:           u.Type,
			Mime:           u.Mime,
			Extension:      u.Extension,
			Size:           u.Size,
			CreatedAt:      u.CreatedAt,
			UpdatedAt:      u.UpdatedAt,
			DeletedAt:      u.DeletedAt,
		})
	}
	return out, nil
}

func (s *UploadService) CreateUpload(in *models.UploadCreateRequest) (*models.Upload, error) {
	newUpload := &models.Upload{
		ID:             uuid.New(),
		Filename:       in.Filename,
		FilenameOrigin: in.FilenameOrigin,
		Category:       in.Category,
		Path:           in.Path,
		Type:           in.Type,
		Mime:           in.Mime,
		Extension:      in.Extension,
		Size:           in.Size,
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	created, err := s.UploadRepository.Insert(tx, newUpload)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// ambil ulang (kalau repo melakukan preload dsb.)
	out, err := s.UploadRepository.FindById(nil, created.ID.String(), true)
	if err != nil {
		// kalau gagal fetch, setidaknya kembalikan hasil create
		return created, nil
	}
	return out, nil
}

func (s *UploadService) UpdateUpload(uploadID string, in *models.UploadUpdateRequest) (*models.Upload, error) {
	// find dulu (di luar tx boleh, atau di dalam juga boleh; di sini pakai luar untuk kesederhanaan)
	exists, err := s.UploadRepository.FindById(nil, uploadID, true)
	if err != nil {
		return nil, err
	}
	if exists == nil {
		return nil, errors.New("upload not found")
	}

	// map perubahan
	if in.Filename != "" {
		exists.Filename = in.Filename
	}
	if in.FilenameOrigin != "" {
		exists.FilenameOrigin = in.FilenameOrigin
	}
	if in.Category != "" {
		exists.Category = in.Category
	}
	if in.Path != "" {
		exists.Path = in.Path
	}
	if in.Type != "" {
		exists.Type = in.Type
	}
	if in.Mime != "" {
		exists.Mime = in.Mime
	}
	if in.Extension != "" {
		exists.Extension = in.Extension
	}
	if in.Size != 0 {
		exists.Size = in.Size
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updated, err := s.UploadRepository.Update(tx, exists)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	out, err := s.UploadRepository.FindById(nil, updated.ID.String(), true)
	if err != nil {
		return updated, nil
	}
	return out, nil
}

func (s *UploadService) DeleteUpload(uploadID string, isHardDelete bool) error {
	exists, err := s.UploadRepository.FindById(nil, uploadID, true)
	if err != nil {
		return err
	}
	if exists == nil {
		return errors.New("upload not found")
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.UploadRepository.Delete(tx, uploadID, isHardDelete); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
