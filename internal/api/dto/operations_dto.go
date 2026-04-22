package dto

import "github.com/danielfillol/waste/internal/domain/entity"

type CreateDriverRequest struct {
	ExternalID  string  `json:"external_id"`
	Name        string  `json:"name" binding:"required"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
	CPF         string  `json:"cpf"`
	CNHNumber   string  `json:"cnh_number"`
	CNHCategory string  `json:"cnh_category"`
	CNHExpiry   *string `json:"cnh_expiry_date"`
}

type UpdateDriverRequest struct {
	Name        *string `json:"name"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	CPF         *string `json:"cpf"`
	CNHNumber   *string `json:"cnh_number"`
	CNHCategory *string `json:"cnh_category"`
	CNHExpiry   *string `json:"cnh_expiry_date"`
	Active      *bool   `json:"active"`
}

type CreateTruckRequest struct {
	Plate      string  `json:"plate" binding:"required"`
	Model      string  `json:"model" binding:"required"`
	Year       int     `json:"year" binding:"required"`
	CapacityKG float64 `json:"capacity_kg"`
	CapacityM3 float64 `json:"capacity_m3"`
}

type UpdateTruckRequest struct {
	Plate      *string  `json:"plate"`
	Model      *string  `json:"model"`
	Year       *int     `json:"year"`
	CapacityKG *float64 `json:"capacity_kg"`
	CapacityM3 *float64 `json:"capacity_m3"`
	Active     *bool    `json:"active"`
}

type CreateRouteRequest struct {
	Name        string      `json:"name" binding:"required"`
	MaterialID  *uint       `json:"material_id"`
	PackagingID *uint       `json:"packaging_id"`
	TreatmentID *uint       `json:"treatment_id"`
	WeekDay     int         `json:"week_day" binding:"required,min=1,max=7"`
	WeekNumber  int         `json:"week_number" binding:"required,min=1,max=5"`
	DriverIDs   []string    `json:"driver_ids"`
}

type UpdateRouteRequest struct {
	Name        *string  `json:"name"`
	MaterialID  *uint    `json:"material_id"`
	PackagingID *uint    `json:"packaging_id"`
	TreatmentID *uint    `json:"treatment_id"`
	WeekDay     *int     `json:"week_day"`
	WeekNumber  *int     `json:"week_number"`
	DriverIDs   []string `json:"driver_ids"`
}

type CreateCollectRequest struct {
	GeneratorID  string                  `json:"generator_id" binding:"required,uuid"`
	ReceiverID   string                  `json:"receiver_id" binding:"required,uuid"`
	MaterialID   *uint                   `json:"material_id"`
	PackagingID  *uint                   `json:"packaging_id"`
	TreatmentID  *uint                   `json:"treatment_id"`
	RouteID      *string                 `json:"route_id"`
	TruckID      *string                 `json:"truck_id"`
	ExternalID   string                  `json:"external_id"`
	CollectType  entity.CollectType      `json:"collect_type"`
	PlannedDate  string                  `json:"planned_date" binding:"required"`
	Notes        string                  `json:"notes"`
}

type UpdateCollectRequest struct {
	Status            *entity.CollectStatus    `json:"status"`
	RouteID           *string                  `json:"route_id"`
	TruckID           *string                  `json:"truck_id"`
	CollectType       *entity.CollectType      `json:"collect_type"`
	CollectedQuantity *float64                 `json:"collected_quantity"`
	CollectedUnit     *entity.MeasurementUnit  `json:"collected_unit"`
	CollectedWeight   *float64                 `json:"collected_weight"`
	Notes             *string                  `json:"notes"`
}

type GenerateCollectsRequest struct {
	TargetDate   string   `json:"target_date" binding:"required"`
	GeneratorIDs []string `json:"generator_ids" binding:"required,min=1"`
	ReceiverID   string   `json:"receiver_id" binding:"required,uuid"`
}

type BulkStatusRequest struct {
	IDs    []string              `json:"ids" binding:"required,min=1"`
	Status entity.CollectStatus  `json:"status" binding:"required"`
}

type BulkCancelRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

type BulkAssignRouteRequest struct {
	IDs     []string `json:"ids" binding:"required,min=1"`
	RouteID string   `json:"route_id" binding:"required,uuid"`
}
