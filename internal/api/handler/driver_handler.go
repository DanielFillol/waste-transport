package handler

import (
	"net/http"

	"github.com/danielfillol/waste/internal/api/dto"
	"github.com/danielfillol/waste/internal/api/middleware"
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/repository"
	opsUC "github.com/danielfillol/waste/internal/usecase/operations"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/gin-gonic/gin"
)

type DriverHandler struct {
	repo    *repository.DriverRepository
	alertUC *opsUC.AlertUseCase
}

func NewDriverHandler(repo *repository.DriverRepository, alertUC *opsUC.AlertUseCase) *DriverHandler {
	return &DriverHandler{repo: repo, alertUC: alertUC}
}

func (h *DriverHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	search := c.Query("search")
	p := pagination.Parse(c)

	var active *bool
	if v := c.Query("active"); v != "" {
		b := v == "true"
		active = &b
	}
	includeDeleted := c.Query("include_deleted") == "true"

	items, total, err := h.repo.List(tenantID, search, active, includeDeleted, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}

func (h *DriverHandler) Get(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	d, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *DriverHandler) Create(c *gin.Context) {
	var req dto.CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	d := &entity.Driver{
		TenantID:    tenantID,
		ExternalID:  req.ExternalID,
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		CPF:         req.CPF,
		CNHNumber:   req.CNHNumber,
		CNHCategory: req.CNHCategory,
	}
	t, err := parseOptionalDate(req.CNHExpiry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d.CNHExpiry = t

	if err := h.repo.Create(d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.alertUC.CheckDriverAlerts(d)
	c.JSON(http.StatusCreated, d)
}

func (h *DriverHandler) Update(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	d, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
		return
	}

	var req dto.UpdateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Email != nil {
		d.Email = *req.Email
	}
	if req.Phone != nil {
		d.Phone = *req.Phone
	}
	if req.CPF != nil {
		d.CPF = *req.CPF
	}
	if req.CNHNumber != nil {
		d.CNHNumber = *req.CNHNumber
	}
	if req.CNHCategory != nil {
		d.CNHCategory = *req.CNHCategory
	}
	if t, err := parseOptionalDate(req.CNHExpiry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else if t != nil {
		d.CNHExpiry = t
	}
	if req.Active != nil {
		d.Active = *req.Active
	}

	if err := h.repo.Update(d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.alertUC.CheckDriverAlerts(d)
	c.JSON(http.StatusOK, d)
}

func (h *DriverHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	if err := h.repo.Delete(id, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
