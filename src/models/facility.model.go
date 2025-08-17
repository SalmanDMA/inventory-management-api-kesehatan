package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Facility struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code           string    `gorm:"size:30;uniqueIndex:ux_fac_code;not null" json:"code"`
	Name           string    `gorm:"size:150;index;not null" json:"name"`
	FacilityTypeID uuid.UUID `gorm:"type:uuid;index;not null" json:"facility_type_id"`
	AreaID         uuid.UUID `gorm:"type:uuid;index;not null" json:"area_id"`

	Address   *string  `gorm:"size:255" json:"address,omitempty"`
	Phone     *string  `gorm:"size:30" json:"phone,omitempty"`
	Email     *string  `gorm:"size:120" json:"email,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	FacilityType FacilityType `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:FacilityTypeID;references:ID" json:"facility_type"`
	Area         Area         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:AreaID;references:ID" json:"area"`
}

type ResponseGetFacility struct {
	ID    uuid.UUID `json:"id"`
	Code  string    `json:"code"`
	Name  string    `json:"name"`
	FacilityTypeID uuid.UUID `json:"facility_type_id"`
	AreaID         uuid.UUID `json:"area_id"`
	Address   *string  `json:"address,omitempty"`
	Phone     *string  `json:"phone,omitempty"`
	Email     *string  `json:"email,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	FacilityType FacilityType `json:"facility_type"`
	Area         Area `json:"area"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type FacilityCreateRequest struct {
	Code           string    `json:"code" validate:"required"`
	Name           string    `json:"name" validate:"required"`
	FacilityTypeID uuid.UUID `json:"facility_type_id" validate:"required"`
	AreaID         uuid.UUID `json:"area_id" validate:"required"`
	Address   *string  `json:"address,omitempty"`
	Phone     *string  `json:"phone,omitempty"`
	Email     *string  `json:"email,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

type FacilityIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type FacilityRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}