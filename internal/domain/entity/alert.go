package entity

import "github.com/google/uuid"

type AlertType string

const (
	AlertTypeLicenseExpiry AlertType = "license_expiry"
	AlertTypeCNHExpiry     AlertType = "cnh_expiry"
	AlertTypeGeneral       AlertType = "general"
)

type Alert struct {
	Base
	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant   *Tenant   `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Type     AlertType `gorm:"type:varchar(30);not null" json:"type"`
	Title    string    `gorm:"not null" json:"title"`
	Message  string    `json:"message"`
	Read     bool      `gorm:"default:false" json:"read"`
}
