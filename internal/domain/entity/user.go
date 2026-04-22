package entity

import "github.com/google/uuid"

type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type User struct {
	Base
	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant   *Tenant   `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Name     string    `gorm:"not null" json:"name"`
	Username string    `gorm:"not null" json:"username"`
	Password string    `gorm:"not null" json:"-"`
	Role     UserRole  `gorm:"type:varchar(10);default:'user'" json:"role"`
}
