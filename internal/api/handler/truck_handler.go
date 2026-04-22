package handler

import (
	"net/http"

	"github.com/danielfillol/waste/internal/api/dto"
	"github.com/danielfillol/waste/internal/api/middleware"
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TruckHandler struct {
	repo *repository.TruckRepository
}

func NewTruckHandler(repo *repository.TruckRepository) *TruckHandler {
	return &TruckHandler{repo: repo}
}

func (h *TruckHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	onlyActive := c.Query("active") == "true"
	search := c.Query("search")
	includeDeleted := c.Query("include_deleted") == "true"
	p := pagination.Parse(c)

	items, total, err := h.repo.List(tenantID, onlyActive, search, includeDeleted, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}

func (h *TruckHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	t, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "truck not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TruckHandler) Create(c *gin.Context) {
	var req dto.CreateTruckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	t := &entity.Truck{
		TenantID:   tenantID,
		Plate:      req.Plate,
		Model:      req.Model,
		Year:       req.Year,
		CapacityKG: req.CapacityKG,
		CapacityM3: req.CapacityM3,
		Active:     true,
	}
	if err := h.repo.Create(t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *TruckHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	t, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "truck not found"})
		return
	}

	var req dto.UpdateTruckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Plate != nil {
		t.Plate = *req.Plate
	}
	if req.Model != nil {
		t.Model = *req.Model
	}
	if req.Year != nil {
		t.Year = *req.Year
	}
	if req.CapacityKG != nil {
		t.CapacityKG = *req.CapacityKG
	}
	if req.CapacityM3 != nil {
		t.CapacityM3 = *req.CapacityM3
	}
	if req.Active != nil {
		t.Active = *req.Active
	}

	if err := h.repo.Update(t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TruckHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	if err := h.repo.Delete(id, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
