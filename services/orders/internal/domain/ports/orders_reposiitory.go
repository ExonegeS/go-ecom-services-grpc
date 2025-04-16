package ports

import (
	"context"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
)

type OrdersRepository interface {
	GetOrderByID(ctx context.Context, id entity.UUID) (*entity.Order, error)
	SaveOrder(ctx context.Context, item entity.Order) error
	UpdateOrderByID(ctx context.Context, id entity.UUID, updateFn func(*entity.Order) (bool, error)) error
	DeleteOrderByID(ctx context.Context, id entity.UUID) error
	GetTotalOrdersCount(ctx context.Context) (int64, error)
	GetAllOrders(ctx context.Context, pagination *entity.Pagination) ([]*entity.Order, error)
}
