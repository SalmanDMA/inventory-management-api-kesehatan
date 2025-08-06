package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Color       string         `json:"color"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Items       []Item         `gorm:"foreignKey:CategoryID" json:"items,omitempty"`
}

type ResponseGetCategory struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Color       string         `json:"color"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
	Items       []Item         `json:"items,omitempty"`
}

type CategoryCreateRequest struct {
	Name        string `json:"name" validate:"required"`
	Color       string `json:"color" validate:"required"`
	Description string `json:"description"`
}

type CategoryUpdateRequest struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

type CategoryIsHardDeleteRequest struct {
	IsHardDelete string `json:"is_hard_delete" xml:"is_hard_delete" form:"is_hard_delete" validate:"required"`
	IDs []uuid.UUID `json:"ids" xml:"ids" form:"ids" validate:"required,dive,required"`
}

type CategoryRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" xml:"ids" form:"ids" validate:"required,dive,required"`
}