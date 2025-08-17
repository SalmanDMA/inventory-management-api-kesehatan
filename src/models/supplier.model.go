package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Supplier struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	Code          string         `gorm:"uniqueIndex;not null" json:"code"`
	Email         *string         `gorm:"uniqueIndex" json:"email"`
	Phone         *string         `json:"phone"`
	Address       *string         `json:"address"`
	ContactPerson *string         `json:"contact_person"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	PurchaseOrders []PurchaseOrder `gorm:"foreignKey:SupplierID" json:"purchase_orders,omitempty"`
}

type ResponseGetSupplier struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Code          string         `json:"code"`
	Email         *string         `json:"email"`
	Phone         *string         `json:"phone"`
	Address       *string         `json:"address"`
	ContactPerson *string         `json:"contact_person"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type SupplierCreateRequest struct {
	Name          string `json:"name" xml:"name" form:"name" validate:"required"`
	Code          string `json:"code" xml:"code" form:"code" validate:"required"`
	Email         *string `json:"email" xml:"email" form:"email" validate:"email"`
	Phone         *string `json:"phone" xml:"phone" form:"phone"`
	Address       *string `json:"address" xml:"address" form:"address"`
	ContactPerson *string `json:"contact_person" xml:"contact_person" form:"contact_person"`
}

type SupplierUpdateRequest struct {
	Name          string `json:"name" xml:"name" form:"name"`
	Code          string `json:"code" xml:"code" form:"code"`
	Email         *string `json:"email" xml:"email" form:"email" validate:"omitempty,email"`
	Phone         *string `json:"phone" xml:"phone" form:"phone"`
	Address       *string `json:"address" xml:"address" form:"address"`
	ContactPerson *string `json:"contact_person" xml:"contact_person" form:"contact_person"`
}

type SupplierIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type SupplierRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}
