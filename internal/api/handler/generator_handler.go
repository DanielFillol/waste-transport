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

type GeneratorHandler struct {
	repo *repository.GeneratorRepository
}

func NewGeneratorHandler(repo *repository.GeneratorRepository) *GeneratorHandler {
	return &GeneratorHandler{repo: repo}
}

func (h *GeneratorHandler) List(c *gin.Context) {
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

func (h *GeneratorHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	g, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "generator not found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *GeneratorHandler) Create(c *gin.Context) {
	var req dto.CreateGeneratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	g := &entity.Generator{
		TenantID:   tenantID,
		ExternalID: req.ExternalID,
		Name:       req.Name,
		CNPJ:       req.CNPJ,
		Address:    req.Address,
		Zipcode:    req.Zipcode,
		CityID:     req.CityID,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}
	if err := h.repo.Create(g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, g)
}

func (h *GeneratorHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	g, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "generator not found"})
		return
	}

	var req dto.UpdateGeneratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		g.Name = *req.Name
	}
	if req.CNPJ != nil {
		g.CNPJ = *req.CNPJ
	}
	if req.Address != nil {
		g.Address = *req.Address
	}
	if req.Zipcode != nil {
		g.Zipcode = *req.Zipcode
	}
	if req.CityID != nil {
		g.CityID = req.CityID
	}
	if req.Latitude != nil {
		g.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		g.Longitude = req.Longitude
	}
	if req.Active != nil {
		g.Active = *req.Active
	}

	if err := h.repo.Update(g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *GeneratorHandler) Delete(c *gin.Context) {
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
