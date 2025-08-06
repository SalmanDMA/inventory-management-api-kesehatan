package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ItemHistory struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ItemID       uuid.UUID      `gorm:"type:uuid;not null" json:"item_id"`
	ChangeType   string         `gorm:"not null" json:"change_type"` // enum: create_price, create_stock, update_stock, update_price,
	Description  string         `json:"description"`
	OldPrice     int            `json:"old_price"`
	NewPrice     int            `json:"new_price"`
	CurrentPrice int            `json:"current_price"`
	OldStock     int            `json:"old_stock"`
	NewStock     int            `json:"new_stock"`
	CurrentStock int            `json:"current_stock"`
	CreatedBy    *uuid.UUID     `gorm:"type:uuid" json:"created_by"`  // nullable agar bisa SET NULL
	UpdatedBy    *uuid.UUID     `gorm:"type:uuid" json:"updated_by"`  // nullable
	DeletedBy    *uuid.UUID     `gorm:"type:uuid" json:"deleted_by"`  // nullable
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Item          Item  `gorm:"foreignKey:ItemID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"item"`
	CreatedByUser *User `gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"created_by_user,omitempty"`
	UpdatedByUser *User `gorm:"foreignKey:UpdatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"updated_by_user,omitempty"`
	DeletedByUser *User `gorm:"foreignKey:DeletedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"deleted_by_user,omitempty"`
}

type ResponseGetItemHistory struct {
	ID           uuid.UUID      `json:"id"`
	ItemID       uuid.UUID      `json:"item_id"`
	ChangeType   string         `json:"change_type"`
	Description  string         `json:"description"`
	OldPrice     int            `json:"old_price"`
	NewPrice     int            `json:"new_price"`
	CurrentPrice int            `json:"current_price"`
	OldStock     int            `json:"old_stock"`
	NewStock     int            `json:"new_stock"`
	CurrentStock int            `json:"current_stock"`
	CreatedBy    uuid.UUID      `json:"created_by"`
	UpdatedBy    uuid.UUID      `json:"updated_by"`
	DeletedBy    uuid.UUID      `json:"deleted_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Item          Item  `json:"item"`
	CreatedByUser *User `json:"created_by_user,omitempty"`
	UpdatedByUser *User `json:"updated_by_user,omitempty"`
	DeletedByUser *User `json:"deleted_by_user,omitempty"`
}

type ItemHistoryCreateRequest struct {
	ItemID      uuid.UUID `json:"item_id" validate:"required"`
	ChangeType  string    `json:"change_type" validate:"required"`
	NewPrice    int       `json:"new_price"`
	NewStock    int       `json:"new_stock"`
	Description string    `json:"description" validate:"required"`
}

type ItemHistoryIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type ItemHistoryRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}