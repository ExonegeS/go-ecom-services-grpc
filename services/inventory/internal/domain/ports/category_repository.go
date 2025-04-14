package ports

import (
	"context"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
)

type CategoryRepository interface {
	GetCategoryByID(ctx context.Context, id entity.UUID) (*entity.Category, error)
	SaveCategory(ctx context.Context, item entity.Category) error
	UpdateCategoryByID(ctx context.Context, id entity.UUID, updateFn func(*entity.Category) (bool, error)) error
	DeleteCategoryByID(ctx context.Context, id entity.UUID) error
	GetTotalCategoriesCount(ctx context.Context) (int64, error)
	GetAllCategories(ctx context.Context, pagination *entity.Pagination) ([]*entity.Category, error)
}
