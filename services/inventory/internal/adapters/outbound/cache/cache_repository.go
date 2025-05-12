package cache

import (
	"context"
	"sync"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/ports"
)

type CacheRepository struct {
	repo          ports.InventoryRepository
	cache         map[entity.UUID]interface{}
	itemsList     []*entity.InventoryItem
	categories    []*entity.Category
	mu            sync.RWMutex
	lastRefreshed time.Time
}

func NewCacheRepository(repo ports.InventoryRepository) *CacheRepository {
	c := &CacheRepository{
		repo:       repo,
		cache:      make(map[entity.UUID]interface{}),
		itemsList:  make([]*entity.InventoryItem, 0),
		categories: make([]*entity.Category, 0),
	}

	go c.startRefreshTicker()
	return c
}

func (c *CacheRepository) startRefreshTicker() {
	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.refreshCache(context.Background())
	}
}

func (c *CacheRepository) refreshCache(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	totalItems, _ := c.repo.GetTotalCount(ctx)
	pagination := &entity.Pagination{Page: 1, PageSize: totalItems}
	items, _ := c.repo.GetAllInventoryItems(ctx, pagination)
	c.itemsList = items

	for _, item := range items {
		c.cache[item.ID] = item
	}

	totalCategories, _ := c.repo.GetTotalCategoriesCount(ctx)
	catPagination := &entity.Pagination{Page: 1, PageSize: totalCategories}
	categories, _ := c.repo.GetAllCategories(ctx, catPagination)
	c.categories = categories

	c.lastRefreshed = time.Now()
}

func (c *CacheRepository) GetByID(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
	c.mu.RLock()
	if item, exists := c.cache[id].(*entity.InventoryItem); exists {
		c.mu.RUnlock()
		return item, nil
	}
	c.mu.RUnlock()

	item, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[id] = item
	c.mu.Unlock()

	return item, nil
}

func (c *CacheRepository) Save(ctx context.Context, item entity.InventoryItem) error {
	if err := c.repo.Save(ctx, item); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[item.ID] = &item
	c.itemsList = append(c.itemsList, &item)
	return nil
}

func (c *CacheRepository) UpdateByID(ctx context.Context, id entity.UUID, updateFn func(*entity.InventoryItem) (bool, error)) error {
	c.mu.RLock()

	err := c.repo.UpdateByID(ctx, id, updateFn)
	if err != nil {
		return err
	}
	c.mu.RUnlock()

	if item, exists := c.cache[id].(*entity.InventoryItem); exists {
		updated, err := updateFn(item)
		if err != nil {
			return err
		}
		if updated {
			c.mu.Lock()
			c.cache[id] = item
			c.mu.Unlock()
		}
	}
	return nil
}

func (c *CacheRepository) DeleteByID(ctx context.Context, id entity.UUID) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, exists := c.cache[id].(*entity.InventoryItem); exists {
		delete(c.cache, id)
	}

	return c.repo.DeleteByID(ctx, id)
}

func (c *CacheRepository) GetTotalCount(ctx context.Context) (int64, error) {
	return c.repo.GetTotalCount(ctx)
}
func (c *CacheRepository) GetAllInventoryItems(ctx context.Context, pagination *entity.Pagination) ([]*entity.InventoryItem, error) {
	return c.repo.GetAllInventoryItems(ctx, pagination)
}

func (c *CacheRepository) GetCategoryByID(ctx context.Context, id entity.UUID) (*entity.Category, error) {
	return c.repo.GetCategoryByID(ctx, id)
}
func (c *CacheRepository) SaveCategory(ctx context.Context, item entity.Category) error {
	return c.repo.SaveCategory(ctx, item)
}
func (c *CacheRepository) UpdateCategoryByID(ctx context.Context, id entity.UUID, updateFn func(*entity.Category) (bool, error)) error {
	return c.repo.UpdateCategoryByID(ctx, id, updateFn)
}
func (c *CacheRepository) DeleteCategoryByID(ctx context.Context, id entity.UUID) error {
	return c.repo.DeleteCategoryByID(ctx, id)
}
func (c *CacheRepository) GetTotalCategoriesCount(ctx context.Context) (int64, error) {
	return c.repo.GetTotalCategoriesCount(ctx)
}
func (c *CacheRepository) GetAllCategories(ctx context.Context, pagination *entity.Pagination) ([]*entity.Category, error) {
	return c.repo.GetAllCategories(ctx, pagination)
}
