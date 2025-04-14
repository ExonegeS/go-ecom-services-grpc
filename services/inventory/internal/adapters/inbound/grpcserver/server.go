package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/grpc"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
)

type InventoryServer struct {
	pb.UnimplementedInventoryServiceServer
	service application.InventoryService
	logger  *slog.Logger
}

func NewInventoryServer(service application.InventoryService, logger *slog.Logger) *InventoryServer {
	return &InventoryServer{
		service: service,
		logger:  logger,
	}
}

func (s *InventoryServer) GetProductByID(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("Received GetProductByID gRPC request", "id", req.GetId())
	domainID := req.GetId()

	product, err := s.service.GetInventoryItemByID(ctx, entity.UUID(domainID))
	if err != nil {
		s.logger.Error("Error fetching product", "error", err.Error())
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	respProduct := &pb.Product{
		Id:          product.ID.String(),
		Name:        product.Name,
		Description: product.Description,
		Category: &pb.Category{
			Id:          string(product.Category.ID),
			Name:        product.Category.Name,
			Description: product.Category.Description,
			CreatedAt:   product.Category.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   product.Category.UpdatedAt.Format(time.RFC3339),
		},
		Price:     product.Price,
		Quantity:  product.Quantity,
		Unit:      product.Unit,
		CreatedAt: product.CreatedAt.Format(time.RFC3339),
		UpdatedAt: product.UpdatedAt.Format(time.RFC3339),
	}

	return &pb.ProductResponse{Product: respProduct}, nil
}

func (s *InventoryServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("Received CreateProduct gRPC request")
	return nil, fmt.Errorf("CreateProduct not implemented")
}

func (s *InventoryServer) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
	s.logger.Info("Received UpdateProduct gRPC request", "id", req.GetId())
	return nil, fmt.Errorf("UpdateProduct not implemented")
}

func (s *InventoryServer) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.Empty, error) {
	s.logger.Info("Received DeleteProduct gRPC request", "id", req.GetId())
	return nil, fmt.Errorf("DeleteProduct not implemented")
}

func (s *InventoryServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	s.logger.Info("Received ListProducts gRPC request", "page", req.GetPage(), "page_size", req.GetPageSize(), "sort_by", req.GetSortBy())
	pagination := entity.NewPagination(int64(req.GetPage()), int64(req.GetPageSize()), entity.SortOption(req.GetSortBy()))
	paginatedData, err := s.service.GetPaginatedInventoryItems(ctx, pagination)
	if err != nil {
		s.logger.Error("Failed to get inventory items", "error", err.Error())
		return nil, fmt.Errorf("failed to get inventory items: %w", err)
	}

	grpcData := make([]*pb.Product, 0)
	for _, product := range paginatedData.Data {
		grpcProduct := pb.Product{
			Id:          product.ID.String(),
			Name:        product.Name,
			Description: product.Description,
			Category: &pb.Category{
				Id:          string(product.Category.ID),
				Name:        product.Category.Name,
				Description: product.Category.Description,
				CreatedAt:   product.Category.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   product.Category.UpdatedAt.Format(time.RFC3339),
			},
			Price:     product.Price,
			Quantity:  product.Quantity,
			Unit:      product.Unit,
			CreatedAt: product.CreatedAt.Format(time.RFC3339),
			UpdatedAt: product.UpdatedAt.Format(time.RFC3339),
		}
		grpcData = append(grpcData, &grpcProduct)
	}
	resp := pb.ListProductsResponse{
		CurrentPage: int32(paginatedData.CurrentPage),
		HasNextPage: paginatedData.HasNextPage,
		PageSize:    int32(paginatedData.PageSize),
		TotalPages:  int32(paginatedData.TotalPages),
		Products:    grpcData,
	}

	return &resp, nil
}

func StartGRPCServer(grpcPort string, invService application.InventoryService, logger *slog.Logger) error {
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	invServer := NewInventoryServer(invService, logger)
	pb.RegisterInventoryServiceServer(grpcServer, invServer)
	reflection.Register(grpcServer)

	logger.Info("gRPC server listening", "port", grpcPort)
	return grpcServer.Serve(lis)
}
