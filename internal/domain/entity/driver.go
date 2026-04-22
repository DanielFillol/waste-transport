package entity

import (
	"time"

	"github.com/google/uuid"
)

type Driver struct {
	Base
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant      *Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	ExternalID  string     `json:"external_id"`
	Name        string     `gorm:"not null" json:"name"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	CPF         string     `json:"cpf"`
	CNHNumber   string     `json:"cnh_number"`
	CNHCategory string     `json:"cnh_category"`
	CNHExpiry   *time.Time `json:"cnh_expiry_date"`
	Active      bool       `gorm:"default:true" json:"active"`
}
