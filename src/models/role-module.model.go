package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleModule struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	RoleID    *uuid.UUID     `gorm:"type:uuid" json:"role_id"`
	Role      *Role          `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	ModuleID  *int     `gorm:"type:uuid" json:"module_id"`
	Module    *Module        `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Checked   bool           `json:"checked"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}


type ResponseGetRoleModule struct {
	ID        uuid.UUID `json:"id"`
	RoleID   	*uuid.UUID `json:"role_id"`
	Role      *Role      `json:"role"`
	ModuleID  *int `json:"module_id"`
	Module    *Module    `json:"module"`
	Checked   bool      `json:"checked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
} 

type RoleModuleRequest struct {
	ModuleID  int `json:"module_id" validate:"required"`
	Checked   bool      `json:"checked"`
}