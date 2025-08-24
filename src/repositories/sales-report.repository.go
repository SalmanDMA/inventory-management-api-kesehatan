package repositories

import (
	"fmt"
	"strings"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"gorm.io/gorm"
)

// ==============================
// Interface (transaction-aware)
// ==============================

type SalesReportRepository interface {
	GetSalesReportSummaryData(tx *gorm.DB, filters *models.PaginationRequest) (*SalesReportSummaryRaw, error)
	GetSalesReportDetails(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesOrder, int64, error)
	GetSalesTrendData(tx *gorm.DB, filters *models.PaginationRequest) (*SalesTrendRaw, error)
	GetStatusDistributionData(tx *gorm.DB, filters *models.PaginationRequest) (*StatusDistributionRaw, error)
	GetSalesByPersonData(tx *gorm.DB, filters *models.PaginationRequest) (*SalesByPersonRaw, error)
	GetPaymentComparisonData(tx *gorm.DB, filters *models.PaginationRequest) (*PaymentComparisonRaw, error)
	GetHeatmapData(tx *gorm.DB, filters *models.PaginationRequest) ([]struct {
		Date  string  `json:"date"`
		Value float64 `json:"value"`
	}, error)
	GetTopItems(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesReportTopItems, error)
	GetTopCustomers(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesReportTopCustomers, error)
	GetOverdueInvoices(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesReportOverdueInvoices, error)
	GetPerformanceData(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesReportPerformance, error)
}

// ==============================
// Implementation
// ==============================

type SalesReportRepositoryImpl struct {
	DB *gorm.DB
}

func NewSalesReportRepository(db *gorm.DB) *SalesReportRepositoryImpl {
	return &SalesReportRepositoryImpl{DB: db}
}

func (r *SalesReportRepositoryImpl) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}

// ==============================
// DTOs
// ==============================

type SalesReportSummaryRaw struct {
	TotalSales          int
	TotalPaid           int
	OutstandingPayment  int
	TotalOrders         int
	UniqueCustomers     int
	TopSalesPersonName  string
	TopSalesPersonValue int
	TopAreaName         string
	TopAreaValue        int
}

type SalesTrendRaw struct {
	Labels   []string
	Datasets []models.ChartDataset
}

type StatusDistributionRaw struct {
	Labels   []string
	Datasets []models.ChartDataset
}

type SalesByPersonRaw struct {
	Labels   []string
	Datasets []models.ChartDataset
}

type PaymentComparisonRaw struct {
	Labels   []string
	Datasets []models.ChartDataset
}

// -------------------------------------------------------
// Summary
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetSalesReportSummaryData(tx *gorm.DB, filters *models.PaginationRequest) (*SalesReportSummaryRaw, error) {
	db := r.useDB(tx)

	query := db.Model(&models.SalesOrder{}).
		Joins("LEFT JOIN sales_person ON sales_orders.sales_person_id = sales_person.id").
		Joins("LEFT JOIN customers ON sales_orders.customer_id = customers.id").
		Joins("LEFT JOIN areas ON customers.area_id = areas.id").
		Where("sales_orders.deleted_at IS NULL")

	query = r.applyFilters(query, db, filters)

	var result struct {
		TotalSales         int `gorm:"column:total_sales"`
		TotalPaid          int `gorm:"column:total_paid"`
		OutstandingPayment int `gorm:"column:outstanding_payment"`
		TotalOrders        int `gorm:"column:total_orders"`
		UniqueCustomers    int `gorm:"column:unique_customers"`
	}

	if err := query.Select(`
		COALESCE(SUM(sales_orders.total_amount), 0) as total_sales,
		COALESCE(SUM(sales_orders.paid_amount), 0) as total_paid,
		COALESCE(SUM(sales_orders.total_amount - sales_orders.paid_amount), 0) as outstanding_payment,
		COUNT(sales_orders.id) as total_orders,
		COUNT(DISTINCT sales_orders.customer_id) as unique_customers
	`).Scan(&result).Error; err != nil {
		return nil, err
	}

	var topSalesPerson struct {
		Name  string `gorm:"column:name"`
		Value int    `gorm:"column:value"`
	}
	qTopSP := db.Model(&models.SalesOrder{}).
		Joins("LEFT JOIN sales_person ON sales_orders.sales_person_id = sales_person.id").
		Where("sales_orders.deleted_at IS NULL")
	qTopSP = r.applyFilters(qTopSP, db, filters)
	if err := qTopSP.Select(`
		sales_person.name, COALESCE(SUM(sales_orders.total_amount), 0) as value
	`).Group("sales_person.id, sales_person.name").
		Order("value DESC").Limit(1).
		Scan(&topSalesPerson).Error; err != nil {
		return nil, err
	}

	var topArea struct {
		Name  string `gorm:"column:name"`
		Value int    `gorm:"column:value"`
	}
	qTopArea := db.Model(&models.SalesOrder{}).
		Joins("LEFT JOIN customers ON sales_orders.customer_id = customers.id").
		Joins("LEFT JOIN areas ON customers.area_id = areas.id").
		Where("sales_orders.deleted_at IS NULL")
	qTopArea = r.applyFilters(qTopArea, db, filters)
	if err := qTopArea.Select(`
		areas.name, COALESCE(SUM(sales_orders.total_amount), 0) as value
	`).Group("areas.id, areas.name").
		Order("value DESC").Limit(1).
		Scan(&topArea).Error; err != nil {
		return nil, err
	}

	return &SalesReportSummaryRaw{
		TotalSales:          result.TotalSales,
		TotalPaid:           result.TotalPaid,
		OutstandingPayment:  result.OutstandingPayment,
		TotalOrders:         result.TotalOrders,
		UniqueCustomers:     result.UniqueCustomers,
		TopSalesPersonName:  topSalesPerson.Name,
		TopSalesPersonValue: topSalesPerson.Value,
		TopAreaName:         topArea.Name,
		TopAreaValue:        topArea.Value,
	}, nil
}

// -------------------------------------------------------
// Details (pagination)
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetSalesReportDetails(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesOrder, int64, error) {
	db := r.useDB(tx)

	var (
		rows       []models.SalesOrder
		totalCount int64
	)

	base := db.Model(&models.SalesOrder{}).Where("sales_orders.deleted_at IS NULL")
	base = r.applyFilters(base, db, filters)

	if strings.TrimSpace(filters.Search) != "" {
		sp := "%" + strings.ToLower(filters.Search) + "%"
		subQuery := db.Model(&models.SalesOrder{}).
			Select("DISTINCT sales_orders.id").
			Joins("LEFT JOIN customers ON customers.id = sales_orders.customer_id").
			Joins("LEFT JOIN sales_person ON sales_person.id = sales_orders.sales_person_id").
			Where("sales_orders.deleted_at IS NULL").
			Where(`
				LOWER(sales_orders.so_number) LIKE ? OR
				LOWER(COALESCE(sales_orders.notes, '')) LIKE ? OR
				LOWER(COALESCE(customers.name, '')) LIKE ? OR
				LOWER(COALESCE(sales_person.name, '')) LIKE ?
			`, sp, sp, sp, sp)

		subQuery = r.applyFilters(subQuery, db, filters)
		base = base.Where("sales_orders.id IN (?)", subQuery)
	}

	if err := base.Count(&totalCount).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "sales_order")
	}
	if totalCount == 0 {
		return []models.SalesOrder{}, 0, nil
	}

	offset := (filters.Page - 1) * filters.Limit
	if err := base.
		Preload("SalesPerson").
		Preload("Customer").
		Preload("Customer.Area").
		Order("sales_orders.so_date DESC, sales_orders.created_at DESC").
		Offset(offset).
		Limit(filters.Limit).
		Find(&rows).Error; err != nil {
		return nil, 0, HandleDatabaseError(err, "sales_order")
	}

	return rows, totalCount, nil
}

// -------------------------------------------------------
// Trend
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetSalesTrendData(tx *gorm.DB, filters *models.PaginationRequest) (*SalesTrendRaw, error) {
	db := r.useDB(tx)

	q := db.Model(&models.SalesOrder{}).
		Where("sales_orders.deleted_at IS NULL")
	q = r.applyFilters(q, db, filters)

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	hasStart := !filters.StartDate.IsZero()
	hasEnd := !filters.EndDate.IsZero()

	var startRange, endRange time.Time
	period := strings.ToLower(strings.TrimSpace(filters.Period))

	if hasStart && hasEnd {
		startRange = filters.StartDate.In(loc)
		endRange = filters.EndDate.In(loc)
	} else {
		switch period {
		case "day":
			startRange = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -29)
			endRange = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, loc)
		case "week":
			wd := int(now.Weekday())
			if wd == 0 {
				wd = 7
			}
			thisWeekStart := time.Date(now.Year(), now.Month(), now.Day()-wd+1, 0, 0, 0, 0, loc)
			startRange = thisWeekStart.AddDate(0, 0, -7*11)
			endRange = thisWeekStart.AddDate(0, 0, 7).Add(-time.Nanosecond)
		case "year":
			startRange = time.Date(now.Year()-4, 1, 1, 0, 0, 0, 0, loc)
			endRange = time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, loc).Add(-time.Nanosecond)
		case "", "month":
			startRange = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, -11, 0)
			endRange = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, loc).Add(-time.Nanosecond)
		case "custom":
			return nil, fmt.Errorf("start_date and end_date are required for custom period")
		default:
			return nil, fmt.Errorf("invalid period: %s", filters.Period)
		}
	}

	if !(hasStart && hasEnd) {
		q = q.Where("sales_orders.so_date >= ? AND sales_orders.so_date <= ?", startRange, endRange)
	}

	const tz = "Asia/Jakarta"
	var groupExpr, labelExpr string
	switch period {
	case "week":
		groupExpr = fmt.Sprintf("DATE_TRUNC('week', sales_orders.so_date AT TIME ZONE '%s')", tz)
		labelExpr = fmt.Sprintf("TO_CHAR(DATE_TRUNC('week', sales_orders.so_date AT TIME ZONE '%s'), 'IYYY-IW')", tz)
	case "month", "":
		groupExpr = fmt.Sprintf("DATE_TRUNC('month', sales_orders.so_date AT TIME ZONE '%s')", tz)
		labelExpr = fmt.Sprintf("TO_CHAR(DATE_TRUNC('month', sales_orders.so_date AT TIME ZONE '%s'), 'YYYY-MM')", tz)
	case "year":
		groupExpr = fmt.Sprintf("DATE_TRUNC('year', sales_orders.so_date AT TIME ZONE '%s')", tz)
		labelExpr = fmt.Sprintf("TO_CHAR(DATE_TRUNC('year', sales_orders.so_date AT TIME ZONE '%s'), 'YYYY')", tz)
	default:
		groupExpr = fmt.Sprintf("DATE(sales_orders.so_date AT TIME ZONE '%s')", tz)
		labelExpr = fmt.Sprintf("TO_CHAR(DATE(sales_orders.so_date AT TIME ZONE '%s'), 'YYYY-MM-DD')", tz)
	}

	type row struct {
		Label  string `gorm:"column:label"`
		Amount int64  `gorm:"column:amount"`
	}
	var rows []row

	if err := q.Select(fmt.Sprintf(`
        %s AS grp,
        %s AS label,
        COALESCE(SUM(sales_orders.total_amount), 0) AS amount
    `, groupExpr, labelExpr)).
		Group("grp, label").
		Order("grp").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	labels := make([]string, 0, len(rows))
	data := make([]float64, 0, len(rows))
	for _, rr := range rows {
		labels = append(labels, rr.Label)
		data = append(data, float64(rr.Amount))
	}

	return &SalesTrendRaw{
		Labels: labels,
		Datasets: []models.ChartDataset{{
			Label:       "Sales Amount",
			Data:        data,
			BorderColor: "#3B82F6",
		}},
	}, nil
}

// -------------------------------------------------------
// Status Distribution
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetStatusDistributionData(tx *gorm.DB, filters *models.PaginationRequest) (*StatusDistributionRaw, error) {
	db := r.useDB(tx)

	var results []struct {
		Status string `gorm:"column:status"`
		Count  int    `gorm:"column:count"`
	}

	query := db.Model(&models.SalesOrder{}).
		Where("sales_orders.deleted_at IS NULL")
	query = r.applyFilters(query, db, filters)

	if err := query.Select(`
		sales_orders.so_status as status, COUNT(*) as count
	`).Group("sales_orders.so_status").Scan(&results).Error; err != nil {
		return nil, err
	}

	labels := make([]string, 0, len(results))
	data := make([]float64, 0, len(results))
	colorsBase := []string{"#EF4444", "#F59E0B", "#10B981", "#3B82F6", "#8B5CF6", "#EC4899", "#14B8A6", "#F97316", "#84CC16", "#6366F1"}
	colors := make([]string, 0, len(results))
	for i := range results {
		labels = append(labels, results[i].Status)
		data = append(data, float64(results[i].Count))
		colors = append(colors, colorsBase[i%len(colorsBase)])
	}

	return &StatusDistributionRaw{
		Labels: labels,
		Datasets: []models.ChartDataset{{
			Data:            data,
			BackgroundColor: colors,
		}},
	}, nil
}

// -------------------------------------------------------
// Sales By Person
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetSalesByPersonData(tx *gorm.DB, filters *models.PaginationRequest) (*SalesByPersonRaw, error) {
	db := r.useDB(tx)

	var results []struct {
		Name   string `gorm:"column:name"`
		Amount int    `gorm:"column:amount"`
	}

	query := db.Model(&models.SalesOrder{}).
		Joins("LEFT JOIN sales_person ON sales_orders.sales_person_id = sales_person.id").
		Where("sales_orders.deleted_at IS NULL")
	query = r.applyFilters(query, db, filters)

	if err := query.Select(`
		sales_person.name, COALESCE(SUM(sales_orders.total_amount), 0) as amount
	`).Group("sales_person.id, sales_person.name").
		Order("amount DESC").Limit(10).
		Scan(&results).Error; err != nil {
		return nil, err
	}

	labels := make([]string, 0, len(results))
	data := make([]float64, 0, len(results))
	colorsBase := []string{"#3B82F6", "#10B981", "#F59E0B", "#EF4444", "#8B5CF6", "#EC4899", "#14B8A6", "#F97316", "#84CC16", "#6366F1"}
	bg := make([]string, 0, len(results))
	for i := range results {
		labels = append(labels, results[i].Name)
		data = append(data, float64(results[i].Amount))
		bg = append(bg, colorsBase[i%len(colorsBase)])
	}

	return &SalesByPersonRaw{
		Labels: labels,
		Datasets: []models.ChartDataset{{
			Label:           "Sales Amount",
			Data:            data,
			BackgroundColor: bg,
		}},
	}, nil
}

// -------------------------------------------------------
// Payment Comparison
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetPaymentComparisonData(tx *gorm.DB, filters *models.PaginationRequest) (*PaymentComparisonRaw, error) {
	db := r.useDB(tx)

	var result struct {
		Paid   int `gorm:"column:paid"`
		Unpaid int `gorm:"column:unpaid"`
		DP     int `gorm:"column:dp"`
	}

	query := db.Model(&models.SalesOrder{}).
		Where("sales_orders.deleted_at IS NULL")
	query = r.applyFilters(query, db, filters)

	if err := query.Select(`
		COALESCE(SUM(CASE WHEN LOWER(sales_orders.payment_status) = 'paid' THEN sales_orders.total_amount ELSE 0 END), 0) as paid,
		COALESCE(SUM(CASE WHEN LOWER(sales_orders.payment_status) = 'unpaid' THEN sales_orders.total_amount ELSE 0 END), 0) as unpaid,
		COALESCE(SUM(sales_orders.dp_amount), 0) as dp
	`).Scan(&result).Error; err != nil {
		return nil, err
	}

	return &PaymentComparisonRaw{
		Labels: []string{"Paid", "Unpaid", "DP"},
		Datasets: []models.ChartDataset{{
			Label:           "Amount",
			Data:            []float64{float64(result.Paid), float64(result.Unpaid), float64(result.DP)},
			BackgroundColor: []string{"#10B981", "#EF4444", "#F59E0B"},
		}},
	}, nil
}

// -------------------------------------------------------
// Heatmap
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetHeatmapData(tx *gorm.DB, filters *models.PaginationRequest) ([]struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}, error) {
	db := r.useDB(tx)

	var results []struct {
		Date  string  `gorm:"column:date"`
		Value float64 `gorm:"column:value"`
	}

	query := db.Model(&models.SalesOrder{}).
		Where("sales_orders.deleted_at IS NULL")
	query = r.applyFilters(query, db, filters)

	if err := query.Select(`
		TO_CHAR(DATE(sales_orders.so_date), 'YYYY-MM-DD') as date,
		COALESCE(SUM(sales_orders.total_amount), 0) as value
	`).Group("DATE(sales_orders.so_date)").
		Order("DATE(sales_orders.so_date)").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	out := make([]struct {
		Date  string  `json:"date"`
		Value float64 `json:"value"`
	}, 0, len(results))
	for _, r := range results {
		out = append(out, struct {
			Date  string  `json:"date"`
			Value float64 `json:"value"`
		}{Date: r.Date, Value: r.Value})
	}
	return out, nil
}

// -------------------------------------------------------
// Top Items
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetTopItems(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesReportTopItems, error) {
	db := r.useDB(tx)

	var results []models.SalesReportTopItems

	query := db.Model(&models.SalesOrderItem{}).
		Joins("LEFT JOIN sales_orders ON sales_order_items.sales_order_id = sales_orders.id").
		Joins("LEFT JOIN items ON sales_order_items.item_id = items.id").
		Where("sales_order_items.deleted_at IS NULL AND sales_orders.deleted_at IS NULL")

	query = r.applyFilters(query, db, filters)

	if err := query.Select(`
		items.id as item_id,
		items.name as item_name,
		items.code as item_code,
		COALESCE(SUM(sales_order_items.quantity), 0) as total_quantity,
		COALESCE(SUM(sales_order_items.total_price), 0) as total_revenue,
		COUNT(DISTINCT sales_orders.id) as order_count
	`).Group("items.id, items.name, items.code").
		Order("total_revenue DESC").Limit(10).
		Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// -------------------------------------------------------
// Top Customers
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetTopCustomers(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesReportTopCustomers, error) {
	db := r.useDB(tx)

	var results []models.SalesReportTopCustomers

	query := db.Model(&models.SalesOrder{}).
		Joins("LEFT JOIN customers ON sales_orders.customer_id = customers.id").
		Joins("LEFT JOIN areas ON customers.area_id = areas.id").
		Where("sales_orders.deleted_at IS NULL")

	query = r.applyFilters(query, db, filters)

	if err := query.Select(`
		customers.id as customer_id,
		customers.name as customer_name,
		areas.name as area_name,
		COUNT(sales_orders.id) as total_orders,
		COALESCE(SUM(sales_orders.total_amount), 0) as total_amount,
		MAX(sales_orders.so_date) as last_order_date
	`).Group("customers.id, customers.name, areas.name").
		Order("total_amount DESC").Limit(10).
		Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// -------------------------------------------------------
// Overdue Invoices
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetOverdueInvoices(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesReportOverdueInvoices, error) {
	db := r.useDB(tx)

	var results []models.SalesReportOverdueInvoices

	query := db.Model(&models.SalesOrder{}).
		Where("sales_orders.deleted_at IS NULL").
		Where("sales_orders.due_date < ?", time.Now()).
		Where("LOWER(sales_orders.payment_status) <> 'paid'")

	query = r.applyFilters(query, db, filters)

	if err := query.Select(`
		sales_orders.id,
		sales_orders.so_number,
		(SELECT f.name FROM customers f WHERE f.id = sales_orders.customer_id) as customer_name,
		(SELECT sp.name FROM sales_person sp WHERE sp.id = sales_orders.sales_person_id) as sales_person_name,
		sales_orders.total_amount,
		sales_orders.paid_amount,
		(sales_orders.total_amount - sales_orders.paid_amount) as unpaid_amount,
		sales_orders.due_date,
		CAST(EXTRACT(DAY FROM (NOW() - sales_orders.due_date)) AS INT) as days_overdue
	`).Order("days_overdue DESC").Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// -------------------------------------------------------
// Performance
// -------------------------------------------------------

func (r *SalesReportRepositoryImpl) GetPerformanceData(tx *gorm.DB, filters *models.PaginationRequest) ([]models.SalesReportPerformance, error) {
	db := r.useDB(tx)

	var results []models.SalesReportPerformance

	query := db.Model(&models.SalesOrder{}).
		Joins("LEFT JOIN sales_person ON sales_orders.sales_person_id = sales_person.id").
		Joins("LEFT JOIN customers ON sales_orders.customer_id = customers.id").
		Joins("LEFT JOIN areas ON customers.area_id = areas.id").
		Where("sales_orders.deleted_at IS NULL")

	query = r.applyFilters(query, db, filters)

	if err := query.Select(`
		sales_person.id as sales_person_id,
		sales_person.name as sales_person_name,
		STRING_AGG(DISTINCT areas.name, ', ') as area_names,
		COUNT(sales_orders.id) as total_orders,
		COALESCE(SUM(sales_orders.total_amount), 0) as total_revenue,
		COALESCE(SUM(sales_orders.paid_amount), 0) as paid_revenue,
		COALESCE(SUM(sales_orders.total_amount - sales_orders.paid_amount), 0) as unpaid_revenue,
		CASE WHEN COUNT(sales_orders.id) = 0 THEN 0
		     ELSE (COUNT(CASE WHEN LOWER(sales_orders.payment_status) = 'paid' THEN 1 END) * 100.0 / COUNT(sales_orders.id))
		END as conversion_rate,
		CASE WHEN COUNT(sales_orders.id) = 0 THEN 0
		     ELSE (COALESCE(SUM(sales_orders.total_amount), 0) / COUNT(sales_orders.id))
		END as avg_order_value,
		ROW_NUMBER() OVER (ORDER BY COALESCE(SUM(sales_orders.total_amount), 0) DESC) as rank
	`).Group("sales_person.id, sales_person.name").
		Order("total_revenue DESC").Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// -------------------------------------------------------
// Filters (pakai DB dari tx agar konsisten)
// -------------------------------------------------------

func isEmpty(v string) bool {
	s := strings.TrimSpace(strings.ToLower(v))
	return s == "" || s == "all" || s == "undefined" || s == "null"
}

func (r *SalesReportRepositoryImpl) applyFilters(q *gorm.DB, db *gorm.DB, f *models.PaginationRequest) *gorm.DB {
	if !f.StartDate.IsZero() {
		q = q.Where("sales_orders.so_date >= ?", f.StartDate)
	}
	if !f.EndDate.IsZero() {
		q = q.Where("sales_orders.so_date <= ?", f.EndDate)
	}
	if !isEmpty(f.SalesPersonID) {
		q = q.Where("sales_orders.sales_person_id = ?", f.SalesPersonID)
	}
	if !isEmpty(f.CustomerID) {
		q = q.Where("sales_orders.customer_id = ?", f.CustomerID)
	}
	if !isEmpty(f.SOStatus) {
		q = q.Where("LOWER(sales_orders.so_status) = LOWER(?)", f.SOStatus)
	}
	if !isEmpty(f.PaymentStatus) {
		q = q.Where("LOWER(sales_orders.payment_status) = LOWER(?)", f.PaymentStatus)
	}
	if !isEmpty(f.AreaID) {
		sub := db.Table("customers").
			Select("id").
			Where("area_id = ? AND deleted_at IS NULL", f.AreaID)
		q = q.Where("sales_orders.customer_id IN (?)", sub)
	}
	return q
}
