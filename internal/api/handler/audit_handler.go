package handler

import (
	"net/http"
	"time"

	"github.com/danielfillol/waste/internal/api/middleware"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditHandler struct {
	repo *repository.AuditRepository
}

func NewAuditHandler(repo *repository.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

func (h *AuditHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	p := pagination.Parse(c)

	f := repository.AuditFilters{}

	if v := c.Query("entity_type"); v != "" {
		f.EntityType = &v
	}
	if v := c.Query("entity_id"); v != "" {
		f.EntityID = &v
	}
	if v := c.Query("actor_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.ActorID = &id
		}
	}
	if v := c.Query("action"); v != "" {
		f.Action = &v
	}
	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.DateFrom = &t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.DateTo = &t
		}
	}

	items, total, err := h.repo.List(tenantID, f, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}
