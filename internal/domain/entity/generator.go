package entity

import "github.com/google/uuid"

type Generator struct {
	Base
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant     *Tenant   `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	CityID     *uint     `json:"city_id"`
	City       *City     `gorm:"foreignKey:CityID" json:"city,omitempty"`
	ExternalID string    `json:"external_id"`
	Name       string    `gorm:"not null" json:"name"`
	CNPJ       string    `json:"cnpj"`
	Address    string    `json:"address"`
	Zipcode    string    `json:"zipcode"`
	Latitude   *float64  `json:"latitude"`
	Longitude  *float64  `json:"longitude"`
	Active     bool      `gorm:"default:true" json:"active"`
}
