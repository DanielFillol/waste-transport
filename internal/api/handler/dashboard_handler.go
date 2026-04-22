package handler

import (
	"net/http"

	"github.com/danielfillol/waste/internal/api/middleware"
	opsUC "github.com/danielfillol/waste/internal/usecase/operations"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	uc *opsUC.DashboardUseCase
}

func NewDashboardHandler(uc *opsUC.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

func (h *DashboardHandler) Get(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	data, err := h.uc.Get(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
