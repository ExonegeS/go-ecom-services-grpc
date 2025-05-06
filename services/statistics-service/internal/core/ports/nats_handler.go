package ports

import "github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/domain"

type EventHandler interface {
	HandleOrderCreated(order domain.Order) error
	HandleInventoryUpdated(inventory domain.OrderItem) error
}
