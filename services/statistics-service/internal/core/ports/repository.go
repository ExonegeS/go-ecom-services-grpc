package ports

import (
	"context"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/domain"
)

type StatisticsRepository interface {
	UpdateOrderStatistics(ctx context.Context, event domain.OrderEvent) error
	GetUserOrderStats(ctx context.Context, userID domain.UUID) (*domain.UserOrderStats, error)
	GetUserStatistics(ctx context.Context, userID domain.UUID) (*domain.UserStats, error)
}
