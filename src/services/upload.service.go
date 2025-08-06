package services

import (
	"errors"

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

func (service *UploadService) GetAllUploads() ([]models.ResponseGetUpload, error) {
	uploads, err := service.UploadRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var uploadsResponse []models.ResponseGetUpload
	for _, upload := range uploads {
		uploadsResponse = append(uploadsResponse, models.ResponseGetUpload{
			ID:          upload.ID,
		 Filename:    upload.Filename,
			FilenameOrigin: upload.FilenameOrigin,
			Category: upload.Category,
			Path: upload.Path,
			Type: upload.Type,
			Mime: upload.Mime,
			Extension: upload.Extension,
			Size: upload.Size,
			CreatedAt: upload.CreatedAt,
			UpdatedAt: upload.UpdatedAt,
			DeletedAt: upload.DeletedAt,
		})
	}

	return uploadsResponse, nil
}

func (service *UploadService) CreateUpload(uploadRequest *models.UploadCreateRequest) ( *models.Upload, error) {
	newUpload := &models.Upload{
		ID: 									uuid.New(),
		Filename:    uploadRequest.Filename,
		FilenameOrigin: uploadRequest.FilenameOrigin,
		Category: uploadRequest.Category,
		Path: uploadRequest.Path,
		Type: uploadRequest.Type,
		Mime: uploadRequest.Mime,
		Extension: uploadRequest.Extension,
		Size: uploadRequest.Size,
	}

	upload, err := service.UploadRepository.Insert(newUpload)

	if err != nil {
		return nil, err
	}

	return upload, nil
}

func (service *UploadService) UpdateUpload(uploadID string, uploadUpdate *models.UploadUpdateRequest) (*models.Upload, error) {
	uploadExists, err := service.UploadRepository.FindById(uploadID, true)
	if err != nil {
		return nil, err
	}
	if uploadExists == nil {
		return nil, errors.New("upload not found")
	}

	if uploadUpdate.Filename != "" {
		uploadExists.Filename = uploadUpdate.Filename
	}

	if uploadUpdate.FilenameOrigin != "" {
		uploadExists.FilenameOrigin = uploadUpdate.FilenameOrigin
	}

	if uploadUpdate.Category != "" {
		uploadExists.Category = uploadUpdate.Category
	}

	if uploadUpdate.Path != "" {
		uploadExists.Path = uploadUpdate.Path
	}

	if uploadUpdate.Type != "" {
		uploadExists.Type = uploadUpdate.Type
	}

	if uploadUpdate.Mime != "" {
		uploadExists.Mime = uploadUpdate.Mime
	}

	if uploadUpdate.Extension != "" {
		uploadExists.Extension = uploadUpdate.Extension
	}

	if uploadUpdate.Size != 0 {
		uploadExists.Size = uploadUpdate.Size
	}

	updateUpload , err := service.UploadRepository.Update(uploadExists)
	if err != nil {
		return nil, err
	}

	return updateUpload, nil
}

func (service *UploadService) DeleteUpload(uploadID string, isHardDelete bool) error {
	uploadExists, err := service.UploadRepository.FindById(uploadID, true)
	if err != nil {
		return err
	}
	if uploadExists == nil {
		return errors.New("upload not found")
	}

	if isHardDelete {
		if err := service.UploadRepository.Delete(uploadID, true); err != nil {
			return err
		}
	} else {
		if err := service.UploadRepository.Delete(uploadID, false); err != nil {
			return err
		}
	}

	return nil
}