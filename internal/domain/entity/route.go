package entity

import "github.com/google/uuid"

type Route struct {
	Base
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant      *Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Name        string     `gorm:"not null" json:"name"`
	MaterialID  *uint      `json:"material_id"`
	Material    *Material  `gorm:"foreignKey:MaterialID" json:"material,omitempty"`
	PackagingID *uint      `json:"packaging_id"`
	Packaging   *Packaging `gorm:"foreignKey:PackagingID" json:"packaging,omitempty"`
	TreatmentID *uint      `json:"treatment_id"`
	Treatment   *Treatment `gorm:"foreignKey:TreatmentID" json:"treatment,omitempty"`
	WeekDay     int        `gorm:"not null" json:"week_day"`
	WeekNumber  int        `gorm:"not null" json:"week_number"`
	Drivers     []*Driver  `gorm:"many2many:driver_routes;" json:"drivers,omitempty"`
}
