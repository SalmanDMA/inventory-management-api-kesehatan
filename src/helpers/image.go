package helpers

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetFileURL(upload *models.Upload) string {
	baseAPI := os.Getenv("BASE_API")
	if baseAPI == "" {
		baseAPI = "http://localhost:8000"
	}

	return fmt.Sprintf("%s/uploads/%s", baseAPI, upload.Filename)
}

func SaveFile(ctx *fiber.Ctx, file *multipart.FileHeader, tableName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fiber.ErrBadRequest
	}

	if file.Size > 2*1024*1024 {
		return "", fiber.ErrBadRequest
	}

	uniqueFileName := uuid.New().String() + ext
	today := time.Now().Format("2006-01-02") 
	saveDir := filepath.Join("public", "uploads", tableName, today)

	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return "", err
	}

	fullPath := filepath.Join(saveDir, uniqueFileName)
	if err := ctx.SaveFile(file, fullPath); err != nil {
		return "", err
	}

	meta := &models.Upload{
		ID:             uuid.New(),
		Filename:       fmt.Sprintf("%s/%s/%s", tableName, today, uniqueFileName),
		FilenameOrigin: file.Filename,
		Category:       tableName,
		Path:           fmt.Sprintf("%s/%s", tableName, today),
		Type:           file.Header.Get("Content-Type"),
		Mime:           file.Header.Get("Content-Type"),
		Extension:      strings.TrimPrefix(ext, "."),
		Size:           int64(file.Size),
	}

	upload, err := repositories.NewUploadRepository(configs.DB).Insert(meta)
	if err != nil {
		return "", err
	}

	return upload.ID.String(), nil
}

func DeleteLocalFileImmediate(uploadId string) error {
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	upload, err := uploadRepo.FindById(uploadId, false)
	if err != nil {
		return fmt.Errorf("failed to find upload: %w", err)
	}

	filePath := filepath.Join("public", "uploads", filepath.FromSlash(upload.Filename))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if err := uploadRepo.Delete(uploadId, true); err != nil {
		return fmt.Errorf("file deleted but failed to delete DB record: %w", err)
	}

	return nil
}




