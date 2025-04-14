package model

import (
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
)

type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func CategoryToModel(c *entity.Category) *Category {
	return &Category{
		ID:          string(c.ID),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func ModelToCategory(m *Category) (*entity.Category, error) {
	return &entity.Category{
		ID:          entity.UUID(m.ID),
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}
