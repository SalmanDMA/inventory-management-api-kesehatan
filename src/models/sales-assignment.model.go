package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesAssignment struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SalesPersonID uuid.UUID    `gorm:"type:uuid;index;not null" json:"sales_person_id"`
	AreaID        uuid.UUID    `gorm:"type:uuid;index;not null" json:"area_id"`
	Checked       bool      `json:"checked"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	SalesPerson SalesPerson `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:SalesPersonID;references:ID" json:"sales_person"`
	Area        Area        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:AreaID;references:ID" json:"area"`
}

type ResponseGetSalesAssignment struct {
	ID    uuid.UUID `json:"id"`
	SalesPersonID uuid.UUID    `json:"sales_person_id"`
	AreaID        uuid.UUID   `json:"area_id"`
	Checked       bool      `json:"checked"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	SalesPerson SalesPerson `json:"sales_person"`
	Area        Area         `json:"area"`
}

type SalesAssignmentRequest struct {
	AreaID        uuid.UUID `json:"area_id"  validate:"required"`
	Checked       bool   `json:"checked"`
}