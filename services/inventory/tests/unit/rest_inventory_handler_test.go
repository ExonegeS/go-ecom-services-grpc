package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/inbound/rest"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
	"github.com/ExonegeS/prettyslog"
)

type mockInventoryService struct {
	getByIDFn func(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error)
	createFn  func(ctx context.Context, item *entity.InventoryItem) error
	updateFn  func(ctx context.Context, id entity.UUID, params application.UpdateInventoryItemParams) (*entity.InventoryItem, error)
}

func (f *mockInventoryService) CreateInventoryItem(ctx context.Context, item *entity.InventoryItem) error {
	return f.createFn(ctx, item)
}
func (f *mockInventoryService) GetInventoryItemByID(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
	return f.getByIDFn(ctx, id)
}
func (f *mockInventoryService) UpdateInventoryItem(ctx context.Context, id entity.UUID, params application.UpdateInventoryItemParams) (*entity.InventoryItem, error) {
	return f.updateFn(ctx, id, params)
}
func (f *mockInventoryService) DeleteInventoryItem(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
	return nil, entity.ErrNotImplemented
}
func (m *mockInventoryService) GetPaginatedInventoryItems(ctx context.Context, pagination *entity.Pagination) (*entity.PaginationResponse[*entity.InventoryItem], error) {
	return nil, nil
}

func (f *mockInventoryService) CreateCategory(ctx context.Context, category *entity.Category) error {
	return nil
}
func (f *mockInventoryService) GetCategoryByID(ctx context.Context, id entity.UUID) (*entity.Category, error) {
	return &entity.Category{
		ID:          id,
		Name:        "Fake Category",
		Description: "This is a fake category",
		CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
	}, nil
}
func (f *mockInventoryService) UpdateCategory(ctx context.Context, id entity.UUID, params application.UpdateCategoryParams) (*entity.Category, error) {
	return nil, entity.ErrNotImplemented
}
func (f *mockInventoryService) DeleteCategory(ctx context.Context, id entity.UUID) (*entity.Category, error) {
	return nil, entity.ErrNotImplemented
}

// Test when a valid id is provided and the service returns a valid inventory item.
func TestGetInventoryItemById_Handler_Success(t *testing.T) {
	mockService := &mockInventoryService{
		getByIDFn: func(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
			return &entity.InventoryItem{
				ID:          id,
				Name:        "Fake Item",
				Description: "This is a fake item",
				Category:    entity.Category{},
				Price:       10.0,
				Quantity:    100,
				Unit:        "pcs",
				CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	logger := prettyslog.SetupPrettySlog(os.Stdout)
	handler := rest.NewInventoryHandler(mockService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/inventory/{id}", handler.GetInventoryItemById)

	id := "15e3191e-8124-42d6-b5f1-f4bef428f4f8"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/"+id, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status code 200, got %d", res.StatusCode)
	}

	var item entity.InventoryItem
	if err := json.NewDecoder(res.Body).Decode(&item); err != nil {
		t.Errorf("error decoding JSON response: %v", err)
	}
	if item.ID != entity.UUID(id) || item.Name != "Fake Item" {
		t.Errorf("unexpected item returned: %+v", item)
	}
}

func TestGetInventoryItemById_Handler_InvalidID_1(t *testing.T) {
	mockService := &mockInventoryService{
		getByIDFn: func(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
			return &entity.InventoryItem{
				ID:          id,
				Name:        "Fake Item",
				Description: "This is a fake item",
				Category:    entity.Category{},
				Price:       10.0,
				Quantity:    100,
				Unit:        "pcs",
				CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	logger := prettyslog.SetupPrettySlog(os.Stdout)
	handler := rest.NewInventoryHandler(mockService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/inventory/{id}", handler.GetInventoryItemById)

	id := "1"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/"+id, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code 400 for invalid id, got %d", res.StatusCode)
		return
	}
	var errResponse map[string]string
	if err := json.NewDecoder(res.Body).Decode(&errResponse); err != nil {
		t.Errorf("error decoding JSON response: %v", err)
		return
	}
	if errResponse["error"] != "cannot convert inventory item id to UUID" {
		t.Errorf("unexpected error message: %s", errResponse["error"])
		return
	}
}

func TestGetInventoryItemById_Handler_InvalidID_2(t *testing.T) {
	mockService := &mockInventoryService{
		getByIDFn: func(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
			return nil, entity.ErrItemNotFound
		},
	}
	logger := prettyslog.SetupPrettySlog(os.Stdout)
	handler := rest.NewInventoryHandler(mockService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/inventory/{id}", handler.GetInventoryItemById)

	id := "15e3191e-8124-42d6-b5f1-f4bef428f4f7"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/"+id, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	res := rec.Result()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code 404 for missing item, got %d", res.StatusCode)
		return
	}
}

func TestGetInventoryItemById_Handler_InvalidID_3(t *testing.T) {
	mockService := &mockInventoryService{
		getByIDFn: func(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
			return &entity.InventoryItem{
				ID:          id,
				Name:        "Fake Item",
				Description: "This is a fake item",
				Category:    entity.Category{},
				Price:       10.0,
				Quantity:    100,
				Unit:        "pcs",
				CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	logger := prettyslog.SetupPrettySlog(os.Stdout)
	handler := rest.NewInventoryHandler(mockService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/inventory/{id}", handler.GetInventoryItemById)

	id := "15e3191e-8124-32d6-b5f1-f4bef428f4f7"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/"+id, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code 400 for invalid id, got %d", res.StatusCode)
		return
	}
	var errResponse map[string]string
	if err := json.NewDecoder(res.Body).Decode(&errResponse); err != nil {
		t.Errorf("error decoding JSON response: %v", err)
		return
	}
	if errResponse["error"] != "cannot convert inventory item id to UUID" {
		t.Errorf("unexpected error message: %s", errResponse["error"])
		return
	}
}
