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
	// Items
	currentTotalItems, err := s.ItemRepository.CountAllThisMonth(nil)
	if err != nil {
		return nil, err
	}
	previousTotalItems, err := s.ItemRepository.CountAllLastMonth(nil)
	if err != nil {
		return nil, err
	}

	// Low stock
	lowNow, err := s.ItemRepository.CountLowStockNow(nil)
	if err != nil {
		return nil, err
	}
	lowPrev, err := s.ItemRepository.CountLowStockLastMonth(nil)
	if err != nil {
		return nil, err
	}

	// Active orders (Confirmed/On-going)
	activeNow, err := s.SalesOrderRepository.CountActiveThisMonth(nil)
	if err != nil {
		return nil, err
	}
	activePrev, err := s.SalesOrderRepository.CountActiveLastMonth(nil)
	if err != nil {
		return nil, err
	}

	// Closed value
	valueNow, err := s.SalesOrderRepository.SumClosedValueThisMonth(nil)
	if err != nil {
		return nil, err
	}
	valuePrev, err := s.SalesOrderRepository.SumClosedValueLastMonth(nil)
	if err != nil {
		return nil, err
	}

	return &models.DashboardSummary{
		TotalItems:    calculateMetric(currentTotalItems, previousTotalItems),
		LowStockItems: calculateMetric(lowNow, lowPrev),
		ActiveOrders:  calculateMetric(activeNow, activePrev),
		TotalValue:    calculateMetric(valueNow, valuePrev),
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
