package models

type PaginationRequest struct {
	Page       int    `query:"page"`
	Limit      int    `query:"limit"`
	Search     string `query:"search"`
	Status     string `query:"status"`
	RoleID     string `query:"role_id"`     // untuk paginated model user
	CategoryID string `query:"category_id"` // untuk paginated model item
	ItemID     string `query:"item_id"`     // untuk paginated model item history
	ChangeType string `query:"change_type"` // untuk paginated model item history
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