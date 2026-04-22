package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Params struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type Result[T any] struct {
	Data  []T   `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Pages int   `json:"pages"`
}

func Parse(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return Params{Page: page, Limit: limit}
}

func Apply(db *gorm.DB, p Params) *gorm.DB {
	offset := (p.Page - 1) * p.Limit
	return db.Offset(offset).Limit(p.Limit)
}

func NewResult[T any](data []T, total int64, p Params) Result[T] {
	pages := int(total) / p.Limit
	if int(total)%p.Limit != 0 {
		pages++
	}
	return Result[T]{
		Data:  data,
		Total: total,
		Page:  p.Page,
		Limit: p.Limit,
		Pages: pages,
	}
}
