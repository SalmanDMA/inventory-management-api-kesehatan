package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (SalesPerson) TableName() string { return "sales_person" }

type SalesPerson struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name     string     `gorm:"size:120;index;not null" json:"name"`
	Phone    *string    `gorm:"size:30" json:"phone,omitempty"`
	Email    *string    `gorm:"size:120" json:"email,omitempty"`
	HireDate *time.Time `json:"hire_date,omitempty"`
	Address  *string    `gorm:"size:255" json:"address,omitempty"`
	NPWP					*string    `gorm:"size:30" json:"npwp,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Assignments []SalesAssignment `gorm:"foreignKey:SalesPersonID;references:ID" json:"assignments,omitempty"`
}

type ResponseGetSalesPerson struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Phone *string   `json:"phone,omitempty"`
	Email *string   `json:"email,omitempty"`
	HireDate *time.Time `json:"hire_date,omitempty"`
	Address  *string    `gorm:"size:255" json:"address,omitempty"`
	NPWP    *string    `gorm:"size:30" json:"npwp,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Assignments []SalesAssignment `gorm:"foreignKey:SalesPersonID;references:ID" json:"assignments,omitempty"`
}

type SalesPersonCreateRequest struct {
	Name     string     `json:"name" validate:"required"`
	Phone    *string    `json:"phone,omitempty"`
	Email    *string    `json:"email,omitempty"`
	HireDate *time.Time `json:"hire_date,omitempty"`
	Address  *string    `json:"address,omitempty"`
	NPWP     *string    `json:"npwp,omitempty"`
}

type SalesPersonUpdateRequest struct {
	Name     string     `json:"name"`
	Phone    *string    `json:"phone,omitempty"`
	Email    *string    `json:"email,omitempty"`
	HireDate *time.Time `json:"hire_date,omitempty"`
	Address  *string    `json:"address,omitempty"`
	NPWP     *string    `json:"npwp,omitempty"`
}

type SalesPersonIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type SalesPersonRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}