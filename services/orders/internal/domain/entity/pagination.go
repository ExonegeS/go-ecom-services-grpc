package entity

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type SortOption int

const (
	SortByUnknown SortOption = iota
	SortByID
	SortByPrice
	SortByQuantity
	SortByName
	SortByCreatedAt
	SortByUpdatedAt
)

var validSortOptions = map[string]SortOption{
	"id":         SortByID,
	"price":      SortByPrice,
	"quantity":   SortByQuantity,
	"name":       SortByName,
	"created_at": SortByCreatedAt,
	"updated_at": SortByUpdatedAt,
}

func ParseSortOption(s string) (SortOption, error) {
	if s == "" {
		return SortByUnknown, nil
	}
	if option, ok := validSortOptions[strings.ToLower(s)]; ok {
		return option, nil
	}
	return SortByUnknown, errors.New("invalid sortBy value")
}

func (s SortOption) ColumnName() string {
	switch s {
	case SortByID:
		return "id"
	case SortByPrice:
		return "price"
	case SortByQuantity:
		return "stock_quantity"
	case SortByName:
		return "name"
	case SortByCreatedAt:
		return "created_at"
	case SortByUpdatedAt:
		return "updated_at"
	default:
		return ""
	}
}

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
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortBy,
	}
}

func NewPaginationFromRequest(r *http.Request) (*Pagination, error) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")
	sortByStr := r.URL.Query().Get("sortBy")
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

	sortBy, err := ParseSortOption(sortByStr)
	if err != nil {
		return nil, err
	}

	return NewPagination(int64(page), int64(pageSize), sortBy), nil
}
