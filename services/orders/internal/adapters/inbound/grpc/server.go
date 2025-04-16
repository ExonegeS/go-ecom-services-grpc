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
	s.logger.Info("Received GetProductByID gRPC request", "id", req.GetId())
	domainID := req.GetId()
	id, err := utils.ParseUUID(domainID)
	if err != nil {
		s.logger.Error("Failed to parse order id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid order ID format")
	}

	product, err := s.service.GetOrderByID(ctx, id)
	if err != nil {
		s.logger.Error("Error fetching product", "error", err.Error())
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	respOrder := &Order{
		Id:        product.ID.String(),
		CreatedAt: timestamppb.New(product.CreatedAt),
		UpdatedAt: timestamppb.New(product.UpdatedAt),
	}

	return &OrderResponse{Order: respOrder}, nil
}

func ValidateCreateOrderRequest(req *CreateOrderRequest) error {
	if _, err := utils.ParseUUID(req.UserId); err != nil {
		return status.Error(codes.InvalidArgument, "invalid user ID format")
	}
	if len(req.Name) == 0 {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if len(req.Items) == 0 {
		return status.Error(codes.InvalidArgument, "order items cannot be empty")
	}
	return nil
}

func (s *OrdersServer) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*OrderResponse, error) {
	s.logger.Info("Received CreateOrder gRPC request")

	if err := ValidateCreateOrderRequest(req); err != nil {
		s.logger.Error("Invalid create request", "error", err)
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
		if errors.Is(err, entity.ErrInvalidQuantity) {
			return nil, status.Error(codes.InvalidArgument, "invalid item quantity")
		}
		if errors.Is(err, entity.ErrInsufficientQuantity) {
			return nil, status.Error(codes.NotFound, "insufficient item quantity in storage")
		}
		if errors.Is(err, entity.ErrItemNotFound) {
			return nil, status.Error(codes.InvalidArgument, "specified item not exist")
		}
		s.logger.Error("Failed to create order", "error", err)
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	respOrder := &Order{
		Id: order.ID.String(),
	}

	return &OrderResponse{
		Order: respOrder,
	}, nil
}
