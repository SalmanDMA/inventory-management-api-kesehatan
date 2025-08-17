package models

type MetricWithChange[T any] struct {
	Value     T      `json:"value"`
	Change    int    `json:"change"`
	ChangePct int    `json:"change_pct"`
	Trend     string `json:"trend"`
}

type DashboardSummary struct {
	TotalItems    MetricWithChange[int64]   `json:"total_items"`
	LowStockItems MetricWithChange[int64]   `json:"low_stock_items"`
	ActiveOrders  MetricWithChange[int64]   `json:"active_orders"`
	TotalValue    MetricWithChange[float64] `json:"total_value"`
}