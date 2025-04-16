package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/ports"
)

type InventoryService interface {
	GetInventoryItemByID(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error)
	CreateInventoryItem(ctx context.Context, item *entity.InventoryItem) error
	UpdateInventoryItem(ctx context.Context, id entity.UUID, params UpdateInventoryItemParams) (*entity.InventoryItem, error)
	DeleteInventoryItem(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error)
	GetPaginatedInventoryItems(ctx context.Context, pagination *entity.Pagination) (*entity.PaginationResponse[*entity.InventoryItem], error)

	GetCategoryByID(ctx context.Context, id entity.UUID) (*entity.Category, error)
	CreateCategory(ctx context.Context, category *entity.Category) error
	UpdateCategory(ctx context.Context, id entity.UUID, params UpdateCategoryParams) (*entity.Category, error)
	DeleteCategory(ctx context.Context, id entity.UUID) (*entity.Category, error)
	GetPaginatedCategories(ctx context.Context, pagination *entity.Pagination) (*entity.PaginationResponse[*entity.Category], error)

	ReserveProduct(ctx context.Context, id entity.UUID, quantity int64) error
}

type UpdateInventoryItemParams struct {
	Name        *string
	Description *string
	CategoryID  *entity.UUID
	Price       *float64
	Quantity    *float64
	Unit        *string
}

type UpdateCategoryParams struct {
	Name        *string
	Description *string
}

type inventoryService struct {
	inventoryRepo ports.InventoryRepository
	timeSource    func() time.Time
}

func NewInventoryService(inventoryRepo ports.InventoryRepository, timeSource func() time.Time) InventoryService {
	return &inventoryService{
		inventoryRepo: inventoryRepo,
		timeSource:    timeSource,
	}
}

func (s *inventoryService) GetInventoryItemByID(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
	item, err := s.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrItemNotFound) {
			return nil, err
		}
		return nil, err
	}
	return item, nil
}

func (s *inventoryService) CreateInventoryItem(ctx context.Context, item *entity.InventoryItem) error {
	if s.inventoryRepo == nil || s.timeSource == nil {
		return fmt.Errorf("dependencies not initialized")
	}

	if err := item.Validate(); err != nil {
		return fmt.Errorf("invalid inventory item: %w", err)
	}

	category, err := s.inventoryRepo.GetCategoryByID(ctx, item.Category.ID)
	if err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return err
		}
		if errors.Is(err, entity.ErrNotImplemented) {
			return err
		}
		return fmt.Errorf("failed to get category: %w", err)
	}

	item.ID = entity.NewUUID()
	item.CreatedAt = s.timeSource().UTC()
	item.UpdatedAt = item.CreatedAt
	item.Category = *category

	if err := s.inventoryRepo.Save(ctx, *item); err != nil {

		return fmt.Errorf("failed to save inventory item: %w", err)
	}

	return nil
}
func (s *inventoryService) DeleteInventoryItem(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
	item, err := s.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrItemNotFound) {
			return nil, err
		}
		if errors.Is(err, entity.ErrNotImplemented) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}
	if err := s.inventoryRepo.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, entity.ErrItemNotFound) {
			return nil, err
		}
		if errors.Is(err, entity.ErrNotImplemented) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to delete inventory item: %w", err)
	}
	return item, nil
}

func (s *inventoryService) GetPaginatedInventoryItems(ctx context.Context, pagination *entity.Pagination) (*entity.PaginationResponse[*entity.InventoryItem], error) {
	const op = "inventoryService.GetPaginatedInventoryItems"

	totalItems, err := s.inventoryRepo.GetTotalCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	totalPages := (totalItems + pagination.PageSize - 1) / pagination.PageSize

	items, err := s.inventoryRepo.GetAllInventoryItems(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.PaginationResponse[*entity.InventoryItem]{
		CurrentPage: pagination.Page,
		HasNextPage: pagination.Page < totalPages,
		PageSize:    pagination.PageSize,
		TotalPages:  totalPages,
		Data:        items,
	}, nil
}

func (s *inventoryService) GetPaginatedCategories(ctx context.Context, pagination *entity.Pagination) (*entity.PaginationResponse[*entity.Category], error) {
	const op = "inventoryService.GetPaginatedCategories"

	totalItems, err := s.inventoryRepo.GetTotalCategoriesCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	totalPages := (totalItems + pagination.PageSize - 1) / pagination.PageSize

	items, err := s.inventoryRepo.GetAllCategories(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.PaginationResponse[*entity.Category]{
		CurrentPage: pagination.Page,
		HasNextPage: pagination.Page < totalPages,
		PageSize:    pagination.PageSize,
		TotalPages:  totalPages,
		Data:        items,
	}, nil
}

func (s *inventoryService) GetCategoryByID(ctx context.Context, id entity.UUID) (*entity.Category, error) {
	category, err := s.inventoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, err
		}
		if errors.Is(err, entity.ErrNotImplemented) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return category, nil
}

func (s *inventoryService) CreateCategory(ctx context.Context, category *entity.Category) error {
	if s.inventoryRepo == nil || s.timeSource == nil {
		return fmt.Errorf("dependencies not initialized")
	}

	if err := category.Validate(); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}

	category.ID = entity.NewUUID()
	category.CreatedAt = s.timeSource().UTC()
	category.UpdatedAt = category.CreatedAt

	if err := s.inventoryRepo.SaveCategory(ctx, *category); err != nil {
		return fmt.Errorf("failed to save category: %w", err)
	}

	return nil
}

func (s *inventoryService) UpdateCategory(ctx context.Context, id entity.UUID, params UpdateCategoryParams) (*entity.Category, error) {
	// const op = "service.UpdateCategoryById"
	var categoryData *entity.Category

	return categoryData, s.inventoryRepo.UpdateCategoryByID(ctx, id, func(category *entity.Category) (updated bool, err error) {
		if params.Name != nil && *params.Name != category.Name {
			category.Name = *params.Name
			updated = true
		}

		if params.Description != nil && *params.Description != category.Description {
			category.Description = *params.Description
			updated = true
		}

		if !updated {
			return
		}

		category.UpdatedAt = s.timeSource().UTC()
		categoryData = category

		return
	})
}

func (s *inventoryService) DeleteCategory(ctx context.Context, id entity.UUID) (*entity.Category, error) {
	category, err := s.inventoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, err
		}
		if errors.Is(err, entity.ErrNotImplemented) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if err := s.inventoryRepo.DeleteCategoryByID(ctx, id); err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, err
		}
		if errors.Is(err, entity.ErrNotImplemented) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to delete category: %w", err)
	}
	return category, nil
}

func (s *inventoryService) UpdateInventoryItem(ctx context.Context, id entity.UUID, params UpdateInventoryItemParams) (*entity.InventoryItem, error) {
	// const op = "service.UpdateInventoryItemById"
	var itemData *entity.InventoryItem

	return itemData, s.inventoryRepo.UpdateByID(ctx, id, func(item *entity.InventoryItem) (updated bool, err error) {
		if params.Name != nil && *params.Name != item.Name {
			item.Name = *params.Name
			updated = true
		}

		if params.Description != nil && *params.Description != item.Description {
			item.Description = *params.Description
			updated = true
		}

		if params.CategoryID != nil && *params.CategoryID != item.Category.ID {
			_, err = s.inventoryRepo.GetCategoryByID(ctx, *params.CategoryID)
			if err != nil {
				if errors.Is(err, entity.ErrCategoryNotFound) {
					return
				}
				if errors.Is(err, entity.ErrNotImplemented) {
					return
				}
				err = fmt.Errorf("failed to get category: %w", err)
				return
			}

			item.Category.ID = *params.CategoryID
			updated = true
		}
		if params.Price != nil && *params.Price != item.Price {
			if *params.Price < 0 {
				err = entity.ErrInvalidPrice
				return
			}
			item.Price = *params.Price
			updated = true
		}

		if params.Quantity != nil && *params.Quantity != item.Quantity {
			if *params.Quantity < 0 {
				err = entity.ErrInvalidQuantity
				return
			}
			item.Quantity = *params.Quantity
			updated = true
		}

		if params.Unit != nil && *params.Unit != item.Unit {
			item.Unit = *params.Unit
			updated = true
		}

		if !updated {
			return
		}

		item.UpdatedAt = s.timeSource().UTC()
		itemData = item

		return
	})
}

func (s *inventoryService) ReserveProduct(ctx context.Context, id entity.UUID, quantity int64) error {
	return s.inventoryRepo.UpdateByID(ctx, id, func(item *entity.InventoryItem) (updated bool, err error) {
		if quantity < 0 {
			err = entity.ErrInvalidQuantity
			return
		}
		if quantity > int64(item.Quantity) {
			err = entity.ErrInsufficientQuantity
			return
		}

		item.Quantity -= float64(quantity)
		item.UpdatedAt = s.timeSource().UTC()

		return true, nil
	})
}
