package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/ports"
	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/utils"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	status "google.golang.org/grpc/status"
)

type Server struct {
	UnimplementedStatisticsServiceServer
	service ports.StatisticsService
	logger  *slog.Logger
}

func NewStatisticsServer(service ports.StatisticsService, logger *slog.Logger) *Server {
	return &Server{
		service: service,
		logger:  logger,
	}
}

func StartGRPCServer(grpcPort string, statisticsService ports.StatisticsService, logger *slog.Logger) error {
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	statisticsServer := NewStatisticsServer(statisticsService, logger)
	RegisterStatisticsServiceServer(grpcServer, statisticsServer)
	reflection.Register(grpcServer)

	logger.Info("gRPC server listening", "port", grpcPort)
	return grpcServer.Serve(lis)
}

func NewServer(service ports.StatisticsService) *Server {
	return &Server{service: service}
}

func (s *Server) GetUserOrdersStatistics(ctx context.Context, req *UserOrderStatisticsRequest) (*UserOrderStatisticsResponse, error) {
	s.logger.Info("Received GetUserOrdersStatistics gRPC request", "user_id", req.GetUserId())

	domainID := req.GetUserId()
	id, err := utils.ParseUUID(domainID)
	if err != nil {
		s.logger.Error("Failed to parse user id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	stats, err := s.service.GetUserOrderStats(ctx, id)
	if err != nil {
		return nil, err
	}
	if stats == nil {
		s.logger.Error("User order statistics not found", "user_id", req.GetUserId())
		return nil, status.Error(codes.NotFound, "user order statistics not found")
	}
	hourlyDistribution := make(map[string]int32, len(stats.HourlyDistribution))
	for hour, count := range stats.HourlyDistribution {
		hourlyDistribution[hour] = int32(count)
	}
	return &UserOrderStatisticsResponse{
		TotalOrders:        int32(stats.TotalOrders),
		HourlyDistribution: hourlyDistribution,
	}, nil
}

func (s *Server) GetUserStatistics(ctx context.Context, req *UserStatisticsRequest) (*UserStatisticsResponse, error) {
	s.logger.Info("Received GetUserStatistics gRPC request", "user_id", req.GetUserId())

	domainID := req.GetUserId()
	id, err := utils.ParseUUID(domainID)
	if err != nil {
		s.logger.Error("Failed to parse user id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	stats, err := s.service.GetUserStatistics(ctx, id)
	if err != nil {
		return nil, err
	}
	if stats == nil {
		s.logger.Error("User statistics not found", "user_id", req.GetUserId())
		return nil, status.Error(codes.NotFound, "user order statistics not found")
	}
	return &UserStatisticsResponse{
		TotalItemsPurchased:  int32(stats.TotalItemsPurchased),
		AverageOrderValue:    stats.AverageOrderValue,
		MostPurchasedItem:    stats.MostPurchasedItem,
		TotalCompletedOrders: int32(stats.TotalCompletedOrders),
	}, nil
}
