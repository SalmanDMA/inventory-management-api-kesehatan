package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Type      string         `gorm:"type:text;not null" json:"type"`// "low_stock", "system_alert", "other"
	Title     string         `gorm:"type:text;not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	IsRead    bool           `gorm:"default:false" json:"is_read"`
	ReadAt    *time.Time     `json:"read_at,omitempty"`
	Metadata  JSONB          `gorm:"type:jsonb" json:"metadata"`
	User      User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type ResponseGetNotification struct {
	ID          uuid.UUID      `json:"id"`
	UserID      uuid.UUID      `json:"user_id"`
	Type        string         `json:"type"`
	Title       string         `json:"title"`
	Message     string         `json:"message"`
	IsRead      bool           `json:"is_read"`
	ReadAt      *time.Time     `json:"read_at,omitempty"`
	Metadata    JSONB          `json:"metadata"`
	User        User           `json:"user"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type NotificationIsHardDeleteRequest struct {
	IsHardDelete string `json:"is_hard_delete" xml:"is_hard_delete" form:"is_hard_delete" validate:"required"`
	IDs []uuid.UUID `json:"ids" xml:"ids" form:"ids" validate:"required,dive,required"`
}

type NotificationRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" xml:"ids" form:"ids" validate:"required,dive,required"`
}

type NotificationMarkMultipleRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
}