package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/config"
)

type HealthHandler struct {
	logger *slog.Logger
	db     *sql.DB
}

func NewHealthHandler(logger *slog.Logger, db *sql.DB) *HealthHandler {
	return &HealthHandler{logger: logger, db: db}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func (h *HealthHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	if err := h.db.Ping(); err != nil {
		dbStatus = fmt.Sprintf("error: %v", err)
	}
	status := "ok"
	if dbStatus != "ok" {
		status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	response := HealthResponse{
		Status:   status,
		Database: dbStatus,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) RegisterEndpoints(mux *http.ServeMux, cfg config.Config) {
	prefix := fmt.Sprintf("/api/%s", cfg.Version)
	mux.HandleFunc(prefix+"/health", h.healthCheck)
}
