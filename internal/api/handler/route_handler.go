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



type RouteHandler struct {
	repo        *repository.RouteRepository
	driverRepo  *repository.DriverRepository
	collectRepo *repository.CollectRepository
}

func NewRouteHandler(repo *repository.RouteRepository, driverRepo *repository.DriverRepository, collectRepo *repository.CollectRepository) *RouteHandler {
	return &RouteHandler{repo: repo, driverRepo: driverRepo, collectRepo: collectRepo}
}

func (h *RouteHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	search := c.Query("search")
	p := pagination.Parse(c)

	includeDeleted := c.Query("include_deleted") == "true"
	items, total, err := h.repo.List(tenantID, search, includeDeleted, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}

func (h *RouteHandler) Get(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	route, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		return
	}
	c.JSON(http.StatusOK, route)
}

func (h *RouteHandler) Create(c *gin.Context) {
	var req dto.CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)

	route := &entity.Route{
		TenantID:    tenantID,
		Name:        req.Name,
		MaterialID:  req.MaterialID,
		PackagingID: req.PackagingID,
		TreatmentID: req.TreatmentID,
		WeekDay:     req.WeekDay,
		WeekNumber:  req.WeekNumber,
	}
	if err := h.repo.Create(route); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(req.DriverIDs) > 0 {
		drivers := h.parseDrivers(req.DriverIDs, tenantID)
		_ = h.repo.SetDrivers(route.ID, drivers)
	}

	route, _ = h.repo.FindByID(route.ID, tenantID)
	c.JSON(http.StatusCreated, route)
}

func (h *RouteHandler) Update(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	route, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		return
	}

	var req dto.UpdateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		route.Name = *req.Name
	}
	if req.MaterialID != nil {
		route.MaterialID = req.MaterialID
	}
	if req.PackagingID != nil {
		route.PackagingID = req.PackagingID
	}
	if req.TreatmentID != nil {
		route.TreatmentID = req.TreatmentID
	}
	if req.WeekDay != nil {
		route.WeekDay = *req.WeekDay
	}
	if req.WeekNumber != nil {
		route.WeekNumber = *req.WeekNumber
	}

	if err := h.repo.Update(route); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.DriverIDs != nil {
		drivers := h.parseDrivers(req.DriverIDs, tenantID)
		_ = h.repo.SetDrivers(route.ID, drivers)
	}

	route, _ = h.repo.FindByID(route.ID, tenantID)
	c.JSON(http.StatusOK, route)
}

func (h *RouteHandler) Delete(c *gin.Context) {
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

func (h *RouteHandler) GenerateCollects(c *gin.Context) {
	routeID, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	route, err := h.repo.FindByID(routeID, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		return
	}

	var req dto.GenerateCollectsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	receiverID, err := parseUUID(req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receiver_id"})
		return
	}
	targetDate, err := parseDate(req.TargetDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collects := make([]entity.Collect, 0, len(req.GeneratorIDs))
	for _, gStr := range req.GeneratorIDs {
		gID, err := parseUUID(gStr)
		if err != nil {
			continue
		}
		collects = append(collects, entity.Collect{
			TenantID:    tenantID,
			GeneratorID: gID,
			ReceiverID:  receiverID,
			MaterialID:  route.MaterialID,
			PackagingID: route.PackagingID,
			TreatmentID: route.TreatmentID,
			RouteID:     &routeID,
			PlannedDate: targetDate,
			Status:      entity.CollectStatusPlanned,
		})
	}

	if len(collects) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid generator_ids provided"})
		return
	}
	if err := h.collectRepo.BulkCreate(collects); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"created": len(collects), "collects": collects})
}

func (h *RouteHandler) parseDrivers(ids []string, tenantID uuid.UUID) []*entity.Driver {
	var drivers []*entity.Driver
	for _, s := range ids {
		id, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		d, err := h.driverRepo.FindByID(id, tenantID)
		if err == nil {
			drivers = append(drivers, d)
		}
	}
	return drivers
}
