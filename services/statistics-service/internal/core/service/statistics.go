package service

import (
	"context"
	"log/slog"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/domain"
	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/ports"
)

type StatisticsService struct {
	statisticsRepo ports.StatisticsRepository
	logger         *slog.Logger
}

func NewStatisticsService(statisticsRepo ports.StatisticsRepository, logger *slog.Logger) *StatisticsService {
	return &StatisticsService{
		statisticsRepo: statisticsRepo,
		logger:         logger,
	}
}

func (s *StatisticsService) HandleOrderCreated(event domain.OrderEvent) {
	const op = "StatisticsService.HandleOrderCreated"
	err := s.statisticsRepo.UpdateOrderStatistics(context.Background(), event)
	if err != nil {
		s.logger.Error("Failed to update order statistics", slog.String("err", err.Error()))
	}
}
func (s *StatisticsService) HandleOrderUpdated(event domain.OrderEvent) {
	const op = "StatisticsService.HandleOrderUpdated"
	err := s.statisticsRepo.UpdateOrderStatistics(context.Background(), event)
	if err != nil {
		s.logger.Error("Failed to update order statistics", slog.String("err", err.Error()))
	}
}
func (s *StatisticsService) HandleOrderDeleted(event domain.OrderEvent) {
	const op = "StatisticsService.HandleOrderDeleted"
	err := s.statisticsRepo.UpdateOrderStatistics(context.Background(), event)
	if err != nil {
		s.logger.Error("Failed to update order statistics", slog.String("err", err.Error()))
	}
}

func (s *StatisticsService) GetUserOrderStats(ctx context.Context, userID domain.UUID) (*domain.UserOrderStats, error) {
	return s.statisticsRepo.GetUserOrderStats(ctx, userID)
}

func (s *StatisticsService) GetUserStatistics(ctx context.Context, userID domain.UUID) (*domain.UserStats, error) {
	return s.statisticsRepo.GetUserStatistics(ctx, userID)
}
