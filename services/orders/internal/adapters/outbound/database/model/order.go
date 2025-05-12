package model

import (
	"database/sql"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/utils"
)

type Order struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID           string    `json:"id"`
	OrderID      string    `json:"order_id"`
	ProductName  string    `json:"product_name"`
	ProductPrice float64   `json:"product_price"`
	Quantity     int       `json:"quantity"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Payment struct {
	ID            string         `json:"id"`
	OrderID       string         `json:"order_id"`
	Amount        float64        `json:"amount"`
	Currency      string         `json:"currency"`
	PaymentMethod string         `json:"payment_method"`
	Status        string         `json:"status"`
	TransactionID sql.NullString `json:"transaction_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func OrderToModel(order *entity.Order) (*Order, error) {
	return &Order{
		ID:          order.ID.String(),
		UserID:      order.UserID.String(),
		UserName:    order.UserName,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}, nil
}

func ModelToOrder(m *Order, items []*OrderItem) (*entity.Order, error) {
	orderID, err := utils.ParseUUID(m.ID)
	if err != nil {
		return nil, err
	}

	userID, err := utils.ParseUUID(m.UserID)
	if err != nil {
		return nil, err
	}

	order := &entity.Order{
		ID:          orderID,
		UserID:      userID,
		UserName:    m.UserName,
		TotalAmount: m.TotalAmount,
		Status:      entity.OrderStatus(m.Status),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	if items != nil {
		for _, item := range items {
			order.Items = append(order.Items, entity.OrderItem{
				ProductID:    entity.UUID(item.ID),
				ProductName:  item.ProductName,
				ProductPrice: item.ProductPrice,
				Quantity:     int64(item.Quantity),
				CreatedAt:    item.CreatedAt,
				UpdatedAt:    item.UpdatedAt,
			})
		}

	}

	return order, nil
}

func OrderItemToModel(item *entity.OrderItem) (*OrderItem, error) {
	return &OrderItem{
		ID:           item.ProductID.String(),
		ProductName:  item.ProductName,
		ProductPrice: item.ProductPrice,
		Quantity:     int(item.Quantity),
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}, nil
}

func ModelToOrderItem(m *OrderItem) (*entity.OrderItem, *entity.UUID, error) {
	productID, err := utils.ParseUUID(m.ID)
	if err != nil {
		return nil, nil, err
	}

	orderID, err := utils.ParseUUID(m.OrderID)
	if err != nil {
		return nil, nil, err
	}

	return &entity.OrderItem{
		ProductID:    productID,
		ProductName:  m.ProductName,
		ProductPrice: m.ProductPrice,
		Quantity:     int64(m.Quantity),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, &orderID, nil
}