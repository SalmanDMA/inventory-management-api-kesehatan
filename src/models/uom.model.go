package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (UoM) TableName() string { return "uoms" }

type UoM struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"size:100;uniqueIndex:ux_ft_name_ci;not null" json:"name"`
	Color       string         `gorm:"size:20" json:"color"`
	Description *string        `json:"description"`

	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Items []Item `gorm:"foreignKey:UoMID" json:"items,omitempty"`
}

type ResponseGetUoM struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"size:100;uniqueIndex:ux_ft_name_ci;not null" json:"name"`
	Color       string         `gorm:"size:20" json:"color"`
	Description *string        `json:"description"`

	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Items []Item `gorm:"foreignKey:UoMID" json:"items,omitempty"`
}

type UoMCreateRequest struct {
	Name        string `json:"name" validate:"required"`
	Color       string `json:"color"`
	Description *string `json:"description"`
}

type UoMUpdateRequest struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description *string `json:"description"`
}

type UoMIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type UoMRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}