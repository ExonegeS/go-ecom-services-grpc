package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/config"
	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/handlers"
	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/handlers/middleware"
)

type APIServer struct {
	mux    *http.ServeMux
	cfg    config.Config
	logger *slog.Logger
}

func NewAPIServer(mux *http.ServeMux, config config.Config, logger *slog.Logger) *APIServer {
	return &APIServer{mux, config, logger}
}

func (s *APIServer) Run() error {
	for i := range s.cfg.Services {
		svc := &s.cfg.Services[i]
		s.logger.Info("Connecting to service", "service", svc.Name, "URLBase", svc.URLBase)
		handlers.NewHandler(svc).RegisterEndpoints(s.mux)
	}

	go s.startHealthWorkers()

	timeoutMW := middleware.NewTimeoutContextMW(15)
	loggerMW := middleware.NewLoggerMW(s.logger)
	CORS_MW := middleware.NewCORS(s.cfg.CorsURLs)
	MWChain := middleware.NewMiddlewareChain(middleware.RecoveryMW, CORS_MW, loggerMW, timeoutMW)

	serverAddress := fmt.Sprintf(":%s", s.cfg.Port)
	s.logger.Info("STARTING GATEWAY SERVICE", "Environment", s.cfg.Environment, "Version", s.cfg.Version, "Host", serverAddress)

	return http.ListenAndServe(serverAddress, MWChain(s.mux))
}

func (s *APIServer) startHealthWorkers() {
	for i := range s.cfg.Services {
		svc := &s.cfg.Services[i]
		go func(svc *config.Service) {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				url := fmt.Sprintf("%s/api/%s/health", svc.URLBase, svc.ApiVersion)
				resp, err := http.Get(url)
				if err != nil {
					svc.Status = "down"
					s.logger.Error("Health check failed", "service", svc.Name, "error", err)
				} else {
					if resp.StatusCode == http.StatusOK {
						svc.Status = "up"
					} else {
						svc.Status = "down"
						body, err := io.ReadAll(resp.Body)

						if err != nil {
							s.logger.Error("Failed to read response body", "service", svc.Name, "error", err)
						} else {
							var response map[string]interface{}
							if err := json.Unmarshal(body, &response); err != nil {
								s.logger.Error("Failed to unmarshal response body", "service", svc.Name, "error", err)
							} else {
								s.logger.Info("Health check response", "service", svc.Name, "response", response)
							}
						}
					}
					resp.Body.Close()
				}
				<-ticker.C
			}
		}(svc)
	}
}
