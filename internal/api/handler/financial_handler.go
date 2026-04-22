package handler

import (
	"net/http"
	"time"

	"github.com/danielfillol/waste/internal/api/dto"
	"github.com/danielfillol/waste/internal/api/middleware"
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/usecase/financial"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FinancialHandler struct {
	uc *financial.UseCase
}

func NewFinancialHandler(uc *financial.UseCase) *FinancialHandler {
	return &FinancialHandler{uc: uc}
}

// Pricing Rules

func (h *FinancialHandler) ListPricingRules(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	onlyActive := c.Query("active") != "false"
	p := pagination.Parse(c)

	items, total, err := h.uc.ListPricingRules(tenantID, onlyActive, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}

func (h *FinancialHandler) GetPricingRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	rule, err := h.uc.GetPricingRule(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pricing rule not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *FinancialHandler) CreatePricingRule(c *gin.Context) {
	var req dto.CreatePricingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	rule, err := h.uc.CreatePricingRule(tenantID, financial.CreatePricingRuleInput{
		CollectType:  req.CollectType,
		MaterialID:   req.MaterialID,
		PackagingID:  req.PackagingID,
		PricePerUnit: req.PricePerUnit,
		Unit:         req.Unit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *FinancialHandler) UpdatePricingRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req dto.UpdatePricingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	rule, err := h.uc.UpdatePricingRule(id, tenantID, financial.UpdatePricingRuleInput{
		CollectType:  req.CollectType,
		MaterialID:   req.MaterialID,
		PackagingID:  req.PackagingID,
		PricePerUnit: req.PricePerUnit,
		Unit:         req.Unit,
		Active:       req.Active,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *FinancialHandler) DeletePricingRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	if err := h.uc.DeletePricingRule(id, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// Invoices

func (h *FinancialHandler) Summary(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	fromStr := c.Query("period_start")
	toStr := c.Query("period_end")
	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_start and period_end are required (YYYY-MM-DD)"})
		return
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_start format, use YYYY-MM-DD"})
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period_end format, use YYYY-MM-DD"})
		return
	}
	summary, err := h.uc.FinancialSummary(tenantID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *FinancialHandler) ListInvoices(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	p := pagination.Parse(c)

	var genID *uuid.UUID
	if v := c.Query("generator_id"); v != "" {
		id, _ := uuid.Parse(v)
		genID = &id
	}
	var status *entity.InvoiceStatus
	if v := c.Query("status"); v != "" {
		s := entity.InvoiceStatus(v)
		status = &s
	}

	items, total, err := h.uc.ListInvoices(tenantID, genID, status, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}

func (h *FinancialHandler) GetInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	inv, err := h.uc.GetInvoice(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}
	c.JSON(http.StatusOK, inv)
}

func (h *FinancialHandler) GenerateInvoice(c *gin.Context) {
	var req dto.GenerateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	generatorID, _ := uuid.Parse(req.GeneratorID)
	from, _ := time.Parse("2006-01-02", req.PeriodStart)
	to, _ := time.Parse("2006-01-02", req.PeriodEnd)

	inv, err := h.uc.GenerateInvoice(tenantID, generatorID, from, to, req.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, inv)
}

func (h *FinancialHandler) IssueInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req dto.IssueInvoiceRequest
	_ = c.ShouldBindJSON(&req)
	dueDays := 30
	if req.DueDays != nil {
		dueDays = *req.DueDays
	}
	tenantID := middleware.GetTenantID(c)
	inv, err := h.uc.IssueInvoice(id, tenantID, dueDays)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inv)
}

func (h *FinancialHandler) MarkInvoicePaid(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	inv, err := h.uc.MarkInvoicePaid(id, tenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inv)
}

// Truck Costs

func (h *FinancialHandler) ListTruckCosts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	p := pagination.Parse(c)

	var truckID *uuid.UUID
	if v := c.Query("truck_id"); v != "" {
		id, _ := uuid.Parse(v)
		truckID = &id
	}
	var from, to *time.Time
	if v := c.Query("from"); v != "" {
		t, _ := time.Parse("2006-01-02", v)
		from = &t
	}
	if v := c.Query("to"); v != "" {
		t, _ := time.Parse("2006-01-02", v)
		to = &t
	}

	items, total, err := h.uc.ListTruckCosts(tenantID, truckID, from, to, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}

func (h *FinancialHandler) GetTruckCost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	tc, err := h.uc.GetTruckCost(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "truck cost not found"})
		return
	}
	c.JSON(http.StatusOK, tc)
}

func (h *FinancialHandler) CreateTruckCost(c *gin.Context) {
	var req dto.CreateTruckCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	truckID, _ := uuid.Parse(req.TruckID)
	from, _ := time.Parse("2006-01-02", req.PeriodStart)
	to, _ := time.Parse("2006-01-02", req.PeriodEnd)

	tc, err := h.uc.CreateTruckCost(tenantID, financial.CreateTruckCostInput{
		TruckID:     truckID,
		Type:        req.Type,
		PeriodStart: from,
		PeriodEnd:   to,
		TotalAmount: req.TotalAmount,
		TotalKM:     req.TotalKM,
		Notes:       req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tc)
}

func (h *FinancialHandler) UpdateTruckCost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req dto.UpdateTruckCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	tc, err := h.uc.UpdateTruckCost(id, tenantID, financial.UpdateTruckCostInput{
		TotalAmount: req.TotalAmount,
		TotalKM:     req.TotalKM,
		Notes:       req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tc)
}

func (h *FinancialHandler) DeleteTruckCost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	if err := h.uc.DeleteTruckCost(id, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// Personnel Costs

func (h *FinancialHandler) ListPersonnelCosts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	p := pagination.Parse(c)

	var driverID *uuid.UUID
	if v := c.Query("driver_id"); v != "" {
		id, _ := uuid.Parse(v)
		driverID = &id
	}
	var month *time.Time
	if v := c.Query("month"); v != "" {
		t, _ := time.Parse("2006-01", v)
		month = &t
	}

	items, total, err := h.uc.ListPersonnelCosts(tenantID, driverID, month, p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pagination.NewResult(items, total, p))
}

func (h *FinancialHandler) GetPersonnelCost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	pc, err := h.uc.GetPersonnelCost(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "personnel cost not found"})
		return
	}
	c.JSON(http.StatusOK, pc)
}

func (h *FinancialHandler) CreatePersonnelCost(c *gin.Context) {
	var req dto.CreatePersonnelCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	driverID, _ := uuid.Parse(req.DriverID)
	month, _ := time.Parse("2006-01", req.PeriodMonth)

	pc, err := h.uc.CreatePersonnelCost(tenantID, financial.CreatePersonnelCostInput{
		DriverID:    driverID,
		Role:        req.Role,
		PeriodMonth: month,
		BaseSalary:  req.BaseSalary,
		Benefits:    req.Benefits,
		Notes:       req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pc)
}

func (h *FinancialHandler) UpdatePersonnelCost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req dto.UpdatePersonnelCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := middleware.GetTenantID(c)
	pc, err := h.uc.UpdatePersonnelCost(id, tenantID, financial.UpdatePersonnelCostInput{
		BaseSalary: req.BaseSalary,
		Benefits:   req.Benefits,
		Notes:      req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pc)
}

func (h *FinancialHandler) DeletePersonnelCost(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tenantID := middleware.GetTenantID(c)
	if err := h.uc.DeletePersonnelCost(id, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
