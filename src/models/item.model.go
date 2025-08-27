package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Item struct {
		ID         uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
		Name       string         `gorm:"not null" json:"name"`
		Code       string         `gorm:"uniqueIndex;not null" json:"code"`
		CategoryID uuid.UUID      `gorm:"type:uuid" json:"category_id"`
		UoMID      uuid.UUID      `gorm:"column:uom_id;type:uuid;not null" json:"uom_id"`
		Price      int            `gorm:"not null" json:"price"`
		Stock      int            `gorm:"not null" json:"stock"`
		LowStock   int            `gorm:"not null" json:"low_stock"`
		ImageID    *uuid.UUID     `gorm:"type:uuid" json:"image_id,omitempty"`
		Description string        `json:"description"`
		Batch       int           `gorm:"default:0" json:"batch"`
		IsConsignment bool        `gorm:"default:false" json:"is_consignment"`
		DueDate    *time.Time     `json:"due_date"`
		ExpiredAt  time.Time     `json:"expired_at"`
		CreatedAt  time.Time      `json:"created_at"`
		UpdatedAt  time.Time      `json:"updated_at"`
		DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

		UoM           UoM           `gorm:"foreignKey:UoMID" json:"uom"`
		Image         *Upload       `gorm:"foreignKey:ImageID;constraint:onUpdate:CASCADE,onDelete:SET NULL;" json:"image,omitempty"`
		Category      Category      `gorm:"foreignKey:CategoryID" json:"category"`
		ItemHistories []ItemHistory `gorm:"foreignKey:ItemID" json:"item_histories,omitempty"`
}

type ResponseGetItem struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Code        string         `json:"code"`
	Price       int            `json:"price"`
	Stock       int            `json:"stock"`
	LowStock    int            `json:"low_stock"`
	ImageID     *uuid.UUID     `json:"image_id,omitempty"`
	CategoryID  uuid.UUID      `json:"category_id"`
	UoMID       uuid.UUID      `json:"uom_id"`
	Description string         `json:"description"`
	Batch       int            `json:"batch"`
	IsConsignment bool        `json:"is_consignment"`
	DueDate     *time.Time      `json:"due_date"`
	ExpiredAt   time.Time      `json:"expired_at"`

	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	UoM         UoM            `json:"uom"`
	Image       *Upload        `json:"image,omitempty"`
	Category    Category       `json:"category"`
	ItemHistories []ItemHistory `json:"item_histories,omitempty"`
}

type ItemCreateRequest struct {
	Name        string         `json:"name" xml:"name" form:"name" validate:"required"`
	Code        string         `json:"code" xml:"code" form:"code" validate:"required"`
	Price       int 											`json:"price" xml:"price" form:"price" validate:"required"`
	Stock       int 											`json:"stock" xml:"stock" form:"stock" validate:"required"`
	LowStock    int 											`json:"low_stock" xml:"low_stock" form:"low_stock" validate:"required"`
	CategoryID  uuid.UUID      `json:"category_id" xml:"category_id" form:"category_id" validate:"required"`
	UoMID       uuid.UUID      `json:"uom_id" xml:"uom_id" form:"uom_id" validate:"required"`
	Description string         `json:"description" xml:"description" form:"description"`
	Batch       int            `json:"batch" xml:"batch" form:"batch" validate:"required"`
	IsConsignment bool        `json:"is_consignment" xml:"is_consignment" form:"is_consignment"`
	DueDate     *time.Time      `json:"due_date" xml:"due_date" form:"due_date"`
	ExpiredAt   time.Time      `json:"expired_at" xml:"expired_at" form:"expired_at" validate:"required"`
}

type ItemUpdateRequest struct {
	Name        string         `json:"name" xml:"name" form:"name"`
	Code        string         `json:"code" xml:"code" form:"code"`
	Price       int 			`json:"price" xml:"price" form:"price"`
	Stock       int 			`json:"stock" xml:"stock" form:"stock"`
	LowStock    int 			`json:"low_stock" xml:"low_stock" form:"low_stock"`
	CategoryID  uuid.UUID      `json:"category_id" xml:"category_id" form:"category_id"`
	UoMID       uuid.UUID      `json:"uom_id" xml:"uom_id" form:"uom_id"`
	Description string         `json:"description" xml:"description" form:"description"`
	Batch       int            `json:"batch" xml:"batch" form:"batch"`
	IsConsignment bool        `json:"is_consignment" xml:"is_consignment" form:"is_consignment"`
	DueDate     *time.Time      `json:"due_date" xml:"due_date" form:"due_date"`
	ExpiredAt   time.Time      `json:"expired_at" xml:"expired_at" form:"expired_at"`
}

type ItemIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type ItemRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}