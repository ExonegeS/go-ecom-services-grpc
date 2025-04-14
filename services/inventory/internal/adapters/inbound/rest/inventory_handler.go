package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/application"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/config"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/utils"
)

type InventoryHandler struct {
	service application.InventoryService
	logger  *slog.Logger
}

func NewInventoryHandler(service application.InventoryService, logger *slog.Logger) *InventoryHandler {
	return &InventoryHandler{
		service: service,
		logger:  logger,
	}
}

func (h *InventoryHandler) RegisterEndpoints(mux *http.ServeMux, cfg config.Config) {
	prefix := fmt.Sprintf("/api/%s", cfg.Version)
	addPrefix := func(method, path string) string {
		return fmt.Sprintf("%s %s%s", method, prefix, path)
	}

	mux.HandleFunc(addPrefix("POST", "/inventory"), h.CreateInventoryItem)
	mux.HandleFunc(addPrefix("POST", "/inventory/"), h.CreateInventoryItem)

	mux.HandleFunc(addPrefix("GET", "/inventory"), h.GetPaginatedInventoryItems)
	mux.HandleFunc(addPrefix("GET", "/inventory/"), h.GetPaginatedInventoryItems)

	mux.HandleFunc(addPrefix("GET", "/inventory/{id}"), h.GetInventoryItemById)
	mux.HandleFunc(addPrefix("GET", "/inventory/{id}/"), h.GetInventoryItemById)

	mux.HandleFunc(addPrefix("PUT", "/inventory/{id}"), h.UpdateInventoryItemById)
	mux.HandleFunc(addPrefix("PUT", "/inventory/{id}/"), h.UpdateInventoryItemById)

	mux.HandleFunc(addPrefix("DELETE", "/inventory/{id}"), h.DeleteInventoryItemById)
	mux.HandleFunc(addPrefix("DELETE", "/inventory/{id}/"), h.DeleteInventoryItemById)

	mux.HandleFunc(addPrefix("POST", "/category"), h.CreateCategory)
	mux.HandleFunc(addPrefix("POST", "/category/"), h.CreateCategory)

	mux.HandleFunc(addPrefix("GET", "/category/{id}"), h.GetCategoryByID)
	mux.HandleFunc(addPrefix("GET", "/category/{id}/"), h.GetCategoryByID)

	mux.HandleFunc(addPrefix("PUT", "/category/{id}"), h.UpdateCategoryByID)
	mux.HandleFunc(addPrefix("PUT", "/category/{id}/"), h.UpdateCategoryByID)

	mux.HandleFunc(addPrefix("DELETE", "/category/{id}"), h.DeleteCategoryByID)
	mux.HandleFunc(addPrefix("DELETE", "/category/{id}/"), h.DeleteCategoryByID)

	mux.HandleFunc(addPrefix("GET", "/inventory/check"), h.CheckInventoryItem)
}

func (h *InventoryHandler) GetInventoryItemById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := utils.ParseUUID(idStr)
	if err != nil {
		h.logger.Error("Cannot convert inventory item id to UUID", "error", err.Error())
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cannot convert inventory item id to UUID"))
		return
	}

	item, err := h.service.GetInventoryItemByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get inventory item", "id", id, "error", err.Error())
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("item with id %v not found", id))
		return
	}
	if item.Validate() != nil {
		h.logger.Error("Failed to validate retrieved inventory item", "item", item)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to validate retrieved inventory item"))
		return
	}

	resp := InventoryItemResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Category: Category{
			ID:          item.Category.ID,
			Name:        item.Category.Name,
			Description: item.Category.Description,
		},
		Price:     item.Price,
		Quantity:  item.Quantity,
		Unit:      item.Unit,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
	}

	h.logger.Info("Retrieved inventory item", "id", item.ID, "name", item.Name)
	utils.WriteJSON(w, http.StatusOK, resp)
}

type (
	CreateInventoryItemRequest struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		CategoryID  entity.UUID `json:"category_id"`
		Price       float64     `json:"price"`
		Quantity    float64     `json:"quantity"`
		Unit        string      `json:"unit"`
	}
	InventoryItemResponse struct {
		ID          entity.UUID `json:"id"`
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Category    Category    `json:"category"`
		Price       float64     `json:"price"`
		Quantity    float64     `json:"quantity"`
		Unit        string      `json:"unit"`
		CreatedAt   string      `json:"created_at"`
		UpdatedAt   string      `json:"updated_at"`
	}
	Category struct {
		ID          entity.UUID `json:"id"`
		Name        string      `json:"name"`
		Description string      `json:"description"`
	}
)

func (r *CreateInventoryItemRequest) Validate() error {
	if len(r.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if len(r.Description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}
	if _, err := utils.ParseUUID(string(r.CategoryID)); err != nil {
		return fmt.Errorf("invalid category ID: %w", err)
	}
	if r.Price <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}
	if r.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	if len(r.Unit) == 0 {
		return fmt.Errorf("unit cannot be empty")
	}
	return nil
}

func (h *InventoryHandler) CreateInventoryItem(w http.ResponseWriter, r *http.Request) {
	var req CreateInventoryItemRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("Failed to parse inventory item request", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusBadRequest, entity.ErrInvalidRequestPayload)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate request", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	e := entity.InventoryItem{
		Name:        req.Name,
		Description: req.Description,
		Category: entity.Category{
			ID: req.CategoryID,
		},
		Price:    req.Price,
		Quantity: req.Quantity,
		Unit:     req.Unit,
	}

	err = h.service.CreateInventoryItem(r.Context(), &e)
	if err != nil {
		if err == entity.ErrCategoryNotFound {
			h.logger.Error("Category not found", slog.String("error", err.Error()))
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("category not found"))
			return
		}

		h.logger.Error("Failed to create inventory item", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create inventory item"))
		return
	}

	resp := InventoryItemResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Category: Category{
			ID:          e.Category.ID,
			Name:        e.Category.Name,
			Description: e.Category.Description,
		},
		Price:     e.Price,
		Quantity:  e.Quantity,
		Unit:      e.Unit,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
	h.logger.Info("Created inventory item", "id", e.ID, "name", e.Name)
	utils.WriteJSON(w, http.StatusCreated, resp)
}

func (h *InventoryHandler) GetPaginatedInventoryItems(w http.ResponseWriter, r *http.Request) {
	pagination, err := entity.NewPaginationFromRequest(r, []entity.SortOption{
		entity.SortByID,
		entity.SortByName,
		entity.SortByPrice,
		entity.SortByQuantity,
		entity.SortByCreatedAt,
		entity.SortByUpdatedAt,
	})
	if err != nil {
		h.logger.Error("Invalid pagination request", "error", err.Error())
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	paginatedData, err := h.service.GetPaginatedInventoryItems(r.Context(), pagination)
	if err != nil {
		h.logger.Error("Failed to get menu items", "error", err.Error())
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get menu items"))
		return
	}

	response := entity.PaginationResponse[InventoryItemResponse]{
		CurrentPage: paginatedData.CurrentPage,
		HasNextPage: paginatedData.HasNextPage,
		PageSize:    paginatedData.PageSize,
		TotalPages:  paginatedData.TotalPages,
	}
	for _, item := range paginatedData.Data {
		response.Data = append(response.Data, InventoryItemResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Category: Category{
				ID:          item.Category.ID,
				Name:        item.Category.Name,
				Description: item.Category.Description,
			},
			Price:     item.Price,
			Quantity:  item.Quantity,
			Unit:      item.Unit,
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
			UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
		})
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

type (
	UpdateInventoryItemRequest struct {
		Name        *string      `json:"name"`
		Description *string      `json:"description"`
		CategoryID  *entity.UUID `json:"category_id"`
		Price       *float64     `json:"price"`
		Quantity    *float64     `json:"quantity"`
		Unit        *string      `json:"unit"`
	}
	UpdateInventoryItemResponse struct {
		ID          entity.UUID `json:"id"`
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Category    Category    `json:"category"`
		Price       float64     `json:"price"`
		Quantity    float64     `json:"quantity"`
		Unit        string      `json:"unit"`
		CreatedAt   string      `json:"created_at"`
		UpdatedAt   string      `json:"updated_at"`
	}
)

func (r *UpdateInventoryItemRequest) Validate() error {
	if r.Name != nil && len(*r.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if r.Description != nil && len(*r.Description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}
	if r.CategoryID != nil {
		if _, err := utils.ParseUUID(string(*r.CategoryID)); err != nil {
			return fmt.Errorf("invalid category ID: %w", err)
		}
	}
	if r.Price != nil && *r.Price <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}
	if r.Quantity != nil && *r.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	if r.Unit != nil && len(*r.Unit) == 0 {
		return fmt.Errorf("unit cannot be empty")
	}
	return nil
}

func (h *InventoryHandler) UpdateInventoryItemById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := utils.ParseUUID(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	var req UpdateInventoryItemRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		h.logger.Error("Failed to parse inventory item request", "error", err.Error())
		utils.WriteError(w, http.StatusBadRequest, entity.ErrInvalidRequestPayload)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate request", "error", err.Error())
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	params := application.UpdateInventoryItemParams{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Price:       req.Price,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
	}

	e, err := h.service.UpdateInventoryItem(r.Context(), id, params)
	if err != nil {
		h.logger.Error("Failed to update inventory item", "id", id, "error", err.Error())
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to update inventory item"))
		return
	}
	h.logger.Info("Succeeded to update inventory item", "id", id, "name", e.Name)

	resp := InventoryItemResponse{
		ID:          id,
		Name:        e.Name,
		Description: e.Description,
		Category: Category{
			ID:          e.Category.ID,
			Name:        e.Category.Name,
			Description: e.Category.Description,
		},
		Price:     e.Price,
		Quantity:  e.Quantity,
		Unit:      e.Unit,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}
func (h *InventoryHandler) DeleteInventoryItemById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := utils.ParseUUID(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	e, err := h.service.DeleteInventoryItem(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete inventory item", "id", id, "error", err.Error())
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to update inventory item"))
		return
	}
	h.logger.Info("Succeeded to delete inventory item", "id", id, "name", e.Name)
}

func (h *InventoryHandler) CheckInventoryItem(w http.ResponseWriter, r *http.Request) {}

type (
	CreateCategoryRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	CreateCategoryResponse struct {
		ID          entity.UUID `json:"id"`
		Name        string      `json:"name"`
		Description string      `json:"description"`
		CreatedAt   string      `json:"created_at"`
		UpdatedAt   string      `json:"updated_at"`
	}
)

func (r *CreateCategoryRequest) Validate() error {
	if len(r.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if len(r.Description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}
	return nil
}

func (h *InventoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("Failed to parse category request", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusBadRequest, entity.ErrInvalidRequestPayload)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate request", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	e := entity.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	err = h.service.CreateCategory(r.Context(), &e)
	if err != nil {
		h.logger.Error("Failed to create category", slog.String("error", err.Error()))
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create category"))
		return
	}

	resp := CreateCategoryResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
	utils.WriteJSON(w, http.StatusCreated, resp)
}

func (h *InventoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := utils.ParseUUID(idStr)
	if err != nil {
		h.logger.Error("Cannot convert category id to UUID", "error", err.Error())
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cannot convert category id to UUID"))
		return
	}

	category, err := h.service.GetCategoryByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get category", "id", id, "error", err.Error())
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("category with id %v not found", id))
		return
	}
	if category.Validate() != nil {
		h.logger.Error("Failed to validate retrieved category", "category", category)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to validate retrieved category"))
		return
	}

	resp := CreateCategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   category.UpdatedAt.Format(time.RFC3339),
	}

	h.logger.Info("Retrieved category", "id", category.ID, "name", category.Name)
	utils.WriteJSON(w, http.StatusOK, resp)
}

type (
	UpdateCategoryRequest struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	CategoryResponse struct {
		ID          entity.UUID `json:"id"`
		Name        string      `json:"name"`
		Description string      `json:"description"`
		CreatedAt   string      `json:"created_at"`
		UpdatedAt   string      `json:"updated_at"`
	}
)

func (r *UpdateCategoryRequest) Validate() error {
	if r.Name != nil && len(*r.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if r.Description != nil && len(*r.Description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}
	return nil
}

func (h *InventoryHandler) UpdateCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := utils.ParseUUID(idStr)
	if err != nil {
		h.logger.Error("Cannot convert category id to UUID", "error", err.Error())
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cannot convert category id to UUID"))
		return
	}
	var req UpdateCategoryRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		h.logger.Error("Failed to parse category request", "error", err.Error())
		utils.WriteError(w, http.StatusBadRequest, entity.ErrInvalidRequestPayload)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate request", "error", err.Error())
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	params := application.UpdateCategoryParams{
		Name:        req.Name,
		Description: req.Description,
	}

	e, err := h.service.UpdateCategory(r.Context(), id, params)
	if err != nil {
		h.logger.Error("Failed to update category", "id", id, "error", err.Error())
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to update category"))
		return
	}
	h.logger.Info("Succeeded to update category", "id", id, "name", e.Name)

	resp := CreateCategoryResponse{
		ID:          id,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *InventoryHandler) DeleteCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := utils.ParseUUID(idStr)
	if err != nil {
		h.logger.Error("Cannot convert category id to UUID", "error", err.Error())
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("cannot convert category id to UUID"))
		return
	}

	category, err := h.service.DeleteCategory(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete category", "id", id, "error", err.Error())
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete category"))
		return
	}
	h.logger.Info("Succeeded to delete category", "id", id, "name", category.Name)
}
