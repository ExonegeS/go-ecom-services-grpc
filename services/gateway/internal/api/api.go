package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/clients"
	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/config"
	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/handlers"
	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/handlers/middleware"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type APIServer struct {
	router      *mux.Router
	cfg         config.Config
	logger      *slog.Logger
	clientPools map[string]*clients.GrpcClientPool
}

func NewAPIServer(router *mux.Router, config config.Config, logger *slog.Logger) *APIServer {
	return &APIServer{
		router:      router,
		cfg:         config,
		logger:      logger,
		clientPools: make(map[string]*clients.GrpcClientPool),
	}
}

func (s *APIServer) Run() error {
	for _, svc := range s.cfg.Services {
		s.clientPools[svc.Name] = clients.NewGrpcClientPool(svc.GrpcAddr)
	}

	for i := range s.cfg.Services {
		svc := &s.cfg.Services[i]
		s.logger.Info("Registering service", "service", svc.Name)
		handlers.NewGatewayHandler(svc, s.clientPools[svc.Name]).RegisterRoutes(s.router)
	}

	go s.startHealthWorkers()

	MWChain := middleware.NewMiddlewareChain(
		middleware.RecoveryMW,
		middleware.NewCORS(s.cfg.CorsURLs),
		middleware.NewLoggerMW(s.logger),
		middleware.NewTimeoutContextMW(15),
	)

	serverAddress := fmt.Sprintf(":%s", s.cfg.Port)
	s.logger.Info("STARTING GATEWAY SERVICE", "Environment", s.cfg.Environment, "Version", s.cfg.Version, "Host", serverAddress)

	return http.ListenAndServe(
		fmt.Sprintf(":%s", s.cfg.Port),
		MWChain(s.router),
	)
}

func (s *APIServer) startHealthWorkers() {
	for i := range s.cfg.Services {
		svc := &s.cfg.Services[i]
		go func(svc *config.Service) {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				pool, exists := s.clientPools[svc.Name]
				if !exists {
					s.logger.Error("No client pool for service", "service", svc.Name)
					<-ticker.C
					continue
				}

				conn, err := pool.GetConn()
				if err != nil {
					svc.Status = "down"
					s.logger.Error("gRPC connection failed", "service", svc.Name, "address", svc.GrpcAddr, "error", err)
					<-ticker.C
					continue
				}

				healthClient := grpc_health_v1.NewHealthClient(conn)
				resp, err := healthClient.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{
					Service: svc.Name,
				})

				if err == nil || resp == nil {
					svc.Status = "up"
				} else {
					svc.Status = "down"
					s.logger.Error("Service health check failed",
						"service", svc.Name,
						"status", resp,
						"error", err)
				}

				<-ticker.C
			}
		}(svc)
	}
}
