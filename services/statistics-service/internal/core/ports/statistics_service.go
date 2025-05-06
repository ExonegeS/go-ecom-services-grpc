package ports

import (
	"context"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/domain"
)

type StatisticsService interface {
	HandleOrderCreated(event domain.OrderEvent)
	HandleOrderUpdated(event domain.OrderEvent)
	HandleOrderDeleted(event domain.OrderEvent)

	GetUserOrderStats(ctx context.Context, userID domain.UUID) (*domain.UserOrderStats, error)
	GetUserStatistics(ctx context.Context, userID domain.UUID) (*domain.UserStats, error)
}
