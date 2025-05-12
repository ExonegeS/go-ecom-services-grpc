package ports

import (
	"context"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
)

type OrdersRepository interface {
	Order(ctx context.Context, id entity.UUID) (*entity.Order, error)
	Save(ctx context.Context, item entity.Order) error
	UpdateByID(ctx context.Context, id entity.UUID, updateFn func(*entity.Order) (bool, error)) error
	DeleteByID(ctx context.Context, id entity.UUID) error
	GetTotalOrdersCount(ctx context.Context) (int64, error)
	GetAllOrders(ctx context.Context, pagination *entity.Pagination) ([]*entity.Order, error)
}
