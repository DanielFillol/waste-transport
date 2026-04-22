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

type ReceiverHandler struct {
	repo    *repository.ReceiverRepository
	alertUC *opsUC.AlertUseCase
}

func NewReceiverHandler(repo *repository.ReceiverRepository, alertUC *opsUC.AlertUseCase) *ReceiverHandler {
	return &ReceiverHandler{repo: repo, alertUC: alertUC}
}

func (h *ReceiverHandler) List(c *gin.Context) {
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

func (h *ReceiverHandler) Get(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	rec, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "receiver not found"})
		return
	}
	c.JSON(http.StatusOK, rec)
}

func (h *ReceiverHandler) Create(c *gin.Context) {
	var req dto.CreateReceiverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	rec := &entity.Receiver{
		TenantID:      tenantID,
		ExternalID:    req.ExternalID,
		Name:          req.Name,
		CNPJ:          req.CNPJ,
		Address:       req.Address,
		Zipcode:       req.Zipcode,
		CityID:        req.CityID,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		LicenseNumber: req.LicenseNumber,
	}
	if t, err := parseOptionalDate(req.LicenseExpiry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		rec.LicenseExpiry = t
	}
	if err := h.repo.Create(rec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.alertUC.CheckReceiverAlerts(rec)
	c.JSON(http.StatusCreated, rec)
}

func (h *ReceiverHandler) Update(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	rec, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "receiver not found"})
		return
	}

	var req dto.UpdateReceiverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		rec.Name = *req.Name
	}
	if req.CNPJ != nil {
		rec.CNPJ = *req.CNPJ
	}
	if req.Address != nil {
		rec.Address = *req.Address
	}
	if req.Zipcode != nil {
		rec.Zipcode = *req.Zipcode
	}
	if req.CityID != nil {
		rec.CityID = req.CityID
	}
	if req.Latitude != nil {
		rec.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		rec.Longitude = req.Longitude
	}
	if req.LicenseNumber != nil {
		rec.LicenseNumber = *req.LicenseNumber
	}
	if t, err := parseOptionalDate(req.LicenseExpiry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else if t != nil {
		rec.LicenseExpiry = t
	}
	if req.Active != nil {
		rec.Active = *req.Active
	}

	if err := h.repo.Update(rec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.alertUC.CheckReceiverAlerts(rec)
	c.JSON(http.StatusOK, rec)
}

func (h *ReceiverHandler) Delete(c *gin.Context) {
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
