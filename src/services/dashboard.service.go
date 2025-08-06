package services

import (
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
)

type DashboardService struct {
	ItemRepository      repositories.ItemRepository
}

func NewDashboardService(itemRepository repositories.ItemRepository) *DashboardService {
	return &DashboardService{
		ItemRepository:      itemRepository,
	}
}

func (s *DashboardService) GetDashboardSummary() (*models.DashboardSummary, error) {
	currentTotalItems, _ := s.ItemRepository.CountAllThisMonth()
	previousTotalItems, _ := s.ItemRepository.CountAllLastMonth()

	return &models.DashboardSummary{
		TotalItems:    calculateMetric(currentTotalItems, previousTotalItems),
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
