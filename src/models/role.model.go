package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID           uuid.UUID     `gorm:"type:uuid;primary_key" json:"id"`
	Name         string        `gorm:"not null" json:"name"`
	Alias        string        `json:"alias"`
	Color        string        `json:"color"`
	Description  string        `json:"description"`
	RoleModules  []RoleModule  `gorm:"foreignKey:RoleID" json:"role_modules"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type ResponseGetRole struct {
	ID uuid.UUID `json:"id"`
	Name string `json:"name"`
	Alias string `json:"alias"`
	Color string `json:"color"`
	Description string `json:"description"`
	RoleModules  []RoleModule `json:"role_modules"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at"`
} 

type RoleCreateRequest struct {
	Name string `json:"name" validate:"required"`
	Alias string `json:"alias"`
	Color string `json:"color"`
	Description string `json:"description"`
}

type RoleUpdateRequest struct {
	Name string `json:"name" validate:"required"`
	Alias string `json:"alias"`
	Color string `json:"color"`
	Description string `json:"description"`
}

type RoleIsHardDeleteRequest struct {
	IsHardDelete string `json:"is_hard_delete" xml:"is_hard_delete" form:"is_hard_delete" validate:"required"`
	IDs []uuid.UUID `json:"ids" xml:"ids" form:"ids" validate:"required,dive,required"`
}

type RoleRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" xml:"ids" form:"ids" validate:"required,dive,required"`
}