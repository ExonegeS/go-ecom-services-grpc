package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/grpc"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/rest"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/rest/middleware"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/outbound/cache"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/outbound/database"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/config"

	"log/slog"
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
	cacheRepo := cache.NewCacheRepository(invRepo)

	invService := application.NewInventoryService(cacheRepo, time.Now)

	invHandler := rest.NewInventoryHandler(invService, s.logger)
	healthHandler := rest.NewHealthHandler(s.logger, s.db)

	invHandler.RegisterEndpoints(s.mux, s.cfg)
	healthHandler.RegisterEndpoints(s.mux, s.cfg)

	loggerMW := middleware.NewLoggerMW(s.logger)
	MWChain := middleware.NewMiddlewareChain(middleware.RecoveryMW, loggerMW)

	go grpc.StartGRPCServer(s.cfg.Server.GRPCPort, invService, s.logger)

	serverAddress := fmt.Sprintf(":%s", s.cfg.Server.Port)
	s.logger.Info("HTTP server started", "port", s.cfg.Server.Port)
	return http.ListenAndServe(serverAddress, MWChain(s.mux))
}
