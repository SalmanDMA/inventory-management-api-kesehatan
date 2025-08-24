package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesOrder struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SONumber   string    `gorm:"uniqueIndex;not null" json:"so_number"`
	SalesPersonID uuid.UUID `gorm:"type:uuid;not null" json:"sales_person_id"`
	CustomerID    uuid.UUID `gorm:"type:uuid;not null" json:"customer_id"`
	SODate     time.Time `gorm:"not null" json:"so_date"`
	EstimatedArrival  *time.Time     `json:"estimated_arrival"`
	TermOfPayment     string         `gorm:"not null" json:"term_of_payment"` // Full, DP, Tempo
	SOStatus string `gorm:"not null;default:'Draft'" json:"so_status"` 	// Draft, Confirmed, Shipped, Delivered, Closed
	PaymentStatus string `gorm:"not null;default:'Unpaid'" json:"payment_status"`	// Unpaid, Partial, Paid
	TotalAmount       int            `gorm:"not null" json:"total_amount"`
	PaidAmount        int            `gorm:"default:0" json:"paid_amount"`
	DPAmount          int            `gorm:"default:0" json:"dp_amount"`
	DueDate           *time.Time     `json:"due_date"`
	Notes       string `json:"notes"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	SalesPerson SalesPerson      `gorm:"foreignKey:SalesPersonID" json:"sales_person"`
	Customer    Customer         `gorm:"foreignKey:CustomerID" json:"customer"`
	SalesOrderItems       []SalesOrderItem `gorm:"foreignKey:SalesOrderID" json:"sales_order_items,omitempty"`
	Payments    []Payment        `gorm:"foreignKey:SalesOrderID" json:"payments"`
}

type SalesOrderItem struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SalesOrderID uuid.UUID `gorm:"type:uuid;not null" json:"sales_order_id"`
	ItemID       uuid.UUID `gorm:"type:uuid;not null" json:"item_id"`
	UoMID      uuid.UUID      `gorm:"column:uom_id;type:uuid;not null" json:"uom_id"`
	Quantity     int       `gorm:"not null" json:"quantity"`
	UnitPrice    int       `gorm:"not null" json:"unit_price"`
	TotalPrice   int       `gorm:"not null" json:"total_price"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	SalesOrder SalesOrder `gorm:"foreignKey:SalesOrderID" json:"sales_order"`
	Item       Item       `gorm:"foreignKey:ItemID" json:"item"`
	UoM        UoM        `gorm:"foreignKey:UoMID" json:"uom"`
}

type ResponseGetSalesOrder struct {
	ID            uuid.UUID             `json:"id"`
	SONumber      string                `json:"so_number"`
	SalesPersonID uuid.UUID             `json:"sales_person_id"`
	CustomerID    uuid.UUID             `json:"customer_id"`
	SODate        time.Time             `json:"so_date"`
	EstimatedArrival *time.Time         `json:"estimated_arrival"`
	TermOfPayment string                `json:"term_of_payment"`
	SOStatus      string                `json:"so_status"`
	PaymentStatus string                `json:"payment_status"`
	TotalAmount   int                   `json:"total_amount"`
	PaidAmount    int                   `json:"paid_amount"`
	DPAmount      int                   `json:"dp_amount"`
	DueDate       *time.Time            `json:"due_date"`
	Notes         string                `json:"notes"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	DeletedAt     gorm.DeletedAt        `json:"deleted_at,omitempty"`

	SalesPerson   SalesPerson `json:"sales_person"`
	Customer      Customer    `json:"customer"`
	SalesOrderItems         []SalesOrderItem `json:"sales_order_items,omitempty"`
	Payments      []Payment   `json:"payments,omitempty"`
}

type ResponseGetSalesOrderItem struct {
	ID              uuid.UUID      `json:"id"`
	SalesOrderID    uuid.UUID      `json:"sales_order_id"`
	ItemID          uuid.UUID      `json:"item_id"`
	Quantity        int            `json:"quantity"`
	UnitPrice       int            `json:"unit_price"`
	TotalPrice      int            `json:"total_price"`

	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at,omitempty"`

	Item Item `json:"item"`
}

type SalesOrderItemRequest struct {
	ItemID    uuid.UUID `json:"item_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
	UnitPrice int       `json:"unit_price" validate:"required,min=0"`
}

type SalesOrderCreateRequest struct {
	SalesPersonID    uuid.UUID               `json:"sales_person_id" validate:"required"`
	CustomerID       uuid.UUID               `json:"customer_id" validate:"required"`
	SODate           time.Time               `json:"so_date" validate:"required"`
	SOStatus         string                  `json:"so_status" validate:"required"`
	EstimatedArrival *time.Time              `json:"estimated_arrival"`
	TermOfPayment    string                  `json:"term_of_payment" validate:"required,oneof=Full DP Tempo"`
	DPAmount         int                     `json:"dp_amount"`
	DueDate          *time.Time              `json:"due_date"`
	Notes            string                  `json:"notes"`
	Items            []SalesOrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type SalesOrderUpdateRequest struct {
	SalesPersonID    uuid.UUID               `json:"sales_person_id"`
	CustomerID       uuid.UUID               `json:"customer_id"`
	SODate           time.Time               `json:"so_date"`
	EstimatedArrival *time.Time              `json:"estimated_arrival"`
	TermOfPayment    string                  `json:"term_of_payment" validate:"omitempty,oneof=Full DP Tempo"`
	DPAmount         int                     `json:"dp_amount"`
	DueDate          *time.Time              `json:"due_date"`
	Notes            string                  `json:"notes"`
	Items            []SalesOrderItemRequest `json:"items" validate:"omitempty,min=1,dive"`
}

type SalesOrderStatusUpdateRequest struct {
	SOStatus      string `json:"so_status" validate:"required,oneof=Draft Confirmed Shipped Delivered Closed"`
	PaymentStatus string `json:"payment_status" validate:"omitempty,oneof=Unpaid Partial Paid"`
}

type SalesOrderIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type SalesOrderRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}