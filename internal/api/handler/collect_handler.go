package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/danielfillol/waste/internal/api/dto"
	"github.com/danielfillol/waste/internal/api/middleware"
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CollectHandler struct {
	repo *repository.CollectRepository
}

func NewCollectHandler(repo *repository.CollectRepository) *CollectHandler {
	return &CollectHandler{repo: repo}
}

func (h *CollectHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	p := pagination.Parse(c)

	f := repository.CollectFilters{}

	if v := c.Query("generator_id"); v != "" {
		if id, err := parseUUID(v); err == nil {
			f.GeneratorID = &id
		}
	}
	if v := c.Query("receiver_id"); v != "" {
		if id, err := parseUUID(v); err == nil {
			f.ReceiverID = &id
		}
	}
	if v := c.Query("route_id"); v != "" {
		if id, err := parseUUID(v); err == nil {
			f.RouteID = &id
		}
	}
	if v := c.Query("truck_id"); v != "" {
		if id, err := parseUUID(v); err == nil {
			f.TruckID = &id
		}
	}
	if v := c.Query("material_id"); v != "" {
		var mid uint
		if _, err := fmt.Sscanf(v, "%d", &mid); err == nil {
			f.MaterialID = &mid
		}
	}
	if v := c.Query("packaging_id"); v != "" {
		var pid uint
		if _, err := fmt.Sscanf(v, "%d", &pid); err == nil {
			f.PackagingID = &pid
		}
	}
	if v := c.Query("status"); v != "" {
		var s entity.CollectStatus
		switch v {
		case "1":
			s = entity.CollectStatusPlanned
		case "2":
			s = entity.CollectStatusCollected
		case "3":
			s = entity.CollectStatusCancelled
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value, use 1, 2 or 3"})
			return
		}
		f.Status = &s
	}
	if v := c.Query("collect_type"); v != "" {
		ct := entity.CollectType(v)
		f.CollectType = &ct
	}
	if v := c.Query("date_from"); v != "" {
		t, err := parseDate(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		f.DateFrom = &t
	}
	if v := c.Query("date_to"); v != "" {
		t, err := parseDate(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		f.DateTo = &t
	}
	f.IncludeDeleted = c.Query("include_deleted") == "true"

	items, total, err := h.repo.List(tenantID, f, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}

func (h *CollectHandler) Get(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	col, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collect not found"})
		return
	}
	c.JSON(http.StatusOK, col)
}

func (h *CollectHandler) Create(c *gin.Context) {
	var req dto.CreateCollectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)

	generatorID, err := parseUUID(req.GeneratorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid generator_id"})
		return
	}
	receiverID, err := parseUUID(req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receiver_id"})
		return
	}
	plannedDate, err := parseDate(req.PlannedDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	col := &entity.Collect{
		TenantID:    tenantID,
		GeneratorID: generatorID,
		ReceiverID:  receiverID,
		MaterialID:  req.MaterialID,
		PackagingID: req.PackagingID,
		TreatmentID: req.TreatmentID,
		ExternalID:  req.ExternalID,
		CollectType: req.CollectType,
		PlannedDate: plannedDate,
		Status:      entity.CollectStatusPlanned,
		Notes:       req.Notes,
	}
	if col.CollectType == "" {
		col.CollectType = entity.CollectTypeNormal
	}
	if req.RouteID != nil {
		rid, err := parseUUID(*req.RouteID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route_id"})
			return
		}
		col.RouteID = &rid
	}
	if req.TruckID != nil {
		tid, err := parseUUID(*req.TruckID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid truck_id"})
			return
		}
		col.TruckID = &tid
	}

	if err := h.repo.Create(col); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, col)
}

func (h *CollectHandler) Update(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	col, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collect not found"})
		return
	}

	var req dto.UpdateCollectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != nil && (col.Status == entity.CollectStatusCollected || col.Status == entity.CollectStatusCancelled) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change status of a collected or cancelled collect"})
		return
	}

	if req.Status != nil {
		if *req.Status == entity.CollectStatusCollected && req.CollectedQuantity == nil && col.CollectedQuantity == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "collected_quantity and collected_unit are required when marking as collected"})
			return
		}
		if *req.Status == entity.CollectStatusCollected && req.CollectedUnit == nil && col.CollectedUnit == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "collected_quantity and collected_unit are required when marking as collected"})
			return
		}
		if *req.Status == entity.CollectStatusCollected && col.CollectedAt == nil {
			now := time.Now()
			col.CollectedAt = &now
		}
		col.Status = *req.Status
	}
	if req.CollectType != nil {
		col.CollectType = *req.CollectType
	}
	if req.CollectedQuantity != nil {
		col.CollectedQuantity = req.CollectedQuantity
	}
	if req.CollectedUnit != nil {
		col.CollectedUnit = req.CollectedUnit
	}
	if req.CollectedWeight != nil {
		col.CollectedWeight = req.CollectedWeight
	}
	if req.Notes != nil {
		col.Notes = *req.Notes
	}
	if req.RouteID != nil {
		rid, err := parseUUID(*req.RouteID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route_id"})
			return
		}
		col.RouteID = &rid
		col.Route = nil
	}
	if req.TruckID != nil {
		tid, err := parseUUID(*req.TruckID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid truck_id"})
			return
		}
		col.TruckID = &tid
		col.Truck = nil
	}

	if err := h.repo.Update(col); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, col)
}

func (h *CollectHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	col, err := h.repo.FindByID(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collect not found"})
		return
	}
	col.Status = entity.CollectStatusCancelled
	if err := h.repo.Update(col); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, col)
}

func (h *CollectHandler) BulkStatus(c *gin.Context) {
	var req dto.BulkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)

	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.BulkUpdateStatus(ids, tenantID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": len(ids)})
}

func (h *CollectHandler) BulkCancel(c *gin.Context) {
	var req dto.BulkCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)

	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.BulkUpdateStatus(ids, tenantID, entity.CollectStatusCancelled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cancelled": len(ids)})
}

func (h *CollectHandler) BulkAssignRoute(c *gin.Context) {
	var req dto.BulkAssignRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)

	ids, err := parseUUIDs(req.IDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	routeID, err := parseUUID(req.RouteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route_id"})
		return
	}

	if err := h.repo.BulkAssignRoute(ids, tenantID, &routeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": len(ids)})
}

func parseUUIDs(ss []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(ss))
	for _, s := range ss {
		id, err := parseUUID(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
