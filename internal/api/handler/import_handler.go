package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/danielfillol/waste/internal/api/middleware"
	"github.com/danielfillol/waste/internal/domain/entity"
	csvparser "github.com/danielfillol/waste/internal/infra/csv"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImportHandler struct {
	generatorRepo *repository.GeneratorRepository
	receiverRepo  *repository.ReceiverRepository
	driverRepo    *repository.DriverRepository
	collectRepo   *repository.CollectRepository
}

func NewImportHandler(
	generatorRepo *repository.GeneratorRepository,
	receiverRepo *repository.ReceiverRepository,
	driverRepo *repository.DriverRepository,
	collectRepo *repository.CollectRepository,
) *ImportHandler {
	return &ImportHandler{
		generatorRepo: generatorRepo,
		receiverRepo:  receiverRepo,
		driverRepo:    driverRepo,
		collectRepo:   collectRepo,
	}
}

type importError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type importResult struct {
	Total   int           `json:"total"`
	Created int           `json:"created"`
	Errors  []importError `json:"errors"`
}

func readCSVFromRequest(c *gin.Context) ([]csvparser.Row, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("file field required: %w", err)
	}
	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	return csvparser.Parse(f)
}

func (h *ImportHandler) ImportGenerators(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []entity.Generator
	errs := []importError{}

	for i, row := range rows {
		name := row["name"]
		if name == "" {
			errs = append(errs, importError{Row: i + 2, Message: "name is required"})
			continue
		}
		g := entity.Generator{
			TenantID:   tenantID,
			ExternalID: row["external_id"],
			Name:       name,
			CNPJ:       row["cnpj"],
			Address:    row["address"],
			Zipcode:    row["zipcode"],
		}
		if v := row["city_id"]; v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				uid := uint(id)
				g.CityID = &uid
			}
		}
		if v := row["latitude"]; v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				g.Latitude = &f
			}
		}
		if v := row["longitude"]; v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				g.Longitude = &f
			}
		}
		items = append(items, g)
	}

	created := 0
	if len(items) > 0 {
		if err := h.generatorRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Errors: errs})
}

func (h *ImportHandler) ImportReceivers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []entity.Receiver
	errs := []importError{}

	for i, row := range rows {
		name := row["name"]
		if name == "" {
			errs = append(errs, importError{Row: i + 2, Message: "name is required"})
			continue
		}
		rec := entity.Receiver{
			TenantID:      tenantID,
			ExternalID:    row["external_id"],
			Name:          name,
			CNPJ:          row["cnpj"],
			Address:       row["address"],
			Zipcode:       row["zipcode"],
			LicenseNumber: row["license_number"],
		}
		if v := row["city_id"]; v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				uid := uint(id)
				rec.CityID = &uid
			}
		}
		if v := row["latitude"]; v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				rec.Latitude = &f
			}
		}
		if v := row["longitude"]; v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				rec.Longitude = &f
			}
		}
		if v := row["license_expiry"]; v != "" {
			t, err := parseDate(v)
			if err != nil {
				errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("license_expiry: %s", err.Error())})
				continue
			}
			rec.LicenseExpiry = &t
		}
		items = append(items, rec)
	}

	created := 0
	if len(items) > 0 {
		if err := h.receiverRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Errors: errs})
}

func (h *ImportHandler) ImportDrivers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []entity.Driver
	errs := []importError{}

	for i, row := range rows {
		name := row["name"]
		if name == "" {
			errs = append(errs, importError{Row: i + 2, Message: "name is required"})
			continue
		}
		d := entity.Driver{
			TenantID:    tenantID,
			ExternalID:  row["external_id"],
			Name:        name,
			Email:       row["email"],
			Phone:       row["phone"],
			CPF:         row["cpf"],
			CNHNumber:   row["cnh_number"],
			CNHCategory: row["cnh_category"],
		}
		if v := row["cnh_expiry_date"]; v != "" {
			t, err := parseDate(v)
			if err != nil {
				errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("cnh_expiry_date: %s", err.Error())})
				continue
			}
			d.CNHExpiry = &t
		}
		items = append(items, d)
	}

	created := 0
	if len(items) > 0 {
		if err := h.driverRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Errors: errs})
}

func (h *ImportHandler) ImportCollects(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []entity.Collect
	errs := []importError{}

	for i, row := range rows {
		genIDStr := row["generator_id"]
		recIDStr := row["receiver_id"]
		plannedDateStr := row["planned_date"]

		if genIDStr == "" || recIDStr == "" || plannedDateStr == "" {
			errs = append(errs, importError{Row: i + 2, Message: "generator_id, receiver_id and planned_date are required"})
			continue
		}

		genID, err := uuid.Parse(genIDStr)
		if err != nil {
			errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("generator_id invalid: %s", err.Error())})
			continue
		}
		recID, err := uuid.Parse(recIDStr)
		if err != nil {
			errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("receiver_id invalid: %s", err.Error())})
			continue
		}
		plannedDate, err := parseDate(plannedDateStr)
		if err != nil {
			errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("planned_date: %s", err.Error())})
			continue
		}

		col := entity.Collect{
			TenantID:    tenantID,
			GeneratorID: genID,
			ReceiverID:  recID,
			ExternalID:  row["external_id"],
			PlannedDate: plannedDate,
			Status:      entity.CollectStatusPlanned,
			CollectType: entity.CollectTypeNormal,
		}
		if v := row["collect_type"]; v != "" {
			col.CollectType = entity.CollectType(v)
		}
		if v := row["material_id"]; v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				uid := uint(id)
				col.MaterialID = &uid
			}
		}
		if v := row["packaging_id"]; v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				uid := uint(id)
				col.PackagingID = &uid
			}
		}
		if v := row["treatment_id"]; v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				uid := uint(id)
				col.TreatmentID = &uid
			}
		}
		if v := row["route_id"]; v != "" {
			if id, err := uuid.Parse(v); err == nil {
				col.RouteID = &id
			}
		}
		items = append(items, col)
	}

	created := 0
	if len(items) > 0 {
		if err := h.collectRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Errors: errs})
}
