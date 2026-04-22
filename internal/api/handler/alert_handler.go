package handler

import (
	"net/http"

	"github.com/danielfillol/waste/internal/api/middleware"
	opsUC "github.com/danielfillol/waste/internal/usecase/operations"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/gin-gonic/gin"
)

type AlertHandler struct {
	uc *opsUC.AlertUseCase
}

func NewAlertHandler(uc *opsUC.AlertUseCase) *AlertHandler {
	return &AlertHandler{uc: uc}
}

func (h *AlertHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	onlyUnread := c.Query("unread") == "true"
	p := pagination.Parse(c)

	items, total, err := h.uc.ListAlerts(tenantID, onlyUnread, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}

func (h *AlertHandler) MarkAllRead(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if err := h.uc.MarkAllRead(tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"read": true})
}

func (h *AlertHandler) MarkRead(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)

	if err := h.uc.MarkRead(id, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"read": true})
}
