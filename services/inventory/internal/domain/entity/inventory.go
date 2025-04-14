package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type InventoryItem struct {
	ID          UUID
	Name        string
	Description string
	Category    Category
	Price       float64
	Quantity    float64
	Unit        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Category struct {
	ID          UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UUID string

func NewUUID() UUID {
	return (UUID)(uuid.New().String())
}

func (u UUID) String() string {
	return string(u)
}

var (
	ErrInvalidRequestPayload = fmt.Errorf("invalid request payload")
	ErrNotImplemented        = fmt.Errorf("not implemented")
	ErrNotUpdated            = fmt.Errorf("not updated")
	ErrItemNotFound          = fmt.Errorf("item not found")
	ErrCategoryNotFound      = fmt.Errorf("category not found")
	ErrInvalidUUID           = fmt.Errorf("invalid UUID")
	ErrInvalidCategory       = fmt.Errorf("invalid category")
	ErrInvalidItem           = fmt.Errorf("invalid item")
	ErrInvalidPrice          = fmt.Errorf("invalid price")
	ErrInvalidQuantity       = fmt.Errorf("invalid quantity")
	ErrInvalidUnit           = fmt.Errorf("invalid unit")
)

func (i *InventoryItem) Validate() error {
	if i.Quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	if i.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	if i.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if i.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	if i.Unit == "" {
		return fmt.Errorf("unit cannot be empty")
	}
	return nil
}

func (c *Category) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if c.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	return nil
}
