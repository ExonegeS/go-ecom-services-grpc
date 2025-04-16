package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

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
		s.logger.Error("Failed to parse product ID", "error", err)
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	product, err := s.service.GetInventoryItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		s.logger.Error("Failed to get inventory item", "error", err)
		return nil, status.Error(codes.Internal, "failed to retrieve product")
	}

	respProduct := &Product{
		Id:          product.ID.String(),
		Name:        product.Name,
		Description: product.Description,
		Category: &Category{
			Id:          string(product.Category.ID),
			Name:        product.Category.Name,
			Description: product.Category.Description,
			CreatedAt:   timestamppb.New(product.Category.CreatedAt),
			UpdatedAt:   timestamppb.New(product.Category.UpdatedAt),
		},
		Price:     product.Price,
		Quantity:  product.Quantity,
		Unit:      product.Unit,
		CreatedAt: timestamppb.New(product.CreatedAt),
		UpdatedAt: timestamppb.New(product.UpdatedAt),
	}

	return &ProductResponse{Product: respProduct}, nil
}

func (s *InventoryServer) CreateProduct(ctx context.Context, req *CreateProductRequest) (*ProductResponse, error) {
	s.logger.Info("Received CreateProduct gRPC request")

	if err := ValidateCreateProductRequest(req); err != nil {
		s.logger.Error("Invalid create request", "error", err)
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
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, status.Error(codes.InvalidArgument, "specified category does not exist")
		}
		s.logger.Error("Failed to create product", "error", err)
		return nil, status.Error(codes.Internal, "failed to create product")
	}

	respProduct := &Product{
		Id:          product.ID.String(),
		Name:        product.Name,
		Description: product.Description,
		Category: &Category{
			Id:          string(product.Category.ID),
			Name:        product.Category.Name,
			Description: product.Category.Description,
			CreatedAt:   timestamppb.New(product.Category.CreatedAt),
			UpdatedAt:   timestamppb.New(product.Category.UpdatedAt),
		},
		Price:     product.Price,
		Quantity:  product.Quantity,
		Unit:      product.Unit,
		CreatedAt: timestamppb.New(product.CreatedAt),
		UpdatedAt: timestamppb.New(product.UpdatedAt),
	}

	return &ProductResponse{
		Product: respProduct,
	}, nil
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
		s.logger.Error("Invalid product ID", "error", err)
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	if err := ValidateUpdateProductRequest(req); err != nil {
		s.logger.Error("Invalid update request", "error", err)
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
		if errors.Is(err, entity.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		s.logger.Error("Failed to update product", "error", err)
		return nil, status.Error(codes.Internal, "failed to update product")
	}

	respProduct := &Product{
		Id:          id.String(),
		Name:        product.Name,
		Description: product.Description,
		Category: &Category{
			Id:          string(product.Category.ID),
			Name:        product.Category.Name,
			Description: product.Category.Description,
			CreatedAt:   timestamppb.New(product.Category.CreatedAt),
			UpdatedAt:   timestamppb.New(product.Category.UpdatedAt),
		},
		Price:     product.Price,
		Quantity:  product.Quantity,
		Unit:      product.Unit,
		CreatedAt: timestamppb.New(product.CreatedAt),
		UpdatedAt: timestamppb.New(product.UpdatedAt),
	}

	return &ProductResponse{Product: respProduct}, nil
}

func (s *InventoryServer) DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*ProductResponse, error) {
	s.logger.Info("Received DeleteProduct gRPC request", "id", req.GetId())

	id, err := utils.ParseUUID(req.GetId())
	if err != nil {
		s.logger.Error("Invalid product ID", "error", err)
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	product, err := s.service.DeleteInventoryItem(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		s.logger.Error("Failed to delete product", "error", err)
		return nil, status.Error(codes.Internal, "failed to delete product")
	}

	respProduct := &Product{
		Id:          id.String(),
		Name:        product.Name,
		Description: product.Description,
		Category: &Category{
			Id:          string(product.Category.ID),
			Name:        product.Category.Name,
			Description: product.Category.Description,
			CreatedAt:   timestamppb.New(product.Category.CreatedAt),
			UpdatedAt:   timestamppb.New(product.Category.UpdatedAt),
		},
		Price:     product.Price,
		Quantity:  product.Quantity,
		Unit:      product.Unit,
		CreatedAt: timestamppb.New(product.CreatedAt),
		UpdatedAt: timestamppb.New(product.UpdatedAt),
	}

	return &ProductResponse{Product: respProduct}, nil
}

func (s *InventoryServer) ListProducts(ctx context.Context, req *ListProductsRequest) (*ListProductsResponse, error) {
	s.logger.Info("Received ListProducts gRPC request",
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

	paginatedData, err := s.service.GetPaginatedInventoryItems(ctx, pagination)
	if err != nil {
		s.logger.Error("Failed to list products", "error", err)
		return nil, status.Error(codes.Internal, "failed to list products")
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
				CreatedAt:   timestamppb.New(product.Category.CreatedAt),
				UpdatedAt:   timestamppb.New(product.Category.UpdatedAt),
			},
			Price:     product.Price,
			Quantity:  product.Quantity,
			Unit:      product.Unit,
			CreatedAt: timestamppb.New(product.CreatedAt),
			UpdatedAt: timestamppb.New(product.UpdatedAt),
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
		s.logger.Error("Failed to validate category", "error", err.Error())
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
		s.logger.Error("Failed to parse category id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid category ID format")
	}

	category, err := s.service.GetCategoryByID(ctx, entity.UUID(categoryID))
	if err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, status.Error(codes.NotFound, "category not found")
		}
		s.logger.Error("Failed to GetCategoryByID", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to get category")
	}

	return &CategoryResponse{
		Category: convertDomainCategoryToPB(category),
	}, nil
}

func (s *InventoryServer) UpdateCategory(ctx context.Context, req *UpdateCategoryRequest) (*CategoryResponse, error) {
	s.logger.Info("Received UpdateCategory gRPC request", "id", req.GetId())

	if err := ValidateUpdateCategoryRequest(req); err != nil {
		s.logger.Error("Failed to ValidateUpdateCategoryRequest", "error", err.Error())
		return nil, err
	}

	categoryID, err := utils.ParseUUID(req.GetId())
	if err != nil {
		s.logger.Error("Failed to parse category id", "error", err.Error())
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
		s.logger.Error("Failed to UpdateCategory", "error", err.Error())
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
		s.logger.Error("Failed to parse category id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid category ID format")
	}

	deletedCategory, err := s.service.DeleteCategory(ctx, entity.UUID(categoryID))
	if err != nil {
		if errors.Is(err, entity.ErrCategoryNotFound) {
			return nil, status.Error(codes.NotFound, "category not found")
		}
		s.logger.Error("Failed to DeleteCategory", "error", err.Error())
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

	paginatedData, err := s.service.GetPaginatedCategories(ctx, pagination)
	if err != nil {
		s.logger.Error("Failed to GetPaginatedCategories", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to list categories")
	}

	grpcData := make([]*Category, 0, len(paginatedData.Data))
	for _, category := range paginatedData.Data {
		grpcData = append(grpcData, convertDomainCategoryToPB(category))
	}

	return &ListCategoriesResponse{
		CurrentPage: int32(paginatedData.CurrentPage),
		HasNextPage: paginatedData.HasNextPage,
		PageSize:    int32(paginatedData.PageSize),
		TotalPages:  int32(paginatedData.TotalPages),
		Categories:  grpcData,
	}, nil
}

func (s *InventoryServer) ReserveProducts(ctx context.Context, req *ReserveProductRequest) (*Empty, error) {
	s.logger.Info("Received ReserveProducts gRPC request", "id", req.GetId())

	if err := ValidateReserveProductRequest(req); err != nil {
		s.logger.Error("Failed to ValidateReserveProductRequest", "error", err.Error())
		return nil, err
	}

	err := s.service.ReserveProduct(ctx, entity.UUID(req.GetId()), int64(req.GetQuantity()))
	if err != nil {
		if errors.Is(err, entity.ErrItemNotFound) {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		s.logger.Error("Failed to UpdateCategory", "error", err.Error())
		return nil, status.Error(codes.Internal, "failed to reserve item")
	}

	return nil, nil
}

func ValidateCreateProductRequest(req *CreateProductRequest) error {
	if len(req.Name) == 0 {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if len(req.Description) == 0 {
		return status.Error(codes.InvalidArgument, "description cannot be empty")
	}
	if _, err := utils.ParseUUID(req.GetCategoryId()); err != nil {
		return status.Error(codes.InvalidArgument, "invalid category ID format")
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

func ValidateReserveProductRequest(req *ReserveProductRequest) error {
	if _, err := utils.ParseUUID(req.GetId()); err != nil {
		return status.Error(codes.InvalidArgument, "invalid category ID format")
	}
	if req.Quantity <= 0 {
		return status.Error(codes.InvalidArgument, "quantity must be greater than zero")
	}
	return nil
}

func convertDomainCategoryToPB(category *entity.Category) *Category {
	return &Category{
		Id:          category.ID.String(),
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   timestamppb.New(category.CreatedAt),
		UpdatedAt:   timestamppb.New(category.UpdatedAt),
	}
}
