package app

import (
	"database/sql"
	"log"
	"log/slog"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/config"
	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/adapters/grpc"
	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/adapters/nats"
	database "github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/adapters/repository"
	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/service"
	_ "github.com/lib/pq"
)

type APIServer struct {
	cfg    *config.Config
	logger *slog.Logger
}

func NewAPIServer(config *config.Config, logger *slog.Logger) *APIServer {
	return &APIServer{
		config,
		logger,
	}
}

func (s *APIServer) Run() error {
	connStr := s.cfg.Database.MakeConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("could not open database: %v", err)
	}

	repo := database.NewPostgresStatisticsRepository(db)

	service := service.NewStatisticsService(repo, s.logger)

	sub, err := nats.NewSubscriber(s.cfg.NATS, service, s.logger)
	if err != nil {
		log.Fatalf("could not register subscriber: %v", err)
	}
	err = sub.Subscribe()
	if err != nil {
		log.Fatalf("failed to subscribe to NATS: %w", err)
	}

	return grpc.StartGRPCServer(s.cfg.Server.Port, service, s.logger)
}
