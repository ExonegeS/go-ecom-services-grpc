package grpc

import (
	context "context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/utils"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	status "google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type OrdersServer struct {
	UnimplementedOrdersServiceServer
	service application.OrdersService
	logger  *slog.Logger
}

func NewOrdersServer(service application.OrdersService, logger *slog.Logger) *OrdersServer {
	return &OrdersServer{
		service: service,
		logger:  logger,
	}
}

func StartGRPCServer(grpcPort string, orderService application.OrdersService, logger *slog.Logger) error {
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	orderServer := NewOrdersServer(orderService, logger)
	RegisterOrdersServiceServer(grpcServer, orderServer)
	reflection.Register(grpcServer)

	logger.Info("gRPC server listening", "port", grpcPort)
	return grpcServer.Serve(lis)
}

func (s *OrdersServer) GetOrderByID(ctx context.Context, req *GetOrderRequest) (*OrderResponse, error) {
	s.logger.Info("Received GetOrderByID gRPC request", "id", req.GetId())
	domainID := req.GetId()

	id, err := utils.ParseUUID(domainID)
	if err != nil {
		s.logger.Error("Failed to parse order id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid order ID format")
	}

	order, err := s.service.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		s.logger.Error("Error fetching order", "error", err.Error())
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	items := make([]*Item, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items,
			&Item{
				ProductId:   item.ProductID.String(),
				ProductName: item.ProductName,
				UnitPrice:   item.ProductPrice,
				Quantity:    int32(item.Quantity),
				CreatedAt:   timestamppb.New(item.CreatedAt),
				UpdatedAt:   timestamppb.New(item.UpdatedAt),
			},
		)
	}

	respOrder := &Order{
		Id:          order.ID.String(),
		UserId:      order.UserID.String(),
		UserName:    order.UserName,
		TotalAmount: order.TotalAmount,
		Status:      OrderStatusToPB(order.Status),
		Items:       items,
		CreatedAt:   timestamppb.New(order.CreatedAt),
		UpdatedAt:   timestamppb.New(order.UpdatedAt),
	}

	return &OrderResponse{Order: respOrder}, nil
}

func ValidateCreateOrderRequest(req *CreateOrderRequest) error {
	if _, err := utils.ParseUUID(req.UserId); err != nil {
		return status.Error(codes.InvalidArgument, "invalid user ID format")
	}
	if len(req.Name) == 0 {
		return status.Error(codes.InvalidArgument, "'name' cannot be empty")
	}
	if len(req.Items) == 0 {
		return status.Error(codes.InvalidArgument, "order 'items' cannot be empty")
	}
	return nil
}

func (s *OrdersServer) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*OrderResponse, error) {
	s.logger.Info("Received CreateOrder gRPC request")

	if err := ValidateCreateOrderRequest(req); err != nil {
		s.logger.Error("Invalid create request", slog.String("error", err.Error()))
		return nil, err
	}

	order := entity.Order{
		UserID:   entity.UUID(req.UserId),
		UserName: req.Name,
		Status:   entity.OrderStatusPending,
		Items:    []entity.OrderItem{},
	}

	for _, item := range req.GetItems() {
		order.Items = append(order.Items, entity.OrderItem{
			ProductID: entity.UUID(item.GetProductId()),
			Quantity:  int64(item.GetQuantity()),
		})
	}

	err := s.service.CreateOrder(ctx, &order)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidUUID) {
			return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
		}
		if errors.Is(err, entity.ErrInvalidQuantity) {
			return nil, status.Error(codes.InvalidArgument, "invalid item quantity")
		}
		if errors.Is(err, entity.ErrInsufficientQuantity) {
			return nil, status.Error(codes.NotFound, "insufficient item quantity in storage")
		}
		if errors.Is(err, entity.ErrItemNotFound) {
			return nil, status.Error(codes.InvalidArgument, "specified item not exist")
		}
		s.logger.Error("Failed to create order", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	respOrder := &Order{
		Id: order.ID.String(),
	}

	return &OrderResponse{
		Order: respOrder,
	}, nil
}

func (s *OrdersServer) UpdateOrder(ctx context.Context, req *UpdateOrderRequest) (*OrderResponse, error) {
	s.logger.Info("Received UpdateOrder gRPC request", "id", req.GetId())

	domainID, err := utils.ParseUUID(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order ID format")
	}

	var statusVal entity.OrderStatus
	switch req.GetStatus() {
	case "pending":
		statusVal = entity.OrderStatusPending
	case "processing":
		statusVal = entity.OrderStatusProcessing
	case "completed":
		statusVal = entity.OrderStatusCompleted
	case "cancelled":
		statusVal = entity.OrderStatusCancelled
	case "refunded":
		statusVal = entity.OrderStatusRefunded
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid order status")
	}

	params := application.UpdateOrderParams{
		Status: &statusVal,
	}
	if req.GetUserName() != "" {
		params.UserName = &req.UserName
	}

	updatedOrder, err := s.service.UpdateOrder(ctx, domainID, params)
	if err != nil {
		if errors.Is(err, entity.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		s.logger.Error("Failed to update order", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to update order")
	}

	return &OrderResponse{
		Order: convertOrderToProto(updatedOrder),
	}, nil
}

func (s *OrdersServer) DeleteOrder(ctx context.Context, req *DeleteOrderRequest) (*OrderResponse, error) {
	s.logger.Info("Received DeleteOrder gRPC request", "id", req.GetId())

	domainID, err := utils.ParseUUID(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order ID format")
	}

	deletedOrder, err := s.service.DeleteOrder(ctx, domainID)
	if err != nil {
		if errors.Is(err, entity.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		s.logger.Error("Failed to delete order", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to delete order")
	}

	return &OrderResponse{
		Order: convertOrderToProto(deletedOrder),
	}, nil
}

func (s *OrdersServer) ListOrders(ctx context.Context, req *ListOrdersRequest) (*ListOrdersResponse, error) {
	s.logger.Info("Received ListOrders gRPC request",
		"page", req.GetPage(),
		"page_size", req.GetPageSize(),
		"sort_by", req.GetSortBy(),
	)

	sortBy, err := entity.ParseSortOption(req.GetSortBy())
	if err != nil {
		s.logger.Error("Invalid sort option", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pagination := entity.NewPagination(
		int64(req.GetPage()),
		int64(req.GetPageSize()),
		sortBy,
	)

	result, err := s.service.GetPaginatedOrders(ctx, pagination)
	if err != nil {
		s.logger.Error("Failed to list orders", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to list orders")
	}

	response := &ListOrdersResponse{
		CurrentPage: int32(result.CurrentPage),
		HasNextPage: result.HasNextPage,
		PageSize:    int32(result.PageSize),
		TotalPages:  int32(result.TotalPages),
		Orders:      make([]*Order, 0, len(result.Data)),
	}

	for _, order := range result.Data {
		response.Orders = append(response.Orders, convertOrderToProto(order))
	}

	return response, nil
}

var orderStatusMap = map[entity.OrderStatus]OrderStatus{
	entity.OrderStatusPending:    OrderStatus_ORDER_STATUS_PENDING,
	entity.OrderStatusProcessing: OrderStatus_ORDER_STATUS_PROCESSING,
	entity.OrderStatusCompleted:  OrderStatus_ORDER_STATUS_COMPLETED,
	entity.OrderStatusCancelled:  OrderStatus_ORDER_STATUS_CANCELLED,
	entity.OrderStatusRefunded:   OrderStatus_ORDER_STATUS_REFUNDED,
}

func OrderStatusToPB(s entity.OrderStatus) OrderStatus {
	if v, ok := orderStatusMap[s]; ok {
		return v
	}
	return OrderStatus_ORDER_STATUS_PENDING
}

func convertOrderToProto(order *entity.Order) *Order {
	protoOrder := &Order{
		Id:          order.ID.String(),
		UserId:      order.UserID.String(),
		UserName:    order.UserName,
		TotalAmount: order.TotalAmount,
		Status:      OrderStatusToPB(order.Status),
		CreatedAt:   timestamppb.New(order.CreatedAt),
		UpdatedAt:   timestamppb.New(order.UpdatedAt),
		Items:       make([]*Item, 0, len(order.Items)),
	}

	for _, item := range order.Items {
		protoOrder.Items = append(protoOrder.Items, &Item{
			ProductId:   item.ProductID.String(),
			ProductName: item.ProductName,
			UnitPrice:   item.ProductPrice,
			Quantity:    int32(item.Quantity),
			CreatedAt:   timestamppb.New(item.CreatedAt),
			UpdatedAt:   timestamppb.New(item.UpdatedAt),
		})
	}

	return protoOrder
}
