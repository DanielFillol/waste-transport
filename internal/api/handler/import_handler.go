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
	truckRepo     *repository.TruckRepository
	routeRepo     *repository.RouteRepository
}

func NewImportHandler(
	generatorRepo *repository.GeneratorRepository,
	receiverRepo *repository.ReceiverRepository,
	driverRepo *repository.DriverRepository,
	collectRepo *repository.CollectRepository,
	truckRepo *repository.TruckRepository,
	routeRepo *repository.RouteRepository,
) *ImportHandler {
	return &ImportHandler{
		generatorRepo: generatorRepo,
		receiverRepo:  receiverRepo,
		driverRepo:    driverRepo,
		collectRepo:   collectRepo,
		truckRepo:     truckRepo,
		routeRepo:     routeRepo,
	}
}

type importError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type importResult struct {
	Total   int           `json:"total"`
	Created int           `json:"created"`
	Updated int           `json:"updated"`
	Errors  []importError `json:"errors"`
}

type deleteResult struct {
	Total   int           `json:"total"`
	Deleted int64         `json:"deleted"`
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

func handleImportDelete(c *gin.Context, rows []csvparser.Row, deleteFn func([]uuid.UUID, uuid.UUID) (int64, error), tenantID uuid.UUID) {
	errs := []importError{}
	var ids []uuid.UUID
	for i, row := range rows {
		idStr := row["id"]
		if idStr == "" {
			errs = append(errs, importError{Row: i + 2, Message: "id is required"})
			continue
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("invalid id: %s", err.Error())})
			continue
		}
		ids = append(ids, id)
	}
	deleted := int64(0)
	if len(ids) > 0 {
		d, err := deleteFn(ids, tenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		deleted = d
	}
	c.JSON(http.StatusOK, deleteResult{Total: len(rows), Deleted: deleted, Errors: errs})
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
	created, updated := 0, 0

	for i, row := range rows {
		name := row["name"]
		if name == "" {
			errs = append(errs, importError{Row: i + 2, Message: "name is required"})
			continue
		}

		if idStr := row["id"]; idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				if existing, err := h.generatorRepo.FindByID(id, tenantID); err == nil {
					existing.Name = name
					existing.ExternalID = row["external_id"]
					existing.CNPJ = row["cnpj"]
					existing.Address = row["address"]
					existing.Zipcode = row["zipcode"]
					if v := row["latitude"]; v != "" {
						if f, err := strconv.ParseFloat(v, 64); err == nil {
							existing.Latitude = &f
						}
					}
					if v := row["longitude"]; v != "" {
						if f, err := strconv.ParseFloat(v, 64); err == nil {
							existing.Longitude = &f
						}
					}
					if err := h.generatorRepo.Update(existing); err != nil {
						errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("update failed: %s", err.Error())})
					} else {
						updated++
					}
					continue
				}
			}
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

	if len(items) > 0 {
		if err := h.generatorRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Updated: updated, Errors: errs})
}

func (h *ImportHandler) ImportDeleteGenerators(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	handleImportDelete(c, rows, h.generatorRepo.BulkDelete, tenantID)
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
	created, updated := 0, 0

	for i, row := range rows {
		name := row["name"]
		if name == "" {
			errs = append(errs, importError{Row: i + 2, Message: "name is required"})
			continue
		}

		if idStr := row["id"]; idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				if existing, err := h.receiverRepo.FindByID(id, tenantID); err == nil {
					existing.Name = name
					existing.ExternalID = row["external_id"]
					existing.CNPJ = row["cnpj"]
					existing.Address = row["address"]
					existing.Zipcode = row["zipcode"]
					existing.LicenseNumber = row["license_number"]
					if v := row["license_expiry"]; v != "" {
						if t, err := parseDate(v); err == nil {
							existing.LicenseExpiry = &t
						}
					}
					if err := h.receiverRepo.Update(existing); err != nil {
						errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("update failed: %s", err.Error())})
					} else {
						updated++
					}
					continue
				}
			}
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

	if len(items) > 0 {
		if err := h.receiverRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Updated: updated, Errors: errs})
}

func (h *ImportHandler) ImportDeleteReceivers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	handleImportDelete(c, rows, h.receiverRepo.BulkDelete, tenantID)
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
	created, updated := 0, 0

	for i, row := range rows {
		name := row["name"]
		if name == "" {
			errs = append(errs, importError{Row: i + 2, Message: "name is required"})
			continue
		}

		if idStr := row["id"]; idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				if existing, err := h.driverRepo.FindByID(id, tenantID); err == nil {
					existing.Name = name
					existing.ExternalID = row["external_id"]
					existing.Email = row["email"]
					existing.Phone = row["phone"]
					existing.CPF = row["cpf"]
					existing.CNHNumber = row["cnh_number"]
					existing.CNHCategory = row["cnh_category"]
					if v := row["cnh_expiry_date"]; v != "" {
						if t, err := parseDate(v); err == nil {
							existing.CNHExpiry = &t
						}
					}
					if err := h.driverRepo.Update(existing); err != nil {
						errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("update failed: %s", err.Error())})
					} else {
						updated++
					}
					continue
				}
			}
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

	if len(items) > 0 {
		if err := h.driverRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Updated: updated, Errors: errs})
}

func (h *ImportHandler) ImportDeleteDrivers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	handleImportDelete(c, rows, h.driverRepo.BulkDelete, tenantID)
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
	created, updated := 0, 0

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

		if idStr := row["id"]; idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				if existing, err := h.collectRepo.FindByID(id, tenantID); err == nil {
					existing.GeneratorID = genID
					existing.ReceiverID = recID
					existing.PlannedDate = plannedDate
					existing.ExternalID = row["external_id"]
					if v := row["collect_type"]; v != "" {
						existing.CollectType = entity.CollectType(v)
					}
					if v := row["route_id"]; v != "" {
						if rid, err := uuid.Parse(v); err == nil {
							existing.RouteID = &rid
						}
					}
					if err := h.collectRepo.Update(existing); err != nil {
						errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("update failed: %s", err.Error())})
					} else {
						updated++
					}
					continue
				}
			}
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

	if len(items) > 0 {
		if err := h.collectRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Updated: updated, Errors: errs})
}

func (h *ImportHandler) ImportDeleteCollects(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	handleImportDelete(c, rows, h.collectRepo.BulkDelete, tenantID)
}

func (h *ImportHandler) ImportTrucks(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []entity.Truck
	errs := []importError{}
	created, updated := 0, 0

	for i, row := range rows {
		plate := row["plate"]
		model := row["model"]
		if plate == "" || model == "" {
			errs = append(errs, importError{Row: i + 2, Message: "plate and model are required"})
			continue
		}

		if idStr := row["id"]; idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				if existing, err := h.truckRepo.FindByID(id, tenantID); err == nil {
					existing.Plate = plate
					existing.Model = model
					if v := row["year"]; v != "" {
						if n, err := strconv.Atoi(v); err == nil {
							existing.Year = n
						}
					}
					if v := row["capacity_kg"]; v != "" {
						if f, err := strconv.ParseFloat(v, 64); err == nil {
							existing.CapacityKG = f
						}
					}
					if v := row["capacity_m3"]; v != "" {
						if f, err := strconv.ParseFloat(v, 64); err == nil {
							existing.CapacityM3 = f
						}
					}
					if err := h.truckRepo.Update(existing); err != nil {
						errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("update failed: %s", err.Error())})
					} else {
						updated++
					}
					continue
				}
			}
		}

		t := entity.Truck{TenantID: tenantID, Plate: plate, Model: model, Active: true}
		if v := row["year"]; v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				t.Year = n
			}
		}
		if v := row["capacity_kg"]; v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				t.CapacityKG = f
			}
		}
		if v := row["capacity_m3"]; v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				t.CapacityM3 = f
			}
		}
		items = append(items, t)
	}

	if len(items) > 0 {
		if err := h.truckRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Updated: updated, Errors: errs})
}

func (h *ImportHandler) ImportDeleteTrucks(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	handleImportDelete(c, rows, h.truckRepo.BulkDelete, tenantID)
}

func (h *ImportHandler) ImportRoutes(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []entity.Route
	errs := []importError{}
	created, updated := 0, 0

	for i, row := range rows {
		name := row["name"]
		weekDayStr := row["week_day"]
		weekNumStr := row["week_number"]
		if name == "" || weekDayStr == "" || weekNumStr == "" {
			errs = append(errs, importError{Row: i + 2, Message: "name, week_day and week_number are required"})
			continue
		}
		weekDay, err1 := strconv.Atoi(weekDayStr)
		weekNum, err2 := strconv.Atoi(weekNumStr)
		if err1 != nil || err2 != nil {
			errs = append(errs, importError{Row: i + 2, Message: "week_day and week_number must be integers"})
			continue
		}

		if idStr := row["id"]; idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				if existing, err := h.routeRepo.FindByID(id, tenantID); err == nil {
					existing.Name = name
					existing.WeekDay = weekDay
					existing.WeekNumber = weekNum
					if err := h.routeRepo.Update(existing); err != nil {
						errs = append(errs, importError{Row: i + 2, Message: fmt.Sprintf("update failed: %s", err.Error())})
					} else {
						updated++
					}
					continue
				}
			}
		}

		items = append(items, entity.Route{TenantID: tenantID, Name: name, WeekDay: weekDay, WeekNumber: weekNum})
	}

	if len(items) > 0 {
		if err := h.routeRepo.BulkCreate(items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = len(items)
	}

	c.JSON(http.StatusOK, importResult{Total: len(rows), Created: created, Updated: updated, Errors: errs})
}

func (h *ImportHandler) ImportDeleteRoutes(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	rows, err := readCSVFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	handleImportDelete(c, rows, h.routeRepo.BulkDelete, tenantID)
}
