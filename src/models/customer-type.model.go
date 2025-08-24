package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerType struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name  string `gorm:"size:100;uniqueIndex:ux_ft_name_ci;not null" json:"name"`
	Color string `gorm:"size:20" json:"color"`
	Description *string `json:"description"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Customers []Customer `gorm:"foreignKey:CustomerTypeID;references:ID" json:"customers,omitempty"`
}

type ResponseGetCustomerType struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
	Description *string `json:"description"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type CustomerTypeCreateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Description *string `json:"description"`
}

type CustomerTypeIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type CustomerTypeRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}