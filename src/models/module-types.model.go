package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModuleType struct {
	ID          uuid.UUID     `gorm:"type:uuid;primary_key" json:"id"`
	Name        string        `gorm:"not null" json:"name"`
	Icon        string        `gorm:"type:varchar(255)" json:"icon"`
	Description string        `json:"description"`

	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type ResponseGetModuleType struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Icon        string        `json:"icon"`
	Description string        `json:"description"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
}


type ModuleTypeCreateRequest struct {
	Name        string       `json:"name" validate:"required"`
	Icon        string       `json:"icon"`
	Description string       `json:"description"`
}

type ModuleTypeUpdateRequest struct {
	Name        string       `json:"name"`
	Icon        string       `json:"icon"`
	Description string       `json:"description"`
}

type ModuleTypeIsHardDeleteRequest struct {
	IDs []uuid.UUID `json:"ids" xml:"id" form:"ids" validate:"required"`
	IsHardDelete string `json:"is_hard_delete" xml:"is_hard_delete" form:"is_hard_delete" validate:"required"`
}

type ModuleTypeRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" xml:"ids" form:"ids" validate:"required"`
}