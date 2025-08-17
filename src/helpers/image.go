package helpers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

func GetFileURL(upload *models.Upload) string {
	publicBase := os.Getenv("MINIO_BUCKET_URL")
	if publicBase == "" {
		publicBase = fmt.Sprintf("http://%s:%s/%s",
			os.Getenv("MINIO_ENDPOINT"),
			os.Getenv("MINIO_PORT"),
			strings.TrimLeft(os.Getenv("MINIO_BUCKET_NAME"), "/"),
		)
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(publicBase, "/"), strings.TrimLeft(upload.Filename, "/"))
}

func detectContentType(f multipart.File, header *multipart.FileHeader) (string, error) {
	var buf [512]byte
	var n int
	var err error

	rs, ok := f.(io.ReadSeeker)
	if !ok {
		var b bytes.Buffer
		if _, err = io.Copy(&b, f); err != nil {
			return "", fmt.Errorf("copy for sniff failed: %w", err)
		}
		ct := http.DetectContentType(b.Bytes())
		return ct, nil
	}

	n, err = rs.Read(buf[:])
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("sniff read failed: %w", err)
	}
	_, _ = rs.Seek(0, io.SeekStart)

	ct := http.DetectContentType(buf[:n])
	if h := header.Header.Get("Content-Type"); h != "" && h != "application/octet-stream" {
		ct = h
	}

	if ct == "" || ct == "application/octet-stream" {
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != "" {
			if byExt := mime.TypeByExtension(ext); byExt != "" {
				ct = byExt
			}
		}
	}

	if ct == "" {
		ct = "application/octet-stream"
	}
	return ct, nil
}

func classifyFile(contentType, filename string) string {
	ct := strings.ToLower(contentType)
	switch {
	case strings.HasPrefix(ct, "image/"):
		return "image"
	case strings.HasPrefix(ct, "video/"):
		return "video"
	case strings.HasPrefix(ct, "audio/"):
		return "audio"
	case strings.HasPrefix(ct, "application/pdf"),
		strings.HasPrefix(ct, "application/msword"),
		strings.HasPrefix(ct, "application/vnd.openxmlformats-officedocument"),
		strings.HasPrefix(ct, "text/plain"),
		strings.HasPrefix(ct, "application/vnd.ms-excel"),
		strings.HasPrefix(ct, "application/vnd.ms-powerpoint"),
		strings.HasPrefix(ct, "application/rtf"):
		return "document"
	case strings.HasPrefix(ct, "application/zip"),
		strings.HasPrefix(ct, "application/x-7z-compressed"),
		strings.HasPrefix(ct, "application/x-rar"),
		strings.HasPrefix(ct, "application/x-tar"),
		strings.HasPrefix(ct, "application/gzip"):
		return "archive"
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp", ".bmp", ".tiff":
		return "image"
	case ".mp4", ".mov", ".mkv", ".avi", ".webm", ".m4v":
		return "video"
	case ".mp3", ".wav", ".m4a", ".aac", ".flac", ".ogg":
		return "audio"
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".rtf", ".csv", ".md":
		return "document"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	default:
		return "other"
	}
}

func validateFile(header *multipart.FileHeader, category string) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".svg": true, ".webp": true, ".bmp": true, ".tiff": true,
		".mp4": true, ".mov": true, ".mkv": true, ".avi": true, ".webm": true, ".m4v": true,
		".mp3": true, ".wav": true, ".m4a": true, ".aac": true, ".flac": true, ".ogg": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".txt": true, ".rtf": true, ".csv": true, ".md": true,
		".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true,
	}
	if !allowed[ext] {
		return fiber.NewError(400, "invalid file extension")
	}

	const (
		maxImage   = 5 * 1024 * 1024
		maxAudio   = 20 * 1024 * 1024
		maxVideo   = 100 * 1024 * 1024
		maxDoc     = 20 * 1024 * 1024
		maxArchive = 100 * 1024 * 1024
		maxOther   = 10 * 1024 * 1024
	)
	switch category {
	case "image":
		if header.Size > maxImage {
			return fiber.NewError(400, "file size too large (image max 5MB)")
		}
	case "audio":
		if header.Size > maxAudio {
			return fiber.NewError(400, "file size too large (audio max 20MB)")
		}
	case "video":
		if header.Size > maxVideo {
			return fiber.NewError(400, "file size too large (video max 100MB)")
		}
	case "document":
		if header.Size > maxDoc {
			return fiber.NewError(400, "file size too large (document max 20MB)")
		}
	case "archive":
		if header.Size > maxArchive {
			return fiber.NewError(400, "file size too large (archive max 100MB)")
		}
	default:
		if header.Size > maxOther {
			return fiber.NewError(400, "file size too large (other max 10MB)")
		}
	}
	return nil
}

func SaveFile(ctx *fiber.Ctx, file *multipart.FileHeader, tableName string) (string, error) {
	if file == nil {
		return "", errors.New("file is required")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	contentType, err := detectContentType(src, file)
	if err != nil {
		return "", fiber.NewError(400, fmt.Sprintf("failed to detect content type: %v", err))
	}

	category := classifyFile(contentType, file.Filename)
	if err := validateFile(file, category); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	uniqueFileName := uuid.New().String() + ext
	today := time.Now().Format("2006-01-02")
	objectKey := fmt.Sprintf("%s/%s/%s/%s", tableName, today, category, uniqueFileName)

	if rs, ok := src.(io.ReadSeeker); ok {
		_, _ = rs.Seek(0, io.SeekStart)
	}

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}
	info, err := configs.Minio.PutObject(
		context.Background(),
		os.Getenv("MINIO_BUCKET_NAME"),
		objectKey,
		src,
		file.Size,
		opts,
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}

	meta := &models.Upload{
		ID:             uuid.New(),
		Filename:       objectKey,
		FilenameOrigin: file.Filename,
		Category:       tableName,
		Path:           fmt.Sprintf("%s/%s/%s", tableName, category, today),
		Type:           contentType,
		Mime:           contentType,
		Extension:      strings.TrimPrefix(ext, "."),
		Size:           info.Size,
		Bucket:         os.Getenv("MINIO_BUCKET_NAME"),
	}

	upload, err := repositories.NewUploadRepository(configs.DB).Insert(meta)
	if err != nil {
		_ = configs.Minio.RemoveObject(context.Background(), os.Getenv("MINIO_BUCKET_NAME"), objectKey, minio.RemoveObjectOptions{})
		return "", fmt.Errorf("failed to save upload metadata: %w", err)
	}

	return upload.ID.String(), nil
}

func DeleteLocalFileImmediate(uploadId string) error {
	uploadRepo := repositories.NewUploadRepository(configs.DB)
	upload, err := uploadRepo.FindById(uploadId, false)
	if err != nil {
		return fmt.Errorf("failed to find upload: %w", err)
	}

	err = configs.Minio.RemoveObject(
		context.Background(),
		upload.Bucket,
		upload.Filename,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	if err := uploadRepo.Delete(uploadId, true); err != nil {
		return fmt.Errorf("object deleted but failed to delete DB record: %w", err)
	}
	return nil
}
