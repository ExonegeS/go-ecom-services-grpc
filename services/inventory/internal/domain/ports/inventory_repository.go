package ports

import (
	"context"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
)

type InventoryRepository interface {
	GetByID(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error)
	Save(ctx context.Context, item entity.InventoryItem) error
	UpdateByID(ctx context.Context, id entity.UUID, updateFn func(*entity.InventoryItem) (bool, error)) error
	DeleteByID(ctx context.Context, id entity.UUID) error
	GetTotalCount(ctx context.Context) (int64, error)
	GetAllInventoryItems(ctx context.Context, pagination *entity.Pagination) ([]*entity.InventoryItem, error)
	CategoryRepository
}
