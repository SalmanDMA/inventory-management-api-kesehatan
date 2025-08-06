package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Username              string         `gorm:"unique;not null" json:"username"`
	Password              string         `gorm:"not null" json:"password"`
	Name                  string         `gorm:"not null" json:"name"`
	Email                 string         `gorm:"unique;not null" json:"email"`
	Phone                 string         `json:"phone"`
	Address               string         `json:"address"`
	Description           string         `json:"description"`
	AvatarID              *uuid.UUID     `gorm:"type:uuid" json:"avatar_id"`
	Avatar 															*Upload 							`gorm:"foreignKey:AvatarID;constraint:onUpdate:CASCADE,onDelete:SET NULL;" json:"avatar"`
	RoleID                *uuid.UUID     `gorm:"type:uuid" json:"role_id"`
	Role                  *Role          `gorm:"foreignKey:RoleID" json:"role"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type 	UserCreate struct {
	Username    string    `json:"username" xml:"username" form:"username" validate:"required"`
	Name        string    `json:"name" xml:"name" form:"name" validate:"required,min=3"`
	Email       string    `json:"email" xml:"email" form:"email" validate:"required,email"`
	Address     string    `json:"address" xml:"address" form:"address"`
	Phone       string    `json:"phone" xml:"phone" form:"phone"`
	Password    string    `json:"password" xml:"password" form:"password" validate:"required,min=8"`
	AvatarID      uuid.UUID	`json:"avatar_id" xml:"avatar_id" form:"avatar_id"`
	RoleID      uuid.UUID `json:"role_id" xml:"role_id" form:"role_id" validate:"required"`
	Description string    `json:"description" xml:"description" form:"description"`
}

type UserUpdate struct {
	Username    string    `json:"username" xml:"username" form:"username"`
	Name        string    `json:"name" xml:"name" form:"name"`
	Email       string    `json:"email" xml:"email" form:"email"`
	Address     string    `json:"address" xml:"address" form:"address"`
	Password    string    `json:"password" xml:"password" form:"password"`
	Phone       string    `json:"phone" xml:"phone" form:"phone"`
	AvatarID      uuid.UUID `json:"avatar_id" xml:"avatar_id" form:"avatar_id"`
	RoleID      uuid.UUID `json:"role_id" xml:"role_id" form:"role_id"`
	Description string    `json:"description" xml:"description" form:"description"`
}

type UserUpdateProfileRequest struct {
	Name              string `json:"name" xml:"name" form:"name"`
	Email             string `json:"email" xml:"email" form:"email"`
	Phone             string `json:"phone" xml:"phone" form:"phone"`
	AvatarID            uuid.UUID `json:"avatar_id" xml:"avatar_id" form:"avatar_id"`
	Address           string `json:"address" xml:"address" form:"address"`
	Description      string `json:"description" xml:"description" form:"description"`
}

type UserCheckIdentifierRequest struct {
	Identifier    string `json:"identifier" validate:"required"`
}

type UserForgotPasswordRequest struct {
	Identifier    string `json:"identifier" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserLoginRequest struct {
	Identifier    string `json:"identifier" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserResetPasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type UserIsHardDeleteRequest struct {
	IsHardDelete string `json:"is_hard_delete" xml:"is_hard_delete" form:"is_hard_delete" validate:"required"`
	IDs []uuid.UUID `json:"ids" xml:"ids" form:"ids" validate:"required,dive,required"`
}

type UserRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" xml:"ids" form:"ids" validate:"required,dive,required"`
}

type ResponseGetUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone"`
	AvatarID    *uuid.UUID `json:"avatar_id"`
	Avatar 				 *Upload    `json:"avatar"`
	Role        *Role    `json:"role"`
	RoleID      uuid.UUID `json:"role_id"`
	Description string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at"`
}

type ResponseGerUserProfile struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone"`
	AvatarID    *uuid.UUID `json:"avatar_id"`
	Avatar 				 *Upload    `json:"avatar"`
	Role        *Role    `json:"role"`
	Description string    `json:"description"`
	RoleID      *uuid.UUID `json:"role_id"`
}