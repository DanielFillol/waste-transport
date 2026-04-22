package entity

import (
	"time"

	"github.com/google/uuid"
)

// PricingRule defines the price for a combination of factors.
// Specificity score = number of non-null fields among CollectType, MaterialID, PackagingID.
// When multiple rules match a collect, the most specific (highest score) wins.
type PricingRule struct {
	Base
	TenantID    uuid.UUID        `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant      *Tenant          `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	CollectType *CollectType     `gorm:"type:varchar(10)" json:"collect_type"`
	MaterialID  *uint            `json:"material_id"`
	Material    *Material        `gorm:"foreignKey:MaterialID" json:"material,omitempty"`
	PackagingID *uint            `json:"packaging_id"`
	Packaging   *Packaging       `gorm:"foreignKey:PackagingID" json:"packaging,omitempty"`
	PricePerUnit float64         `gorm:"not null" json:"price_per_unit"`
	Unit         MeasurementUnit `gorm:"type:varchar(10);not null" json:"unit"`
	Active       bool            `gorm:"default:true" json:"active"`
}

type InvoiceStatus string

const (
	InvoiceStatusDraft  InvoiceStatus = "draft"
	InvoiceStatusIssued InvoiceStatus = "issued"
	InvoiceStatusPaid   InvoiceStatus = "paid"
)

type Invoice struct {
	Base
	TenantID      uuid.UUID     `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant        *Tenant       `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	GeneratorID   uuid.UUID     `gorm:"type:uuid;not null" json:"generator_id"`
	Generator     *Generator    `gorm:"foreignKey:GeneratorID" json:"generator,omitempty"`
	InvoiceNumber string        `gorm:"type:varchar(20)" json:"invoice_number"`
	PeriodStart   time.Time     `gorm:"not null" json:"period_start"`
	PeriodEnd     time.Time     `gorm:"not null" json:"period_end"`
	TotalAmount   float64       `gorm:"not null;default:0" json:"total_amount"`
	Status        InvoiceStatus `gorm:"type:varchar(10);not null;default:'draft'" json:"status"`
	IssuedAt      *time.Time    `json:"issued_at"`
	DueDate       *time.Time    `json:"due_date"`
	PaidAt        *time.Time    `json:"paid_at"`
	Notes         string        `json:"notes"`
	Items         []InvoiceItem `gorm:"foreignKey:InvoiceID" json:"items,omitempty"`
}

type InvoiceItem struct {
	Base
	InvoiceID    uuid.UUID        `gorm:"type:uuid;not null;index" json:"invoice_id"`
	CollectID    uuid.UUID        `gorm:"type:uuid;not null" json:"collect_id"`
	Collect      *Collect         `gorm:"foreignKey:CollectID" json:"collect,omitempty"`
	Description  string           `json:"description"`
	Quantity     float64          `gorm:"not null" json:"quantity"`
	Unit         MeasurementUnit  `gorm:"type:varchar(10);not null" json:"unit"`
	UnitPrice    float64          `gorm:"not null" json:"unit_price"`
	TotalPrice   float64          `gorm:"not null" json:"total_price"`
}

type TruckCostType string

const (
	TruckCostTypeFuel        TruckCostType = "fuel"
	TruckCostTypeMaintenance TruckCostType = "maintenance"
	TruckCostTypeOther       TruckCostType = "other"
)

type TruckCost struct {
	Base
	TenantID    uuid.UUID     `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant      *Tenant       `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	TruckID     uuid.UUID     `gorm:"type:uuid;not null" json:"truck_id"`
	Truck       *Truck        `gorm:"foreignKey:TruckID" json:"truck,omitempty"`
	Type        TruckCostType `gorm:"type:varchar(20);not null" json:"type"`
	PeriodStart time.Time     `gorm:"not null" json:"period_start"`
	PeriodEnd   time.Time     `gorm:"not null" json:"period_end"`
	TotalAmount float64       `gorm:"not null" json:"total_amount"`
	TotalKM     float64       `json:"total_km"`
	CostPerKM   float64       `json:"cost_per_km"`
	Notes       string        `json:"notes"`
}

type PersonnelCostRole string

const (
	PersonnelCostRoleDriver    PersonnelCostRole = "driver"
	PersonnelCostRoleCollector PersonnelCostRole = "collector"
)

type PersonnelCost struct {
	Base
	TenantID      uuid.UUID         `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant        *Tenant           `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	DriverID      uuid.UUID         `gorm:"type:uuid;not null" json:"driver_id"`
	Driver        *Driver           `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	Role          PersonnelCostRole `gorm:"type:varchar(20);not null" json:"role"`
	PeriodMonth   time.Time         `gorm:"not null" json:"period_month"`
	BaseSalary    float64           `gorm:"not null" json:"base_salary"`
	Benefits      float64           `gorm:"not null;default:0" json:"benefits"`
	TotalCost     float64           `gorm:"not null" json:"total_cost"`
	Notes         string            `json:"notes"`
}
