package utils

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

func GetPagination(c echo.Context) Pagination {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

type PaginatedResponse struct {
	Data       any   `json:"data"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

func NewPaginatedResponse(data any, page, limit int, total int64) PaginatedResponse {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return PaginatedResponse{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
