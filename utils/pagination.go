package utils

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Pagination struct {
	Page    int
	PerPage int
	Offset  int
}

func GetPagination(c *fiber.Ctx) Pagination {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return Pagination{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
	}
}

func BuildMeta(total int64, p Pagination) PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(p.PerPage)))
	return PaginationMeta{
		Total:      total,
		Page:       p.Page,
		PerPage:    p.PerPage,
		TotalPages: totalPages,
	}
}
