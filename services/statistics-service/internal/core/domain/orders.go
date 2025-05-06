package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID          UUID
	UserID      UUID
	UserName    string
	TotalAmount float64
	Status      OrderStatus
	Items       []OrderItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

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
	ErrOrderNotFound         = fmt.Errorf("order not found")
	ErrItemNotFound          = fmt.Errorf("item not found")
	ErrInvalidUUID           = fmt.Errorf("invalid UUID")
	ErrInvalidQuantity       = fmt.Errorf("invalid quantity")
	ErrInsufficientQuantity  = fmt.Errorf("insufficient stored items")
)
