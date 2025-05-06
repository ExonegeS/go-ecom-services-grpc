package model

import (
	"database/sql"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
)

type Product struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	CategoryID  sql.NullString `json:"category_id"`
	Price       float64        `json:"price"`
	Stock       float64        `json:"stock_quantity"`
	Unit        string         `json:"unit"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func InventoryItemToModel(item *entity.InventoryItem) (*Product, *Category, error) {
	category := &Category{
		ID:          item.Category.ID.String(),
		Name:        item.Category.Name,
		Description: item.Category.Description,
		CreatedAt:   item.Category.CreatedAt,
		UpdatedAt:   item.Category.UpdatedAt,
	}

	product := &Product{
		ID:          item.ID.String(),
		Name:        item.Name,
		Description: item.Description,
		CategoryID:  sql.NullString{String: item.Category.ID.String(), Valid: true},
		Price:       item.Price,
		Stock:       item.Quantity,
		Unit:        item.Unit,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	return product, category, nil
}

func ModelToInventoryItem(p *Product, c *Category) (*entity.InventoryItem, error) {
	item := &entity.InventoryItem{
		ID:          entity.UUID(p.ID),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Quantity:    p.Stock,
		Unit:        p.Unit,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if c != nil {
		item.Category = entity.Category{
			ID:          entity.UUID(c.ID),
			Name:        c.Name,
			Description: c.Description,
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		}
	}

	return item, nil
}
