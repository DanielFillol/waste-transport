package entity

import "github.com/google/uuid"

type Truck struct {
	Base
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant     *Tenant   `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Plate      string    `gorm:"not null" json:"plate"`
	Model      string    `gorm:"not null" json:"model"`
	Year       int       `json:"year"`
	CapacityKG float64   `json:"capacity_kg"`
	CapacityM3 float64   `json:"capacity_m3"`
	Active     bool      `gorm:"default:true" json:"active"`
}
