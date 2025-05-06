package service

import (
	"context"
	"fmt"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/domain"
	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/ports"
)

type StatisticsService struct {
	statisticsRepo ports.StatisticsRepository
}

func NewStatisticsService(statisticsRepo ports.StatisticsRepository) *StatisticsService {
	return &StatisticsService{
		statisticsRepo: statisticsRepo,
	}
}

func (s *StatisticsService) HandleOrderCreated(event domain.OrderEvent) {
	const op = "StatisticsService.HandleOrderCreated"
	fmt.Println(op)
}
func (s *StatisticsService) HandleOrderUpdated(event domain.OrderEvent) {
	const op = "StatisticsService.HandleOrderUpdated"
	fmt.Println(op)
}
func (s *StatisticsService) HandleOrderDeleted(event domain.OrderEvent) {
	const op = "StatisticsService.HandleOrderDeleted"
	fmt.Println(op)
}

func (s *StatisticsService) GetUserOrderStats(context.Context, domain.UUID) (*domain.UserOrderStats, error) {
	return nil, domain.ErrNotImplemented
}

func (s *StatisticsService) GetUserStatistics(ctx context.Context, userID domain.UUID) (*domain.UserStats, error) {
	return nil, domain.ErrNotImplemented
}
