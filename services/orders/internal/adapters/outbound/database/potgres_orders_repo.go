package database

import (
	"context"
	"database/sql"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/ports"
)

type postgresOrdersRepository struct {
	db *sql.DB
}

func NewPostgresOrdersRepository(db *sql.DB) ports.OrdersRepository {
	return &postgresOrdersRepository{db: db}
}

func (r *postgresOrdersRepository) GetOrderByID(ctx context.Context, id entity.UUID) (*entity.Order, error) {
	return nil, entity.ErrNotImplemented
}
func (r *postgresOrdersRepository) SaveOrder(ctx context.Context, item entity.Order) error {
	return entity.ErrNotImplemented
}
func (r *postgresOrdersRepository) UpdateOrderByID(ctx context.Context, id entity.UUID, updateFn func(*entity.Order) (bool, error)) error {
	return entity.ErrNotImplemented
}
func (r *postgresOrdersRepository) DeleteOrderByID(ctx context.Context, id entity.UUID) error {
	return entity.ErrNotImplemented
}
func (r *postgresOrdersRepository) GetTotalOrdersCount(ctx context.Context) (int64, error) {
	return 0, entity.ErrNotImplemented
}
func (r *postgresOrdersRepository) GetAllOrders(ctx context.Context, pagination *entity.Pagination) ([]*entity.Order, error) {
	return nil, entity.ErrNotImplemented
}
