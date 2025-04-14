package api

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/grpcserver"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/rest"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/rest/middleware"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/outbound/database"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/config"

	"log/slog"

	"google.golang.org/grpc"
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
	invRepo := database.NewPostgresInventoryRepository(s.db)

	invService := application.NewInventoryService(invRepo, time.Now)

	invHandler := rest.NewInventoryHandler(invService, s.logger)
	healthHandler := rest.NewHealthHandler(s.logger, s.db)

	invHandler.RegisterEndpoints(s.mux, s.cfg)
	healthHandler.RegisterEndpoints(s.mux, s.cfg)

	loggerMW := middleware.NewLoggerMW(s.logger)
	_ = loggerMW
	MWChain := middleware.NewMiddlewareChain()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.cfg.Server.GRPCPort))
		if err != nil {
			s.logger.Error("failed to listen for gRPC", "err", err)
			return
		}

		grpcServer := grpc.NewServer()
		grpcserver.NewInventoryServer(invService, s.logger)

		s.logger.Info("gRPC server started", "port", s.cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			s.logger.Error("gRPC server error", "err", err)
		}
	}()

	serverAddress := fmt.Sprintf(":%s", s.cfg.Server.Port)
	s.logger.Info("HTTP server started", "port", s.cfg.Server.Port)
	return http.ListenAndServe(serverAddress, MWChain(s.mux))
}
