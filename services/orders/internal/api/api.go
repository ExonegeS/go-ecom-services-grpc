package api

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	grpc "github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/adapters/inbound/grpc"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/adapters/outbound/database"
	inventory "github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/adapters/outbound/grpc/inventory"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/adapters/outbound/nats"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/config"
)

type APIServer struct {
	cfg    config.Config
	db     *sql.DB
	logger *slog.Logger
	mux    *http.ServeMux
}

func NewAPIServer(cfg config.Config, db *sql.DB, logger *slog.Logger) *APIServer {
	return &APIServer{
		cfg:    cfg,
		db:     db,
		logger: logger,
		mux:    http.NewServeMux(),
	}
}

func (s *APIServer) Run() error {
	orderRepo := database.NewPostgresOrdersRepository(s.db)

	inventoryAddr := fmt.Sprintf("%s:%s", s.cfg.Clients["inventory client"].Address, s.cfg.Clients["inventory client"].GRPCPort)
	inventoryClient, err := inventory.NewInventoryClient(inventoryAddr)
	if err != nil {
		return err
	}

	publisher, err := nats.NewOrderEventPublisher(s.cfg.NATS.URL)
	if err != nil {
		log.Fatalf("Failed to create NATS publisher: %v", err)
	}
	defer publisher.Close()

	orderService := application.NewOrdersService(orderRepo, inventoryClient, time.Now, publisher, s.logger)

	return grpc.StartGRPCServer(s.cfg.Server.GRPCPort, orderService, s.logger)
}
