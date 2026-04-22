package dto

import "github.com/danielfillol/waste/internal/domain/entity"

type CreatePricingRuleRequest struct {
	CollectType  *entity.CollectType     `json:"collect_type"`
	MaterialID   *uint                   `json:"material_id"`
	PackagingID  *uint                   `json:"packaging_id"`
	PricePerUnit float64                 `json:"price_per_unit" binding:"required,gt=0"`
	Unit         entity.MeasurementUnit  `json:"unit" binding:"required,oneof=KG LITER M3"`
}

type UpdatePricingRuleRequest struct {
	CollectType  *entity.CollectType     `json:"collect_type"`
	MaterialID   *uint                   `json:"material_id"`
	PackagingID  *uint                   `json:"packaging_id"`
	PricePerUnit *float64                `json:"price_per_unit"`
	Unit         *entity.MeasurementUnit `json:"unit"`
	Active       *bool                   `json:"active"`
}

type GenerateInvoiceRequest struct {
	GeneratorID string `json:"generator_id" binding:"required,uuid"`
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
	Notes       string `json:"notes"`
}

type IssueInvoiceRequest struct {
	DueDays *int `json:"due_days"`
}

type CreateTruckCostRequest struct {
	TruckID     string                  `json:"truck_id" binding:"required,uuid"`
	Type        entity.TruckCostType    `json:"type" binding:"required,oneof=fuel maintenance other"`
	PeriodStart string                  `json:"period_start" binding:"required"`
	PeriodEnd   string                  `json:"period_end" binding:"required"`
	TotalAmount float64                 `json:"total_amount" binding:"required,gt=0"`
	TotalKM     float64                 `json:"total_km"`
	Notes       string                  `json:"notes"`
}

type UpdateTruckCostRequest struct {
	TotalAmount *float64 `json:"total_amount"`
	TotalKM     *float64 `json:"total_km"`
	Notes       *string  `json:"notes"`
}

type CreatePersonnelCostRequest struct {
	DriverID    string                    `json:"driver_id" binding:"required,uuid"`
	Role        entity.PersonnelCostRole  `json:"role" binding:"required,oneof=driver collector"`
	PeriodMonth string                    `json:"period_month" binding:"required"`
	BaseSalary  float64                   `json:"base_salary" binding:"required,gt=0"`
	Benefits    float64                   `json:"benefits"`
	Notes       string                    `json:"notes"`
}

type UpdatePersonnelCostRequest struct {
	BaseSalary *float64 `json:"base_salary"`
	Benefits   *float64 `json:"benefits"`
	Notes      *string  `json:"notes"`
}
