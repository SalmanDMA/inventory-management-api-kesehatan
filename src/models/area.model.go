package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Area struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string   `gorm:"size:20;uniqueIndex:ux_area_code;not null" json:"code"`
	Name      string   `gorm:"size:100;uniqueIndex:ux_area_name_ci;not null" json:"name"`
	Color string `gorm:"size:20" json:"color"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Customers []Customer `gorm:"foreignKey:AreaID;references:ID" json:"customers,omitempty"`
}

type ResponseGetArea struct {
	ID        uuid.UUID      `json:"id"`
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	Color string `json:"color"`
	Latitude  *float64    `json:"latitude"`
	Longitude *float64    `json:"longitude"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type AreaCreateRequest struct {
	Code string `json:"code" validate:"required"`
	Name string `json:"name" validate:"required"`
	Color string `json:"color"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type AreaIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type AreaRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}