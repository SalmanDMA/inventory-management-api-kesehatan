package services

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/helpers/documents"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/google/uuid"
)

type SalesReportService struct {
	SalesReportRepo repositories.SalesReportRepository
	SalesPersonRepo repositories.SalesPersonRepository
	SalesOrderRepo  repositories.SalesOrderRepository
}

func NewSalesReportService(
	srRepo repositories.SalesReportRepository,
	spRepo repositories.SalesPersonRepository,
	soRepo repositories.SalesOrderRepository,
) *SalesReportService {
	return &SalesReportService{
		SalesReportRepo: srRepo,
		SalesPersonRepo: spRepo,
		SalesOrderRepo:  soRepo,
	}
}

func (s *SalesReportService) GetSalesReportSummary(filters *models.PaginationRequest) (*models.SalesReportSummary, error) {
	if err := s.setDefaultDateRange(filters); err != nil {
		return nil, err
	}
	return s.getSummaryData(filters)
}

func (s *SalesReportService) GetSalesReportCharts(filters *models.PaginationRequest) (*models.SalesReportChartData, error) {
	return s.getChartsData(filters)
}

func (s *SalesReportService) GetSalesReportDetails(filters *models.PaginationRequest) (*models.SalesReportDetailItemPaginatedResponse, error) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 10
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	if err := s.setDefaultDateRange(filters); err != nil {
		return nil, err
	}

	orders, totalCount, err := s.SalesReportRepo.GetSalesReportDetails(nil, filters)
	if err != nil {
		return nil, err
	}

	detailsResponse := make([]models.SalesReportDetailItem, 0, len(orders))
	for _, so := range orders {
		unpaid := so.TotalAmount - so.PaidAmount
		if unpaid < 0 {
			unpaid = 0
		}
		detailsResponse = append(detailsResponse, models.SalesReportDetailItem{
			ID:            so.ID,
			SONumber:      so.SONumber,
			SODate:        so.SODate,
			SalesPerson:   so.SalesPerson,
			Customer:      so.Customer,
			SOStatus:      so.SOStatus,
			PaymentStatus: so.PaymentStatus,
			TotalAmount:   so.TotalAmount,
			PaidAmount:    so.PaidAmount,
			UnpaidAmount:  unpaid,
			DueDate:       so.DueDate,
			CreatedAt:     so.CreatedAt,
			UpdatedAt:     so.UpdatedAt,
		})
	}

	totalPages := int((totalCount + int64(filters.Limit) - 1) / int64(filters.Limit))
	paginationResponse := models.PaginationResponse{
		CurrentPage:  filters.Page,
		PerPage:      filters.Limit,
		TotalPages:   totalPages,
		TotalRecords: totalCount,
		HasNext:      filters.Page < totalPages,
		HasPrev:      filters.Page > 1,
	}

	return &models.SalesReportDetailItemPaginatedResponse{
		Data:       detailsResponse,
		Pagination: paginationResponse,
	}, nil
}

func (s *SalesReportService) GetSalesReportInsights(filters *models.PaginationRequest) (*models.SalesReportInsights, error) {
	if err := s.setDefaultDateRange(filters); err != nil {
		return nil, err
	}
	return s.getInsightsData(filters)
}

func (s *SalesReportService) GenerateSalesReportExcel(filters *models.PaginationRequest) (string, string, error) {
	if filters == nil {
		filters = &models.PaginationRequest{}
	}
	_ = s.setDefaultDateRange(filters)

	page := 1
	limit := 2000
	all := make([]models.SalesOrder, 0, 4096)

	for {
		f := *filters
		f.Page = page
		f.Limit = limit

		chunk, totalCount, err := s.SalesReportRepo.GetSalesReportDetails(nil, &f)
		if err != nil {
			return "", "", err
		}
		all = append(all, chunk...)

		if int64(page*limit) >= totalCount || len(chunk) == 0 {
			break
		}
		page++
	}

	rows := make([]models.SalesReportDetailItem, 0, len(all))
	for _, so := range all {
		unpaid := so.TotalAmount - so.PaidAmount
		if unpaid < 0 {
			unpaid = 0
		}
		rows = append(rows, models.SalesReportDetailItem{
			ID:            so.ID,
			SONumber:      so.SONumber,
			SODate:        so.SODate,
			SalesPerson:   so.SalesPerson,
			Customer:      so.Customer,
			SOStatus:      so.SOStatus,
			PaymentStatus: so.PaymentStatus,
			TotalAmount:   so.TotalAmount,
			PaidAmount:    so.PaidAmount,
			UnpaidAmount:  unpaid,
			DueDate:       so.DueDate,
			CreatedAt:     so.CreatedAt,
			UpdatedAt:     so.UpdatedAt,
		})
	}

	dir := filepath.Join("public", "temp", "excel")
	fullPath, filename, err := documents.GenerateSalesReportExcel(rows, dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate excel: %w", err)
	}

	return fullPath, filename, nil
}

func (service *SalesReportService) GenerateSalesReportPDF(soId string) (string, []byte, error) {
	so, err := service.SalesOrderRepo.FindById(nil, soId, true)
	if err != nil {
		return "", nil, fmt.Errorf("Sales order not found: %w", err)
	}

	unpaid := so.TotalAmount - so.PaidAmount
	if unpaid < 0 {
		unpaid = 0
	}
	srdi := &models.SalesReportDetailItem{
		ID:            so.ID,
		SONumber:      so.SONumber,
		SODate:        so.SODate,
		SalesPerson:   so.SalesPerson,
		Customer:      so.Customer,
		SOStatus:      so.SOStatus,
		PaymentStatus: so.PaymentStatus,
		TotalAmount:   so.TotalAmount,
		PaidAmount:    so.PaidAmount,
		UnpaidAmount:  unpaid,
		DueDate:       so.DueDate,
		CreatedAt:     so.CreatedAt,
		UpdatedAt:     so.UpdatedAt,
	}

	filename, data, err := documents.GenerateSalesReportPDF(srdi)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate sales report PDF: %w", err)
	}
	return filename, data, nil
}

func (s *SalesReportService) GetSalesPersonIDByUserID(userID uuid.UUID) (uuid.UUID, error) {
	sales, err := s.SalesPersonRepo.FindById(nil, userID.String(), false)
	if err != nil {
		return uuid.UUID{}, err
	}
	return sales.ID, nil
}

// ===== Helper internal (tetap private) =====
func jakartaLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
					return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}

func (s *SalesReportService) setDefaultDateRange(filters *models.PaginationRequest) error {
	loc := jakartaLoc()
 now := time.Now().In(loc)

	hasStart := !filters.StartDate.IsZero()
	hasEnd := !filters.EndDate.IsZero()

	if hasStart || hasEnd {
		if !hasStart || !hasEnd {
			return fmt.Errorf("both start_date and end_date must be provided together")
		}
		start := filters.StartDate.In(loc)
		end := filters.EndDate.In(loc)
		if end.Before(start) {
			return fmt.Errorf("end_date must be after or equal to start_date")
		}
		filters.StartDate = start
		filters.EndDate = end
		if filters.Period == "" {
			filters.Period = "custom"
		}
		return nil
	}

	switch filters.Period {
	case "day":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end := start.Add(24*time.Hour - time.Nanosecond)
		filters.StartDate, filters.EndDate = start, end
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := now.AddDate(0, 0, -weekday+1)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)
		filters.StartDate, filters.EndDate = start, end
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		filters.StartDate, filters.EndDate = start, end
	case "year":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		end := start.AddDate(1, 0, 0).Add(-time.Nanosecond)
		filters.StartDate, filters.EndDate = start, end
	case "custom":
		return fmt.Errorf("start_date and end_date are required for custom period")
	default:
		return fmt.Errorf("invalid period: %s", filters.Period)
	}
	return nil
}

func (s *SalesReportService) getSummaryData(filters *models.PaginationRequest) (*models.SalesReportSummary, error) {
	currentData, err := s.SalesReportRepo.GetSalesReportSummaryData(nil, filters)
	if err != nil {
		return nil, err
	}

	previousFilters := s.getPreviousPeriodFilters(filters)
	previousData, err := s.SalesReportRepo.GetSalesReportSummaryData(nil, previousFilters)
	if err != nil {
		return nil, err
	}

	summary := &models.SalesReportSummary{
		TotalSales: models.SalesReportSummaryItem{
			Value:     float64(currentData.TotalSales),
			ChangePct: s.calculateChangePct(float64(currentData.TotalSales), float64(previousData.TotalSales)),
			Trend:     s.calculateTrend(float64(currentData.TotalSales), float64(previousData.TotalSales)),
		},
		TotalPaid: models.SalesReportSummaryItem{
			Value:     float64(currentData.TotalPaid),
			ChangePct: s.calculateChangePct(float64(currentData.TotalPaid), float64(previousData.TotalPaid)),
			Trend:     s.calculateTrend(float64(currentData.TotalPaid), float64(previousData.TotalPaid)),
		},
		OutstandingPayment: models.SalesReportSummaryItem{
			Value:     float64(currentData.OutstandingPayment),
			ChangePct: s.calculateChangePct(float64(currentData.OutstandingPayment), float64(previousData.OutstandingPayment)),
			Trend:     s.calculateTrend(float64(currentData.OutstandingPayment), float64(previousData.OutstandingPayment)),
		},
		TotalOrders: models.SalesReportSummaryItem{
			Value:     float64(currentData.TotalOrders),
			ChangePct: s.calculateChangePct(float64(currentData.TotalOrders), float64(previousData.TotalOrders)),
			Trend:     s.calculateTrend(float64(currentData.TotalOrders), float64(previousData.TotalOrders)),
		},
		UniqueCustomers: models.SalesReportSummaryItem{
			Value:     float64(currentData.UniqueCustomers),
			ChangePct: s.calculateChangePct(float64(currentData.UniqueCustomers), float64(previousData.UniqueCustomers)),
			Trend:     s.calculateTrend(float64(currentData.UniqueCustomers), float64(previousData.UniqueCustomers)),
		},
		TopSalesPerson: models.SalesReportTopItem{
			Name:      currentData.TopSalesPersonName,
			Value:     float64(currentData.TopSalesPersonValue),
			ChangePct: s.calculateChangePct(float64(currentData.TopSalesPersonValue), float64(previousData.TopSalesPersonValue)),
			Trend:     s.calculateTrend(float64(currentData.TopSalesPersonValue), float64(previousData.TopSalesPersonValue)),
		},
		TopArea: models.SalesReportTopItem{
			Name:      currentData.TopAreaName,
			Value:     float64(currentData.TopAreaValue),
			ChangePct: s.calculateChangePct(float64(currentData.TopAreaValue), float64(previousData.TopAreaValue)),
			Trend:     s.calculateTrend(float64(currentData.TopAreaValue), float64(previousData.TopAreaValue)),
		},
	}
	return summary, nil
}

func (s *SalesReportService) getChartsData(filters *models.PaginationRequest) (*models.SalesReportChartData, error) {
	salesTrendData, err := s.SalesReportRepo.GetSalesTrendData(nil, filters)
	if err != nil {
		return nil, err
	}

	// Sales trend repo sudah handle default date range sendiri,
	// tapi tetap kita set supaya konsisten untuk query lain.
	if err := s.setDefaultDateRange(filters); err != nil {
		return nil, err
	}

	statusDistData, err := s.SalesReportRepo.GetStatusDistributionData(nil, filters)
	if err != nil {
		return nil, err
	}
	salesByPersonData, err := s.SalesReportRepo.GetSalesByPersonData(nil, filters)
	if err != nil {
		return nil, err
	}
	paymentCompData, err := s.SalesReportRepo.GetPaymentComparisonData(nil, filters)
	if err != nil {
		return nil, err
	}
	heatmapData, err := s.SalesReportRepo.GetHeatmapData(nil, filters)
	if err != nil {
		return nil, err
	}

	charts := &models.SalesReportChartData{
		SalesTrend: struct {
			Labels   []string              `json:"labels"`
			Datasets []models.ChartDataset `json:"datasets"`
		}{
			Labels:   salesTrendData.Labels,
			Datasets: salesTrendData.Datasets,
		},
		StatusDistribution: struct {
			Labels   []string              `json:"labels"`
			Datasets []models.ChartDataset `json:"datasets"`
		}{
			Labels:   statusDistData.Labels,
			Datasets: statusDistData.Datasets,
		},
		SalesByPerson: struct {
			Labels   []string              `json:"labels"`
			Datasets []models.ChartDataset `json:"datasets"`
		}{
			Labels:   salesByPersonData.Labels,
			Datasets: salesByPersonData.Datasets,
		},
		PaymentComparison: struct {
			Labels   []string              `json:"labels"`
			Datasets []models.ChartDataset `json:"datasets"`
		}{
			Labels:   paymentCompData.Labels,
			Datasets: paymentCompData.Datasets,
		},
		HeatmapData: heatmapData,
	}
	return charts, nil
}

func (s *SalesReportService) getInsightsData(filters *models.PaginationRequest) (*models.SalesReportInsights, error) {
	topItems, err := s.SalesReportRepo.GetTopItems(nil, filters)
	if err != nil {
		return nil, err
	}
	topCustomers, err := s.SalesReportRepo.GetTopCustomers(nil, filters)
	if err != nil {
		return nil, err
	}
	overdueInvoices, err := s.SalesReportRepo.GetOverdueInvoices(nil, filters)
	if err != nil {
		return nil, err
	}
	performance, err := s.SalesReportRepo.GetPerformanceData(nil, filters)
	if err != nil {
		return nil, err
	}
	insights := &models.SalesReportInsights{
		TopItems:        topItems,
		TopCustomers:    topCustomers,
		OverdueInvoices: overdueInvoices,
		Performance:     performance,
	}
	return insights, nil
}

func (s *SalesReportService) getPreviousPeriodFilters(filters *models.PaginationRequest) *models.PaginationRequest {
	// FIX: buat salinan, jangan referensi objek yang sama
	prev := *filters
	if !filters.StartDate.IsZero() && !filters.EndDate.IsZero() {
		duration := filters.EndDate.Sub(filters.StartDate)
		prevStart := filters.StartDate.Add(-duration)
		prevEnd := filters.StartDate.Add(-time.Nanosecond)
		prev.StartDate = prevStart
		prev.EndDate = prevEnd
	}
	return &prev
}

func (s *SalesReportService) calculateChangePct(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return ((current - previous) / previous) * 100
}

func (s *SalesReportService) calculateTrend(current, previous float64) string {
	if current > previous {
		return "up"
	} else if current < previous {
		return "down"
	}
	return "same"
}
