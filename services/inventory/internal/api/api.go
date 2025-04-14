package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/rest"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/rest/middleware"
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

	invService := application.NewInventoryService(invRepo, time.Now)

	invHandler := rest.NewInventoryHandler(invService, s.logger)
	healthHandler := rest.NewHealthHandler(s.logger, s.db)

	invHandler.RegisterEndpoints(s.mux, s.cfg)
	healthHandler.RegisterEndpoints(s.mux, s.cfg)

	loggerMW := middleware.NewLoggerMW(s.logger)
	_ = loggerMW
	MWChain := middleware.NewMiddlewareChain()

	serverAddress := fmt.Sprintf(":%s", s.cfg.Server.Port)
	s.logger.Info("STARTING INVENTORY SERVICE", "Environment", s.cfg.Environment, "Version", s.cfg.Version, "Host", serverAddress)
	return http.ListenAndServe(serverAddress, MWChain(s.mux))
}
