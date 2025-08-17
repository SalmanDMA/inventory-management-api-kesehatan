package services

import (
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
)

type DashboardService struct {
	ItemRepository       repositories.ItemRepository
	SalesOrderRepository repositories.SalesOrderRepository
}

func NewDashboardService(
	itemRepository repositories.ItemRepository,
	salesOrderRepository repositories.SalesOrderRepository,
) *DashboardService {
	return &DashboardService{
		ItemRepository:       itemRepository,
		SalesOrderRepository: salesOrderRepository,
	}
}

func (s *DashboardService) GetDashboardSummary() (*models.DashboardSummary, error) {
	currentTotalItems, err := s.ItemRepository.CountAllThisMonth()
	if err != nil {
		return nil, err
	}
	previousTotalItems, err := s.ItemRepository.CountAllLastMonth()
	if err != nil {
		return nil, err
	}

	// Low Stock (snapshot vs "last month" — fallback)
	lowNow, err := s.ItemRepository.CountLowStockNow()
	if err != nil {
		return nil, err
	}
	lowPrev, err := s.ItemRepository.CountLowStockLastMonth()
	if err != nil {
		return nil, err
	}

	// Active Orders = minimal Confirmed (Confirmed, Shipped, Delivered)
	activeNow, err := s.SalesOrderRepository.CountActiveThisMonth()
	if err != nil {
		return nil, err
	}
	activePrev, err := s.SalesOrderRepository.CountActiveLastMonth()
	if err != nil {
		return nil, err
	}

	// Total Value = SUM(total_amount) untuk SO Closed
	valueNow, err := s.SalesOrderRepository.SumClosedValueThisMonth()
	if err != nil {
		return nil, err
	}
	valuePrev, err := s.SalesOrderRepository.SumClosedValueLastMonth()
	if err != nil {
		return nil, err
	}

	return &models.DashboardSummary{
		TotalItems:    calculateMetric[int64](currentTotalItems, previousTotalItems),
		LowStockItems: calculateMetric[int64](lowNow, lowPrev),
		ActiveOrders:  calculateMetric[int64](activeNow, activePrev),
		TotalValue:    calculateMetric[float64](valueNow, valuePrev),
	}, nil
}

func calculateMetric[T int64 | float64](current, previous T) models.MetricWithChange[T] {
    change := current - previous

    var pct int
    if previous != 0 {
        pct = int(float64(change) / float64(previous) * 100)
    }

    trend := "same"
    if change > 0 {
        trend = "up"
    } else if change < 0 {
        trend = "down"
    }

    return models.MetricWithChange[T]{
        Value:     current,
        Change:    int(change),
        ChangePct: pct,
        Trend:     trend,
    }
}