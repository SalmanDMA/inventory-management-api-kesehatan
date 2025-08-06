package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Module struct {
	ID            int            `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	ParentID      *int           `gorm:"type:int" json:"parent_id"`
	Parent        *Module        `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"parent"`

	ModuleTypeID  uuid.UUID      `gorm:"type:uuid;not null" json:"module_type_id"`
	ModuleType    ModuleType     `gorm:"foreignKey:ModuleTypeID" json:"module_type"`

	Path          string         `gorm:"type:varchar(255)" json:"path"`
	Icon          string         `gorm:"type:varchar(255)" json:"icon"`
	Route         string         `gorm:"type:varchar(255)" json:"route"`
	Description   string         `json:"description"`

	Children      []Module       `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"children"`
	RoleModules   []RoleModule   `gorm:"foreignKey:ModuleID" json:"role_modules"`

	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
	

type ResponseGetModule struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	ParentID     *int           `json:"parent_id"`
	Parent       *Module        `json:"parent"`
	ModuleTypeID uuid.UUID      `json:"module_type_id"`
	ModuleType   ModuleType     `json:"module_type"`
	Path         string         `json:"path"`         
	Route        string         `json:"route"`
	Icon         string         `json:"icon"`         
	Description  string         `json:"description"`
	RoleModules  []RoleModule   `json:"role_modules"`
	Children     []Module       `json:"children"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at"`
}

type ModuleCreateRequest struct {
	Name         string     `json:"name" validate:"required"`
	ParentID     *int       `json:"parent_id"`               
	ModuleTypeID uuid.UUID  `json:"module_type_id" validate:"required"`
	Path         string     `json:"path"`
	Route        string     `json:"route"`
	Icon         string     `json:"icon"`                    
	Description  string     `json:"description"`
}

type ModuleUpdateRequest struct {
	Name         string     `json:"name"`
	ParentID     *int       `json:"parent_id"`
	ModuleTypeID uuid.UUID  `json:"module_type_id"`
	Path         string     `json:"path"`
	Route        string     `json:"route"`
	Icon         string     `json:"icon"`
	Description  string     `json:"description"`
}

type ModuleIsHardDeleteRequest struct {
	IDs []int `json:"ids" xml:"id" form:"ids" validate:"required"`
	IsHardDelete string `json:"is_hard_delete" xml:"is_hard_delete" form:"is_hard_delete" validate:"required"`
}

type ModuleRestoreRequest struct {
	IDs []int `json:"ids" xml:"ids" form:"ids" validate:"required"`
}