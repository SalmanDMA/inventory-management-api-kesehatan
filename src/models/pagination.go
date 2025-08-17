package models

type PaginationRequest struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Search string `query:"search"`
	Status string `query:"status"`
	RoleID string `query:"role_id"` // untuk paginated model user

	CategoryID string `query:"category_id"` // untuk paginated model item
	ItemID     string `query:"item_id"`     // untuk paginated model item history
	ChangeType string `query:"change_type"` // untuk paginated model item history

	AreaID         string `query:"area_id"`          // untuk paginated model facility
	FacilityTypeID string `query:"facility_type_id"` // untuk paginated model facility

	SupplierID    string `query:"supplier_id"`     // untuk paginated model purchase order
	POStatus      string `query:"po_status"`       // untuk paginated model purchase order
	PaymentStatus string `query:"payment_status"`  // untuk paginated model purchase order
	TermOfPayment string `query:"term_of_payment"` // untuk paginated model purchase order

	PurchaseOrderID string `query:"purchase_order_id"` // untuk paginated model payment
	PaymentType     string `query:"payment_type"`      // untuk paginated model payment

	SOStatus      string `query:"so_status"`       // untuk paginated model sales order
	FacilityID    string `query:"facility_id"`     // untuk paginated model sales order
	SalesPersonID string `query:"sales_person_id"` // untuk paginated model sales order
}

type PaginationResponse struct {
	CurrentPage  int   `json:"current_page"`
	PerPage      int   `json:"per_page"`
	TotalPages   int   `json:"total_pages"`
	TotalRecords int64 `json:"total_records"`
	HasNext      bool  `json:"has_next"`
	HasPrev      bool  `json:"has_prev"`
}

type UserPaginatedResponse struct {
	Data       []ResponseGetUser  `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type RolePaginatedResponse struct {
	Data       []ResponseGetRole  `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type CategoryPaginatedResponse struct {
	Data       []ResponseGetCategory `json:"data"`
	Pagination PaginationResponse    `json:"pagination"`
}

type ItemPaginatedResponse struct {
	Data       []ResponseGetItem  `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type ItemHistoryPaginatedResponse struct {
	Data       []ResponseGetItemHistory `json:"data"`
	Pagination PaginationResponse       `json:"pagination"`
}

type AreaPaginatedResponse struct {
	Data       []ResponseGetArea  `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type FacilityTypePaginatedResponse struct {
	Data       []ResponseGetFacilityType `json:"data"`
	Pagination PaginationResponse        `json:"pagination"`
}

type FacilityPaginatedResponse struct {
	Data       []ResponseGetFacility `json:"data"`
	Pagination PaginationResponse    `json:"pagination"`
}

type SalesPersonPaginatedResponse struct {
	Data       []ResponseGetSalesPerson `json:"data"`
	Pagination PaginationResponse       `json:"pagination"`
}

type SupplierPaginatedResponse struct {
	Data       []ResponseGetSupplier `json:"data"`
	Pagination PaginationResponse    `json:"pagination"`
}

type PurchaseOrderPaginatedResponse struct {
	Data       []ResponseGetPurchaseOrder `json:"data"`
	Pagination PaginationResponse         `json:"pagination"`
}

type PaymentPaginatedResponse struct {
	Data       []ResponseGetPayment `json:"data"`
	Pagination PaginationResponse   `json:"pagination"`
}

type SalesOrderPaginatedResponse struct {
	Data       []ResponseGetSalesOrder `json:"data"`
	Pagination PaginationResponse      `json:"pagination"`
}