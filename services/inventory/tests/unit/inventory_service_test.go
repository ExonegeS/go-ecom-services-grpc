package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
)

type mockInventoryRepository struct {
	getByIDFunc func(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error)
	saveFunc    func(ctx context.Context, item entity.InventoryItem) error
	updateFunc  func(ctx context.Context, id entity.UUID, updateFn func(*entity.InventoryItem) (bool, error)) error
}

func (m *mockInventoryRepository) GetByID(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
	return m.getByIDFunc(ctx, id)
}
func (m *mockInventoryRepository) Save(ctx context.Context, item entity.InventoryItem) error {
	return m.saveFunc(ctx, item)
}
func (m *mockInventoryRepository) UpdateByID(ctx context.Context, id entity.UUID, updateFn func(*entity.InventoryItem) (bool, error)) error {
	return m.updateFunc(ctx, id, updateFn)
}
func (m *mockInventoryRepository) DeleteByID(ctx context.Context, id entity.UUID) error {
	return nil
}
func (m *mockInventoryRepository) GetAllInventoryItems(ctx context.Context, pagination *entity.Pagination) ([]*entity.InventoryItem, error) {
	return nil, nil
}
func (m *mockInventoryRepository) GetTotalCount(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m mockInventoryRepository) GetCategoryByID(ctx context.Context, id entity.UUID) (*entity.Category, error) {
	return &entity.Category{
		ID:          id,
		Name:        "Test Category",
		Description: "Test Description",
	}, nil
}
func (m mockInventoryRepository) SaveCategory(ctx context.Context, item entity.Category) error {
	return nil
}
func (m mockInventoryRepository) UpdateCategoryByID(ctx context.Context, id entity.UUID, updateFn func(*entity.Category) (bool, error)) error {
	return nil
}
func (m mockInventoryRepository) DeleteCategoryByID(ctx context.Context, id entity.UUID) error {
	return nil
}
func (m mockInventoryRepository) GetAllCategories(ctx context.Context, pagination *entity.Pagination) ([]*entity.Category, error) {
	return nil, nil
}
func (m *mockInventoryRepository) GetTotalCategoriesCount(ctx context.Context) (int64, error) {
	return 0, nil
}

func TestGetInventoryItemByID_Success(t *testing.T) {
	expectedItem := &entity.InventoryItem{
		ID:       "1",
		Name:     "Test Item",
		Quantity: 10,
		Unit:     "pcs",
		Price:    19.99,
	}
	repo := &mockInventoryRepository{
		getByIDFunc: func(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
			if id == expectedItem.ID {
				return expectedItem, nil
			}
			return nil, entity.ErrItemNotFound
		},
	}

	service := application.NewInventoryService(repo, nil)
	item, err := service.GetInventoryItemByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.ID != expectedItem.ID || item.Name != expectedItem.Name {
		t.Fatalf("expected %v, got %v", expectedItem, item)
	}
}

func TestGetInventoryItemByID_NotFound(t *testing.T) {
	repo := &mockInventoryRepository{
		getByIDFunc: func(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
			return nil, entity.ErrItemNotFound
		},
	}
	service := application.NewInventoryService(repo, nil)
	_, err := service.GetInventoryItemByID(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, entity.ErrItemNotFound) && err.Error() != "inventory item not found" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateInventoryItem_Success(t *testing.T) {
	expectedItem := entity.InventoryItem{
		Name:        "Test Item",
		Description: "Test Description",
		Quantity:    10,
		Unit:        "pcs",
		Price:       19.99,
	}
	repo := &mockInventoryRepository{
		saveFunc: func(ctx context.Context, item entity.InventoryItem) error {
			c, _ := mockInventoryRepository.GetCategoryByID(mockInventoryRepository{}, ctx, entity.NewUUID())
			item.Category = *c
			if item.ID == expectedItem.ID {
				return nil
			}
			return errors.New("failed to save item")
		},
	}

	// list me a func () time.Time, that simulate different dates
	timeSource := func() time.Time {
		return time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	service := application.NewInventoryService(repo, timeSource)
	err := service.CreateInventoryItem(context.Background(), &expectedItem)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

}
