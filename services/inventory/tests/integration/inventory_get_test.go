package integration

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/rest"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/outbound/database"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/config"
	"github.com/ExonegeS/prettyslog"

	_ "github.com/lib/pq"
)

func TestHTTPGetInventoryItem(t *testing.T) {
	cfg := config.NewConfig("../../.env")
	connStr := cfg.DB.MakeConnectionString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	logger := prettyslog.SetupPrettySlog(os.Stdout)

	inventoryRepo := database.NewPostgresInventoryRepository(db)
	service := application.NewInventoryService(inventoryRepo, time.Now)
	handler := rest.NewInventoryHandler(service, logger)

	mux := http.NewServeMux()
	handler.RegisterEndpoints(mux, cfg)

	url := "/api/v1/inventory/706b1d58-9e09-479a-8e6b-b9b618927918"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status code 200, got %d", res.StatusCode)
	}
	fmt.Println("Integration test successful with status", res.StatusCode)
}
