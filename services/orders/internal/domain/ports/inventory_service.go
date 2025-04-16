package ports

import (
	"context"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
)

type InventoryService interface {
	GetProduct(ctx context.Context, productID entity.UUID) (*entity.OrderItem, error)
	ReserveItem(ctx context.Context, productID entity.UUID, quantity int64) (bool, error)
}
