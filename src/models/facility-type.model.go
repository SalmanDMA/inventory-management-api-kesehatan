package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FacilityType struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name  string `gorm:"size:100;uniqueIndex:ux_ft_name_ci;not null" json:"name"`
	Color string `gorm:"size:20" json:"color"`
	Description *string `json:"description"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Facilities []Facility `gorm:"foreignKey:FacilityTypeID;references:ID" json:"facilities,omitempty"`
}

type ResponseGetFacilityType struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
	Description *string `json:"description"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type FacilityTypeCreateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Description *string `json:"description"`
}

type FacilityTypeIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type FacilityTypeRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}