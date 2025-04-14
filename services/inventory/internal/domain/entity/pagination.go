package entity

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type SortOption string

const (
	SortByID        SortOption = "id"
	SortByPrice     SortOption = "price"
	SortByQuantity  SortOption = "quantity"
	SortByName      SortOption = "name"
	SortByCreatedAt SortOption = "created_at"
	SortByUpdatedAt SortOption = "updated_at"
)

type Pagination struct {
	Page     int64      `json:"page"`
	PageSize int64      `json:"pageSize"`
	SortBy   SortOption `json:"sortBy"`
}

type PaginationResponse[T any] struct {
	CurrentPage int64 `json:"current_page"`
	HasNextPage bool  `json:"has_next_page"`
	PageSize    int64 `json:"page_size"`
	TotalPages  int64 `json:"total_pages"`
	Data        []T   `json:"data"`
}

func NewPagination(page, pageSize int64, sortBy SortOption) *Pagination {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortBy,
	}
}

func NewPaginationFromRequest(r *http.Request, validSortByOptions []SortOption) (*Pagination, error) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")
	sortBy := r.URL.Query().Get("sortBy")

	page, pageSize := 1, 10

	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage < 0 {
			return nil, errors.New("invalid page number")
		}
		page = parsedPage
	}

	if pageSizeStr != "" {
		parsedPageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || parsedPageSize <= 0 {
			return nil, errors.New("invalid page size")
		}
		pageSize = parsedPageSize
	}

	pagination := NewPagination(int64(page), int64(pageSize), SortOption(strings.ToLower(sortBy)))

	if !pagination.Validate(validSortByOptions) {
		return nil, errors.New("invalid sortBy value")
	}

	return pagination, nil
}

func (p Pagination) Validate(validSortByOptions []SortOption) bool {
	if p.SortBy == "" {
		return true
	}

	for _, option := range validSortByOptions {
		if strings.ToLower(string(option)) == string(p.SortBy) {
			return true
		}
	}

	return false
}
