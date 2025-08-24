package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseOrder struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	PONumber          string         `gorm:"uniqueIndex;not null" json:"po_number"`
	SupplierID        uuid.UUID      `gorm:"type:uuid;not null" json:"supplier_id"`
	PODate            time.Time      `gorm:"not null" json:"po_date"`
	EstimatedArrival  *time.Time     `json:"estimated_arrival"`
	TermOfPayment     string         `gorm:"not null" json:"term_of_payment"` // Full, DP, Tempo
	POStatus          string         `gorm:"not null;default:'Draft'" json:"po_status"` // Draft, Ordered, Received, Partial, Returned, Closed
	PaymentStatus     string         `gorm:"not null;default:'Unpaid'" json:"payment_status"` // Unpaid, Partial, Paid
	TotalAmount       int            `gorm:"not null" json:"total_amount"`
	PaidAmount        int            `gorm:"default:0" json:"paid_amount"`
	DPAmount          int            `gorm:"default:0" json:"dp_amount"`
	DueDate           *time.Time     `json:"due_date"`
	Notes             string         `json:"notes"`
	Tax 														float32            `gorm:"default:0" json:"tax"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	Supplier          Supplier            `gorm:"foreignKey:SupplierID" json:"supplier"`
	PurchaseOrderItems []PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order_items,omitempty"`
	Payments          []Payment           `gorm:"foreignKey:PurchaseOrderID" json:"payments,omitempty"`
}

type PurchaseOrderItem struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	PurchaseOrderID  uuid.UUID      `gorm:"type:uuid;not null" json:"purchase_order_id"`
	ItemID           uuid.UUID      `gorm:"type:uuid;not null" json:"item_id"`
	UoMID      uuid.UUID      						`gorm:"column:uom_id;type:uuid;not null" json:"uom_id"`
	Quantity         int            `gorm:"not null" json:"quantity"`           
	UnitPrice        int            `gorm:"not null" json:"unit_price"`         
	TotalPrice       int            `gorm:"not null" json:"total_price"`
	ReceivedQuantity int            `gorm:"default:0" json:"received_quantity"`
	ReturnedQuantity int            `gorm:"default:0" json:"returned_quantity"`
	Status           string         `gorm:"not null;default:'Ordered'" json:"status"` // Ordered, Received, Returned, Partial

	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	PurchaseOrder    PurchaseOrder  `gorm:"foreignKey:PurchaseOrderID" json:"purchase_order"`
	Item             Item           `gorm:"foreignKey:ItemID" json:"item"`
	UoM              UoM            `gorm:"foreignKey:UoMID" json:"uom"`
}

type ResponseGetPurchaseOrder struct {
	ID                uuid.UUID               `json:"id"`
	PONumber          string                  `json:"po_number"`
	SupplierID        uuid.UUID               `json:"supplier_id"`
	PODate            time.Time               `json:"po_date"`
	EstimatedArrival  *time.Time              `json:"estimated_arrival"`
	TermOfPayment     string                  `json:"term_of_payment"`
	POStatus          string                  `json:"po_status"`
	PaymentStatus     string                  `json:"payment_status"`
	TotalAmount       int                     `json:"total_amount"`
	PaidAmount        int                     `json:"paid_amount"`
	DPAmount          int                     `json:"dp_amount"`
	DueDate           *time.Time              `json:"due_date"`
	Notes             string                  `json:"notes"`
 Tax 													float32                     	`json:"tax"`

	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
	DeletedAt         gorm.DeletedAt          `json:"deleted_at,omitempty"`
	
	Supplier          Supplier     `json:"supplier"`
	PurchaseOrderItems []PurchaseOrderItem `json:"purchase_order_items,omitempty"`
	Payments          []Payment    `json:"payments,omitempty"`
}

type ResponseGetPurchaseOrderItem struct {
	ID               uuid.UUID      `json:"id"`
	PurchaseOrderID  uuid.UUID      `json:"purchase_order_id"`
	ItemID           uuid.UUID      `json:"item_id"`
	Quantity         int            `json:"quantity"`
	UnitPrice        int            `json:"unit_price"`
	TotalPrice       int            `json:"total_price"`
	ReceivedQuantity int            `json:"received_quantity"`
	ReturnedQuantity int            `json:"returned_quantity"`
	Status           string         `json:"status"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty"`
	Item             Item `json:"item"`
}

type PurchaseOrderItemRequest struct {
	ItemID    uuid.UUID `json:"item_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
	UnitPrice int       `json:"unit_price" validate:"required,min=0"`
}

type PurchaseOrderCreateRequest struct {
	SupplierID       uuid.UUID                  `json:"supplier_id" validate:"required"`
	PODate           time.Time                  `json:"po_date" validate:"required"`
	POStatus         string                     `json:"po_status" validate:"required"`
	EstimatedArrival *time.Time                 `json:"estimated_arrival"`
	TermOfPayment    string                     `json:"term_of_payment" validate:"required,oneof=Full DP Tempo"`
	DPAmount         int                        `json:"dp_amount"`
	DueDate          *time.Time                 `json:"due_date"`
	Notes            string                     `json:"notes"`
	Tax 													float32                     `json:"tax"`
	Items            []PurchaseOrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type PurchaseOrderUpdateRequest struct {
	SupplierID       uuid.UUID                  `json:"supplier_id"`
	PODate           time.Time                  `json:"po_date"`
	EstimatedArrival *time.Time                 `json:"estimated_arrival"`
	TermOfPayment    string                     `json:"term_of_payment" validate:"omitempty,oneof=Full DP Tempo"`
	DPAmount         int                        `json:"dp_amount"`
	DueDate          *time.Time                 `json:"due_date"`
	Notes            string                     `json:"notes"`
	Tax 													float32                        `json:"tax"`
	Items            []PurchaseOrderItemRequest `json:"items" validate:"omitempty,min=1,dive"`
}

type PurchaseOrderStatusUpdateRequest struct {
	POStatus      string `json:"po_status" validate:"required,oneof=Draft Ordered Received Returned Closed"`
	PaymentStatus string `json:"payment_status" validate:"omitempty,oneof=Unpaid Partial Paid"`
}

type ReceiveItemsRequest struct {
	Items []struct {
		PurchaseOrderItemID uuid.UUID `json:"purchase_order_item_id" validate:"required"`
		ReceivedQuantity    int       `json:"received_quantity" validate:"min=0"`
		ReturnedQuantity    int       `json:"returned_quantity" validate:"min=0"`
	} `json:"items" validate:"required,min=1,dive"`
}

type PurchaseOrderIsHardDeleteRequest struct {
	IsHardDelete string      `json:"is_hard_delete" validate:"required"`
	IDs          []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

type PurchaseOrderRestoreRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,required"`
}

