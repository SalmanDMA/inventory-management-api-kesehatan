package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Item struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Code        string         `gorm:"uniqueIndex;not null" json:"code"`
	CategoryID  uuid.UUID      `gorm:"type:uuid" json:"category_id"`
	Price       int            `gorm:"not null" json:"price"`
	Stock       int            `gorm:"not null" json:"stock"`
	ImageID     *uuid.UUID     `gorm:"type:uuid" json:"image_id,omitempty"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Image       *Upload        `gorm:"foreignKey:ImageID" json:"image,omitempty"`
	Category    Category       `gorm:"foreignKey:CategoryID" json:"category"`
	ItemHistories []ItemHistory `gorm:"foreignKey:ItemID" json:"item_histories,omitempty"`
}

type ResponseGetItem struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Code        string         `json:"code"`
	Price       int            `json:"price"`
	Stock       int            `json:"stock"`
	ImageID     *uuid.UUID     `json:"image_id,omitempty"`
	CategoryID  uuid.UUID      `json:"category_id"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Image       *Upload        `json:"image,omitempty"`
	Category    Category       `json:"category"`
	ItemHistories []ItemHistory `json:"item_histories,omitempty"`
}

type ItemCreateRequest struct {
	Name        string         `json:"name" xml:"name" form:"name" validate:"required"`
	Code        string         `json:"code" xml:"code" form:"code" validate:"required"`
	Price       int 											`json:"price" xml:"price" form:"price" validate:"required"`
	Stock       int 											`json:"stock" xml:"stock" form:"stock" validate:"required"`
	CategoryID  uuid.UUID      `json:"category_id" xml:"category_id" form:"category_id" validate:"required"`
	Description string         `json:"description" xml:"description" form:"description"`
}

type ItemUpdateRequest struct {
	Name        string         `json:"name" xml:"name" form:"name"`
	Code        string         `json:"code" xml:"code" form:"code"`
	Price       int 			`json:"price" xml:"price" form:"price"`
	Stock       int 			`json:"stock" xml:"stock" form:"stock"`
	CategoryID  uuid.UUID      `json:"category_id" xml:"category_id" form:"category_id"`
	Description string         `json:"description" xml:"description" form:"description"`
}

type ItemIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type ItemRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}