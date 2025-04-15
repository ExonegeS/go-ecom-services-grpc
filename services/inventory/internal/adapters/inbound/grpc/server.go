package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/utils"
)

type InventoryServer struct {
	UnimplementedInventoryServiceServer
	service application.InventoryService
	logger  *slog.Logger
}

func NewInventoryServer(service application.InventoryService, logger *slog.Logger) *InventoryServer {
	return &InventoryServer{
		service: service,
		logger:  logger,
	}
}

func StartGRPCServer(grpcPort string, invService application.InventoryService, logger *slog.Logger) error {
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	invServer := NewInventoryServer(invService, logger)
	RegisterInventoryServiceServer(grpcServer, invServer)
	reflection.Register(grpcServer)

	logger.Info("gRPC server listening", "port", grpcPort)
	return grpcServer.Serve(lis)
}

func (s *InventoryServer) GetProductByID(ctx context.Context, req *GetProductRequest) (*ProductResponse, error) {
	s.logger.Info("Received GetProductByID gRPC request", "id", req.GetId())
	domainID := req.GetId()
	id, err := utils.ParseUUID(domainID)
	if err != nil {
		return nil, err
	}

	product, err := s.service.GetInventoryItemByID(ctx, id)
	if err != nil {
		s.logger.Error("Error fetching product", "error", err.Error())
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	respProduct := &Product{
		Id:          product.ID.String(),
		Name:        product.Name,
		Description: product.Description,
		Category: &Category{
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

	return &ProductResponse{Product: respProduct}, nil
}

func ValidateCreateProductRequest(req *CreateProductRequest) error {
	if len(req.Name) == 0 {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if len(req.Description) == 0 {
		return status.Error(codes.InvalidArgument, "description cannot be empty")
	}
	if _, err := utils.ParseUUID(req.CategoryId); err != nil {
		return fmt.Errorf("invalid category ID: %w", err)
	}
	if req.Price <= 0 {
		return status.Error(codes.InvalidArgument, "price must be greater than zero")
	}
	if req.Quantity <= 0 {
		return status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}
	if len(req.Unit) == 0 {
		return status.Error(codes.InvalidArgument, "unit cannot be empty")
	}
	return nil
}

func (s *InventoryServer) CreateProduct(ctx context.Context, req *CreateProductRequest) (*ProductResponse, error) {
	s.logger.Info("Received CreateProduct gRPC request")
	if err := ValidateCreateProductRequest(req); err != nil {
		return nil, err
	}
	product := entity.InventoryItem{
		Name:        req.Name,
		Description: req.Description,
		Category: entity.Category{
			ID: entity.UUID(req.CategoryId),
		},
		Price:    req.Price,
		Quantity: req.Quantity,
		Unit:     req.Unit,
	}

	err := s.service.CreateInventoryItem(ctx, &product)
	if err != nil {
		return nil, err
	}

	respProduct := &Product{
		Id:          product.ID.String(),
		Name:        product.Name,
		Description: product.Description,
		Category: &Category{
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

	return &ProductResponse{Product: respProduct}, nil
}

func ValidateUpdateProductRequest(req *UpdateProductRequest) error {
	if req.Name == nil && req.Description == nil && req.CategoryId == nil &&
		req.Price == nil && req.Quantity == nil && req.Unit == nil {
		return status.Error(codes.InvalidArgument, "at least one field must be provided")
	}
	if req.Name != nil && len(*req.Name) == 0 {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if req.Description != nil && len(*req.Description) == 0 {
		return status.Error(codes.InvalidArgument, "description cannot be empty")
	}
	if req.CategoryId != nil {
		if _, err := utils.ParseUUID(*req.CategoryId); err != nil {
			return status.Error(codes.InvalidArgument, "invalid category ID format")
		}
	}
	if req.Price != nil && *req.Price <= 0 {
		return status.Error(codes.InvalidArgument, "price must be greater than zero")
	}
	if req.Quantity != nil && *req.Quantity <= 0 {
		return status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}
	if req.Unit != nil && len(*req.Unit) == 0 {
		return status.Error(codes.InvalidArgument, "unit cannot be empty")
	}
	return nil
}

func (s *InventoryServer) UpdateProduct(ctx context.Context, req *UpdateProductRequest) (*ProductResponse, error) {
	s.logger.Info("Received UpdateProduct gRPC request", "id", req.GetId())
	id, err := utils.ParseUUID(req.GetId())
	if err != nil {
		return nil, err
	}
	if err := ValidateUpdateProductRequest(req); err != nil {
		return nil, err
	}
	params := application.UpdateInventoryItemParams{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  (*entity.UUID)(req.CategoryId),
		Price:       req.Price,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
	}

	product, err := s.service.UpdateInventoryItem(ctx, id, params)
	if err != nil {
		return nil, err
	}

	respProduct := &Product{
		Id:          id.String(),
		Name:        product.Name,
		Description: product.Description,
		Category: &Category{
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

	return &ProductResponse{Product: respProduct}, nil
}

func (s *InventoryServer) DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*ProductResponse, error) {
	s.logger.Info("Received DeleteProduct gRPC request", "id", req.GetId())
	id, err := utils.ParseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	product, err := s.service.DeleteInventoryItem(ctx, id)
	if err != nil {
		return nil, err
	}

	respProduct := &Product{
		Id:          id.String(),
		Name:        product.Name,
		Description: product.Description,
		Category: &Category{
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

	return &ProductResponse{Product: respProduct}, nil
}

func (s *InventoryServer) ListProducts(ctx context.Context, req *ListProductsRequest) (*ListProductsResponse, error) {
	s.logger.Info("Received ListProducts gRPC request", "page", req.GetPage(), "page_size", req.GetPageSize(), "sort_by", req.GetSortBy())
	pagination := entity.NewPagination(int64(req.GetPage()), int64(req.GetPageSize()), entity.SortOption(req.GetSortBy()))
	paginatedData, err := s.service.GetPaginatedInventoryItems(ctx, pagination)
	if err != nil {
		s.logger.Error("Failed to get inventory items", "error", err.Error())
		return nil, fmt.Errorf("failed to get inventory items: %w", err)
	}

	grpcData := make([]*Product, 0)
	for _, product := range paginatedData.Data {
		grpcProduct := Product{
			Id:          product.ID.String(),
			Name:        product.Name,
			Description: product.Description,
			Category: &Category{
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
	resp := ListProductsResponse{
		CurrentPage: int32(paginatedData.CurrentPage),
		HasNextPage: paginatedData.HasNextPage,
		PageSize:    int32(paginatedData.PageSize),
		TotalPages:  int32(paginatedData.TotalPages),
		Products:    grpcData,
	}

	return &resp, nil
}

func (s *InventoryServer) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*CategoryResponse, error) {
	s.logger.Info("Received CreateCategory gRPC request")

	if err := ValidateCreateCategoryRequest(req); err != nil {
		return nil, err
	}

	category := entity.Category{
		Name:        req.GetName(),
		Description: req.GetDescription(),
	}

	err := s.service.CreateCategory(ctx, &category)
	if err != nil {
		s.logger.Error("Failed to create category", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to create category")
	}

	return &CategoryResponse{
		Category: convertDomainCategoryToPB(&category),
	}, nil
}

func (s *InventoryServer) GetCategoryByID(ctx context.Context, req *GetCategoryRequest) (*CategoryResponse, error) {
	s.logger.Info("Received GetCategoryByID gRPC request", "id", req.GetId())

	categoryID, err := utils.ParseUUID(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid category ID format")
	}

	category, err := s.service.GetCategoryByID(ctx, entity.UUID(categoryID))
	if err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, status.Error(codes.NotFound, "category not found")
		}
		s.logger.Error("Error fetching category", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to get category")
	}

	return &CategoryResponse{
		Category: convertDomainCategoryToPB(category),
	}, nil
}

func (s *InventoryServer) UpdateCategory(ctx context.Context, req *UpdateCategoryRequest) (*CategoryResponse, error) {
	s.logger.Info("Received UpdateCategory gRPC request", "id", req.GetId())

	if err := ValidateUpdateCategoryRequest(req); err != nil {
		return nil, err
	}

	categoryID, err := utils.ParseUUID(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid category ID format")
	}

	updateParams := application.UpdateCategoryParams{
		Name:        req.Name,
		Description: req.Description,
	}

	updatedCategory, err := s.service.UpdateCategory(ctx, entity.UUID(categoryID), updateParams)
	if err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, status.Error(codes.NotFound, "category not found")
		}
		s.logger.Error("Failed to update category", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to update category")
	}

	return &CategoryResponse{
		Category: convertDomainCategoryToPB(updatedCategory),
	}, nil
}

func (s *InventoryServer) DeleteCategory(ctx context.Context, req *DeleteCategoryRequest) (*CategoryResponse, error) {
	s.logger.Info("Received DeleteCategory gRPC request", "id", req.GetId())

	categoryID, err := utils.ParseUUID(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid category ID format")
	}

	deletedCategory, err := s.service.DeleteCategory(ctx, entity.UUID(categoryID))
	if err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, status.Error(codes.NotFound, "category not found")
		}
		s.logger.Error("Failed to delete category", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to delete category")
	}

	return &CategoryResponse{
		Category: convertDomainCategoryToPB(deletedCategory),
	}, nil
}

func (s *InventoryServer) ListCategories(ctx context.Context, req *ListCategoriesRequest) (*ListCategoriesResponse, error) {
	s.logger.Info("Received ListCategories gRPC request",
		"page", req.GetPage(),
		"page_size", req.GetPageSize(),
		"sort_by", req.GetSortBy(),
	)

	pagination := entity.NewPagination(
		int64(req.GetPage()),
		int64(req.GetPageSize()),
		entity.SortOption(req.GetSortBy()),
	)

	paginatedData, err := s.service.GetPaginatedCategories(ctx, pagination)
	if err != nil {
		s.logger.Error("Failed to list categories", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to list categories")
	}

	grpcCategories := make([]*Category, 0, len(paginatedData.Data))
	for _, category := range paginatedData.Data {
		grpcCategories = append(grpcCategories, convertDomainCategoryToPB(category))
	}

	return &ListCategoriesResponse{
		CurrentPage: int32(paginatedData.CurrentPage),
		HasNextPage: paginatedData.HasNextPage,
		PageSize:    int32(paginatedData.PageSize),
		TotalPages:  int32(paginatedData.TotalPages),
		Categories:  grpcCategories,
	}, nil
}

func ValidateCreateCategoryRequest(req *CreateCategoryRequest) error {
	if len(req.GetName()) == 0 {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if len(req.GetDescription()) == 0 {
		return status.Error(codes.InvalidArgument, "description cannot be empty")
	}
	return nil
}

func ValidateUpdateCategoryRequest(req *UpdateCategoryRequest) error {
	if req.Name == nil && req.Description == nil {
		return status.Error(codes.InvalidArgument, "at least one field must be provided")
	}
	if req.Name != nil && len(*req.Name) == 0 {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if req.Description != nil && len(*req.Description) == 0 {
		return status.Error(codes.InvalidArgument, "description cannot be empty")
	}
	return nil
}

func convertDomainCategoryToPB(category *entity.Category) *Category {
	return &Category{
		Id:          category.ID.String(),
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   category.UpdatedAt.Format(time.RFC3339),
	}
}
