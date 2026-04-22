package entity

import (
	"time"

	"github.com/google/uuid"
)

type CollectStatus int

const (
	CollectStatusPlanned   CollectStatus = 1
	CollectStatusCollected CollectStatus = 2
	CollectStatusCancelled CollectStatus = 3
)

type CollectType string

const (
	CollectTypeNormal  CollectType = "normal"
	CollectTypeSpecial CollectType = "special"
)

type MeasurementUnit string

const (
	MeasurementUnitKG     MeasurementUnit = "KG"
	MeasurementUnitLiter  MeasurementUnit = "LITER"
	MeasurementUnitM3     MeasurementUnit = "M3"
)

type Collect struct {
	Base
	TenantID          uuid.UUID       `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant            *Tenant         `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	GeneratorID       uuid.UUID       `gorm:"type:uuid;not null" json:"generator_id"`
	Generator         *Generator      `gorm:"foreignKey:GeneratorID" json:"generator,omitempty"`
	ReceiverID        uuid.UUID       `gorm:"type:uuid;not null" json:"receiver_id"`
	Receiver          *Receiver       `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	MaterialID        *uint           `json:"material_id"`
	Material          *Material       `gorm:"foreignKey:MaterialID" json:"material,omitempty"`
	PackagingID       *uint           `json:"packaging_id"`
	Packaging         *Packaging      `gorm:"foreignKey:PackagingID" json:"packaging,omitempty"`
	TreatmentID       *uint           `json:"treatment_id"`
	Treatment         *Treatment      `gorm:"foreignKey:TreatmentID" json:"treatment,omitempty"`
	RouteID           *uuid.UUID      `gorm:"type:uuid" json:"route_id"`
	Route             *Route          `gorm:"foreignKey:RouteID" json:"route,omitempty"`
	TruckID           *uuid.UUID      `gorm:"type:uuid" json:"truck_id"`
	Truck             *Truck          `gorm:"foreignKey:TruckID" json:"truck,omitempty"`
	ExternalID        string          `json:"external_id"`
	CollectType       CollectType     `gorm:"type:varchar(10);default:'normal'" json:"collect_type"`
	PlannedDate       time.Time       `gorm:"not null" json:"planned_date"`
	Status            CollectStatus   `gorm:"not null;default:1" json:"status"`
	CollectedQuantity *float64         `json:"collected_quantity"`
	CollectedUnit     *MeasurementUnit `json:"collected_unit"`
	CollectedWeight   *float64         `json:"collected_weight"`
	CollectedAt       *time.Time       `json:"collected_at"`
	Notes             string           `json:"notes"`
}
