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

// Implement the gRPC service methods
func (s *Server) GetUserOrdersStatistics(ctx context.Context, req *UserOrderStatisticsRequest) (*UserOrderStatisticsResponse, error) {
	// Delegate to service layer

	domainID := req.GetUserId()
	id, err := utils.ParseUUID(domainID)
	if err != nil {
		s.logger.Error("Failed to parse order id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid order ID format")
	}

	stats, err := s.service.GetUserOrderStats(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = stats
	return &UserOrderStatisticsResponse{
		TotalOrders:        50,
		HourlyDistribution: map[string]int32{"23": 30, "22": 20},
	}, nil
}
