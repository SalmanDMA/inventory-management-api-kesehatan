package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Customer struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nomor          string    `gorm:"size:30;uniqueIndex:ux_fac_nomor;not null" json:"nomor"`
	Name           string    `gorm:"size:150;index;not null" json:"name"`
	CustomerTypeID uuid.UUID `gorm:"type:uuid;index;not null" json:"customer_type_id"`
	AreaID         uuid.UUID `gorm:"type:uuid;index;not null" json:"area_id"`

	Address   *string  `gorm:"size:255" json:"address,omitempty"`
	Phone     *string  `gorm:"size:30" json:"phone,omitempty"`
	Email     *string  `gorm:"size:120" json:"email,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	CustomerType CustomerType `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:CustomerTypeID;references:ID" json:"customer_type"`
	Area         Area         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:AreaID;references:ID" json:"area"`
}

type ResponseGetCustomer struct {
	ID    uuid.UUID `json:"id"`
	Nomor  string    `json:"nomor"`
	Name  string    `json:"name"`
	CustomerTypeID uuid.UUID `json:"customer_type_id"`
	AreaID         uuid.UUID `json:"area_id"`
	Address   *string  `json:"address,omitempty"`
	Phone     *string  `json:"phone,omitempty"`
	Email     *string  `json:"email,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	CustomerType CustomerType `json:"customer_type"`
	Area         Area `json:"area"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type CustomerCreateRequest struct {
	Nomor           string    `json:"nomor" validate:"required"`
	Name           string    `json:"name" validate:"required"`
	CustomerTypeID uuid.UUID `json:"customer_type_id" validate:"required"`
	AreaID         uuid.UUID `json:"area_id" validate:"required"`
	Address   *string  `json:"address,omitempty"`
	Phone     *string  `json:"phone,omitempty"`
	Email     *string  `json:"email,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

type CustomerIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type CustomerRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}