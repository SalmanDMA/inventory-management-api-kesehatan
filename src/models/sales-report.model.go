package models

import (
	"time"

	"github.com/google/uuid"
)

type SalesReportSummaryItem struct {
	Value     float64 `json:"value"`
	ChangePct float64 `json:"change_pct"`
	Trend     string  `json:"trend"` // up, down, same
}

type SalesReportTopItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	ChangePct float64 `json:"change_pct"`
	Trend     string  `json:"trend"`
}

type ChartDataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor interface{} `json:"background_color,omitempty"`
	BorderColor     string    `json:"border_color,omitempty"`
}


type SalesReportTopItems struct {
	ItemID        uuid.UUID `json:"item_id"`
	ItemName      string    `json:"item_name"`
	ItemCode      string    `json:"item_code"`
	TotalQuantity int       `json:"total_quantity"`
	TotalRevenue  int       `json:"total_revenue"`
	OrderCount    int       `json:"order_count"`
}

type SalesReportTopCustomers struct {
	CustomerID    uuid.UUID `json:"customer_id"`
	CustomerName  string    `json:"customer_name"`
	AreaName      string    `json:"area_name"`
	TotalOrders   int       `json:"total_orders"`
	TotalAmount   int       `json:"total_amount"`
	LastOrderDate time.Time `json:"last_order_date"`
}

type SalesReportOverdueInvoices struct {
	ID               uuid.UUID `json:"id"`
	SONumber         string    `json:"so_number"`
	CustomerName     string    `json:"customer_name"`
	SalesPersonName  string    `json:"sales_person_name"`
	TotalAmount      int       `json:"total_amount"`
	PaidAmount       int       `json:"paid_amount"`
	UnpaidAmount     int       `json:"unpaid_amount"`
	DueDate          time.Time `json:"due_date"`
	DaysOverdue      int       `json:"days_overdue"`
}

type SalesReportPerformance struct {
	SalesPersonID   uuid.UUID `json:"sales_person_id"`
	SalesPersonName string    `json:"sales_person_name"`
	AreaNames       string  `json:"area_names"`
	TotalOrders     int       `json:"total_orders"`
	TotalRevenue    int       `json:"total_revenue"`
	PaidRevenue     int       `json:"paid_revenue"`
	UnpaidRevenue   int       `json:"unpaid_revenue"`
	ConversionRate  float64   `json:"conversion_rate"`
	AvgOrderValue   float64   `json:"avg_order_value"`
	Rank            int       `json:"rank"`
}

type SalesReportSummary struct {
	TotalSales        SalesReportSummaryItem `json:"total_sales"`
	TotalPaid         SalesReportSummaryItem `json:"total_paid"`
	OutstandingPayment SalesReportSummaryItem `json:"outstanding_payment"`
	TotalOrders       SalesReportSummaryItem `json:"total_orders"`
	UniqueCustomers   SalesReportSummaryItem `json:"unique_customers"`
	TopSalesPerson    SalesReportTopItem     `json:"top_sales_person"`
	TopArea           SalesReportTopItem     `json:"top_area"`
}

type SalesReportChartData struct {
	SalesTrend struct {
		Labels   []string       `json:"labels"`
		Datasets []ChartDataset `json:"datasets"`
	} `json:"sales_trend"`
	StatusDistribution struct {
		Labels   []string       `json:"labels"`
		Datasets []ChartDataset `json:"datasets"`
	} `json:"status_distribution"`
	SalesByPerson struct {
		Labels   []string       `json:"labels"`
		Datasets []ChartDataset `json:"datasets"`
	} `json:"sales_by_person"`
	PaymentComparison struct {
		Labels   []string       `json:"labels"`
		Datasets []ChartDataset `json:"datasets"`
	} `json:"payment_comparison"`
	HeatmapData []struct {
		Date  string  `json:"date"`
		Value float64 `json:"value"`
	} `json:"heatmap_data"`
}

type SalesReportDetailItem struct {
	ID            uuid.UUID `json:"id"`
	SONumber      string    `json:"so_number"`
	SODate        time.Time `json:"so_date"`
	SalesPerson   SalesPerson `json:"sales_person"`
	Customer      Customer    `json:"customer"`
	SOStatus      string    `json:"so_status"`
	PaymentStatus string    `json:"payment_status"`
	TotalAmount   int       `json:"total_amount"`
	PaidAmount    int       `json:"paid_amount"`
	UnpaidAmount  int       `json:"unpaid_amount"`
	DueDate       *time.Time `json:"due_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SalesReportInsights struct {
	TopItems        []SalesReportTopItems        `json:"top_items"`
	TopCustomers    []SalesReportTopCustomers    `json:"top_customers"`
	OverdueInvoices []SalesReportOverdueInvoices `json:"overdue_invoices"`
	Performance     []SalesReportPerformance     `json:"performance"`
}

// type SalesReportData struct {
// 	Summary  SalesReportSummary    `json:"summary"`
// 	Charts   SalesReportChartData  `json:"charts"`
// 	Details  struct {
// 		Data       []SalesReportDetailItem `json:"data"`
// 		Pagination PaginationResponse      `json:"pagination"`
// 	} `json:"details"`
// 	Insights SalesReportInsights `json:"insights"`
// }

// type SalesReportExportRequest struct {
// 	Filters         PaginationRequest `json:"filters" validate:"required"`
// 	Format          string             `json:"format" validate:"required,oneof=excel csv pdf"`
// 	IncludeSummary  bool               `json:"include_summary"`
// 	IncludeDetails  bool               `json:"include_details"`
// 	IncludeInsights bool               `json:"include_insights"`
// }

// type SalesReportExportResponse struct {
// 	Success  bool   `json:"success"`
// 	Message  string `json:"message"`
// 	FileURL  string `json:"file_url,omitempty"`
// 	FileName string `json:"file_name,omitempty"`
// }

