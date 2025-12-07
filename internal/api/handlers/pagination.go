package handlers

import (
	"strconv"
)

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page     int // 1-based page number
	PageSize int // Number of items per page
	Offset   int // Calculated offset (page - 1) * pageSize
}

// ParsePaginationParams parses pagination parameters from query string
// Defaults: page=1, page_size=50, max page_size=1000
func ParsePaginationParams(c interface {
	Query(key string) string
	DefaultQuery(key, defaultValue string) string
}) PaginationParams {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "50")

	page := 1
	if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
		page = parsedPage
	}

	pageSize := 50
	if parsedPageSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedPageSize > 0 {
		pageSize = parsedPageSize
	}

	// Enforce maximum page size
	if pageSize > 1000 {
		pageSize = 1000
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse[T any] struct {
	Data       []T  `json:"data"`
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalCount int  `json:"total_count"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse[T any](data []T, page, pageSize, totalCount int) PaginatedResponse[T] {
	totalPages := (totalCount + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	return PaginatedResponse[T]{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
