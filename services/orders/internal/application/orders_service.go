package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpc "github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/adapters/outbound/grpc/statistics"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/adapters/outbound/nats"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrdersService interface {
	GetOrderByID(ctx context.Context, id entity.UUID) (*entity.Order, error)
	CreateOrder(ctx context.Context, order *entity.Order) error
	UpdateOrder(ctx context.Context, id entity.UUID, params UpdateOrderParams) (*entity.Order, error)
	DeleteOrder(ctx context.Context, id entity.UUID) (*entity.Order, error)
	GetPaginatedOrders(ctx context.Context, pagination *entity.Pagination) (*entity.PaginationResponse[*entity.Order], error)
}

// services/orders/internal/application/orders_service.go
type UpdateOrderParams struct {
	UserName *string
	Status   *entity.OrderStatus
}

type ordersService struct {
	ordersRepo      ports.OrdersRepository
	inventoryClient ports.InventoryService
	publisher       *nats.OrderEventPublisher
	timeSource      func() time.Time
	logger          *slog.Logger
}

func NewOrdersService(ordersRepo ports.OrdersRepository, inventoryClient ports.InventoryService, timeSource func() time.Time, publisher *nats.OrderEventPublisher, logger *slog.Logger) OrdersService {
	return &ordersService{
		ordersRepo:      ordersRepo,
		inventoryClient: inventoryClient,
		timeSource:      timeSource,
		publisher:       publisher,
		logger:          logger,
	}
}

func (s *ordersService) GetOrderByID(ctx context.Context, id entity.UUID) (*entity.Order, error) {
	return s.ordersRepo.GetOrderByID(ctx, id)
}

func (s *ordersService) CreateOrder(ctx context.Context, order *entity.Order) error {
	if order == nil {
		return entity.ErrInvalidRequestPayload
	}
	for _, item := range order.Items {
		if item.Quantity <= 0 {
			return entity.ErrInvalidQuantity
		}
		product, err := s.inventoryClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			status, ok := status.FromError(err)
			if !ok {
				return err
			}

			switch status.Code() {
			case codes.NotFound:
				return entity.ErrItemNotFound
			default:
				return err
			}
		}
		if item.Quantity > product.Quantity {
			return entity.ErrInsufficientQuantity
		}
		ok, err := s.inventoryClient.ReserveItem(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrInsufficientQuantity
		}
	}
	return entity.ErrNotImplemented
}

func (s *ordersService) UpdateOrder(ctx context.Context, id entity.UUID, params UpdateOrderParams) (*entity.Order, error) {
	order, err := s.ordersRepo.GetOrderByID(ctx, id)
	if err != nil {
		// return nil, fmt.Errorf("failed to get order: %w", err)
		order = &entity.Order{
			ID:     "007ab384-de75-4b56-8ade-890d3408d884",
			UserID: "007ab384-de75-4b56-8ade-890d3408d884",
			Status: "processing",
		}
	}

	// Apply updates
	if params.UserName != nil {
		order.UserName = *params.UserName
	}
	if params.Status != nil {
		order.Status = *params.Status
	}

	// Save updated order
	// updatedOrder, err := s.ordersRepo.UpdateOrderByID(ctx, order.ID, order)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to update order: %w", err)
	// }

	// Publish update event
	event := &grpc.OrderEvent{
		EventId:   uuid.New().String(),
		Operation: "updated",
		OrderId:   order.ID.String(),
		UserId:    order.UserID.String(),
		Status:    string(order.Status),
		CreatedAt: timestamppb.New(order.CreatedAt),
	}
	if err := s.publisher.PublishOrderUpdated(ctx, event); err != nil {
		s.logger.Error("failed to publish order update event", "error", err)
	}

	return order, nil
}

func (s *ordersService) DeleteOrder(ctx context.Context, id entity.UUID) (*entity.Order, error) {
	order, err := s.ordersRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if err := s.ordersRepo.DeleteOrderByID(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to delete order: %w", err)
	}

	// Publish delete event
	event := &grpc.OrderEvent{
		EventId:   uuid.New().String(),
		Operation: "deleted",
		OrderId:   order.ID.String(),
		UserId:    order.UserID.String(),
		CreatedAt: timestamppb.New(order.CreatedAt),
	}
	if err := s.publisher.PublishOrderDeleted(ctx, event); err != nil {
		s.logger.Error("failed to publish order delete event", "error", err)
	}

	return order, nil
}

func (s *ordersService) GetPaginatedOrders(ctx context.Context, pagination *entity.Pagination) (*entity.PaginationResponse[*entity.Order], error) {
	const op = "inventoryService.GetPaginatedInventoryItems"

	totalItems, err := s.ordersRepo.GetTotalOrdersCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	totalPages := (totalItems + pagination.PageSize - 1) / pagination.PageSize

	items, err := s.ordersRepo.GetAllOrders(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.PaginationResponse[*entity.Order]{
		CurrentPage: pagination.Page,
		HasNextPage: pagination.Page < totalPages,
		PageSize:    pagination.PageSize,
		TotalPages:  totalPages,
		Data:        items,
	}, nil

}
