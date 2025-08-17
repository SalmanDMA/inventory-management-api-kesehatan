package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payment struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	OrderType       string         `gorm:"not null" json:"order_type"` // PO, SO
	PurchaseOrderID *uuid.UUID     `gorm:"type:uuid" json:"purchase_order_id,omitempty"`
	SalesOrderID    *uuid.UUID     `gorm:"type:uuid" json:"sales_order_id,omitempty"`

	PaymentType     string         `gorm:"not null" json:"payment_type"` // DP, Full, Installment
	Amount          int            `gorm:"not null" json:"amount"`
	PaymentDate     time.Time      `gorm:"not null" json:"payment_date"`
	PaymentMethod   string         `json:"payment_method"` // Cash, Transfer, etc.
	ReferenceNumber string         `json:"reference_number"`
	Notes           string         `json:"notes"`
	InvoiceID       *uuid.UUID     `gorm:"type:uuid" json:"invoice_id,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	PurchaseOrder *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order,omitempty"`
	SalesOrder    *SalesOrder    `gorm:"foreignKey:SalesOrderID" json:"sales_order,omitempty"`
	Invoice       *Upload       `gorm:"foreignKey:InvoiceID;constraint:onUpdate:CASCADE,onDelete:SET NULL;" json:"invoice,omitempty"`
}

type ResponseGetPayment struct {
	ID               uuid.UUID      `json:"id"`
	OrderType        string         `json:"order_type"`
	PurchaseOrderID  *uuid.UUID      `json:"purchase_order_id"`
	SalesOrderID     *uuid.UUID     `json:"sales_order_id"`
	
	PaymentType      string         `json:"payment_type"`
	Amount           int            `json:"amount"`
	PaymentDate      time.Time      `json:"payment_date"`
	PaymentMethod    string         `json:"payment_method"`
	ReferenceNumber  string         `json:"reference_number"`
	Notes            string         `json:"notes"`
	InvoiceID        *uuid.UUID     `json:"invoice_id,omitempty"`
	
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty"`

	PurchaseOrder    *PurchaseOrder `json:"purchase_order,omitempty"`
	SalesOrder       *SalesOrder    `json:"sales_order,omitempty"`
	Invoice          *Upload        `json:"invoice,omitempty"`
}


type PaymentCreateRequest struct {
	OrderType       string    `json:"order_type" form:"order_type" validate:"required,oneof=PO SO"`
	PurchaseOrderID uuid.UUID `json:"purchase_order_id" form:"purchase_order_id"`
	SalesOrderID    uuid.UUID `json:"sales_order_id" form:"sales_order_id"`
	PaymentType     string    `json:"payment_type" form:"payment_type" validate:"required,oneof=DP Full Installment"`
	Amount          int       `json:"amount" form:"amount" validate:"required,min=1"`
	PaymentDate     time.Time `json:"payment_date" form:"payment_date" validate:"required"`
	PaymentMethod   string    `json:"payment_method" form:"payment_method"`
	ReferenceNumber string    `json:"reference_number" form:"reference_number"`
	Notes           string    `json:"notes" form:"notes"`
}