// internal/core/domain/events.go
package domain

import (
	"time"
)

type OrderEvent struct {
	EventID   string
	Operation string
	OrderID   string
	UserID    string
	Items     []OrderItem
	Total     float64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrderItem struct {
	ProductID string
	Quantity  int32
	Price     float64
}

type InventoryEvent struct {
	EventID   string
	Operation string
	ProductID string
	Stock     int32
	Price     float64
	UpdatedAt time.Time
}
