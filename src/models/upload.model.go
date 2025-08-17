package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Upload struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Filename       string         `gorm:"not null" json:"filename"`
	FilenameOrigin string         `json:"filename_origin"`
	Category       string         `json:"category"`
	Path           string         `json:"path"`
	Type           string         `json:"type"`
	Mime           string         `json:"mime"`
	Extension      string         `json:"extension"`
	Size           int64          `json:"size"`
	Bucket         string 									`json:"bucket"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type ResponseGetUpload struct {
	ID uuid.UUID `json:"id"`
	Filename        string `json:"filename"`
	FilenameOrigin  string `json:"filename_origin"`
	Category 						string `json:"category"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Mime        string `json:"mime"`
	Extension  	string `json:"extension"`
	Size        int64  `json:"size"`
	Bucket      string  `json:"bucket"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at"`
} 

type ResponseGetFile struct {
	URL string `json:"url"`
}

type UploadCreateRequest struct {
	Filename        string `json:"filename" validate:"required"`
	FilenameOrigin  string `json:"filename_origin"`	
	Category 						string `json:"category" validate:"required"`
	Path        string `json:"path" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Mime        string `json:"mime" validate:"required"`
	Extension  	string `json:"extension" validate:"required"`
	Size        int64  `json:"size" validate:"required"`
}

type UploadUpdateRequest struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	Filename        string `json:"filename" validate:"required"`
	FilenameOrigin  string `json:"filename_origin"`
	Category 						string `json:"category"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Mime        string `json:"mime"`
	Extension  	string `json:"extension"`
	Size        int64  `json:"size"`
}