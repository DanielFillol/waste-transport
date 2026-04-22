package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt  time.Time  `gorm:"index" json:"created_at"`
	TenantID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	ActorID    *uuid.UUID `gorm:"type:uuid;index" json:"actor_id"`
	EntityType string     `gorm:"type:varchar(50);not null;index" json:"entity_type"`
	EntityID   string     `gorm:"type:varchar(36);index" json:"entity_id"`
	Action     string     `gorm:"type:varchar(50);not null" json:"action"`
	Payload    string     `gorm:"type:text" json:"payload"`
}

func (a *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
