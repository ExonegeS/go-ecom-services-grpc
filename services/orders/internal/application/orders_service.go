package application

import (
	"context"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrdersService interface {
	GetOrderByID(ctx context.Context, id entity.UUID) (*entity.Order, error)
	CreateOrder(ctx context.Context, order *entity.Order) error
	UpdateOrder(ctx context.Context, id entity.UUID, params UpdateOrderParams) (*entity.Order, error)
	DeleteOrder(ctx context.Context, id entity.UUID) (*entity.Order, error)
	GetPaginatedOrders(ctx context.Context, pagination *entity.Pagination) (*entity.PaginationResponse[*entity.Order], error)
}

type UpdateOrderParams struct {
	Name        *string
	Description *string
	CategoryID  *entity.UUID
	Price       *float64
	Quantity    *float64
	Unit        *string
}

type ordersService struct {
	ordersRepo      ports.OrdersRepository
	inventoryClient ports.InventoryService
	timeSource      func() time.Time
}

func NewOrdersService(ordersRepo ports.OrdersRepository, inventoryClient ports.InventoryService, timeSource func() time.Time) OrdersService {
	return &ordersService{
		ordersRepo:      ordersRepo,
		inventoryClient: inventoryClient,
		timeSource:      timeSource,
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
	return nil, entity.ErrNotImplemented
}

func (s *ordersService) DeleteOrder(ctx context.Context, id entity.UUID) (*entity.Order, error) {
	return nil, entity.ErrNotImplemented
}

func (s *ordersService) GetPaginatedOrders(ctx context.Context, pagination *entity.Pagination) (*entity.PaginationResponse[*entity.Order], error) {
	return nil, entity.ErrNotImplemented
}
