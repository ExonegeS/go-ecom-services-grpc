package model

import (
	"database/sql"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/domain"
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

func InventoryItemToModel(item *domain.Order) (*Product, error) {

	product := &Product{
		ID:        item.ID.String(),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	return product, nil
}

func ModelToInventoryItem(p *Product) (*domain.Order, error) {
	item := &domain.Order{
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}

	return item, nil
}
