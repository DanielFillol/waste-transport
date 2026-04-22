package repository

import (
	"time"

	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepository struct{ db *gorm.DB }

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *entity.AuditLog) error {
	return r.db.Create(log).Error
}

type AuditFilters struct {
	EntityType *string
	EntityID   *string
	ActorID    *uuid.UUID
	Action     *string
	DateFrom   *time.Time
	DateTo     *time.Time
}

func (r *AuditRepository) List(tenantID uuid.UUID, f AuditFilters, p pagination.Params) ([]entity.AuditLog, int64, error) {
	var items []entity.AuditLog
	var total int64

	q := r.db.Model(&entity.AuditLog{}).Where("tenant_id = ?", tenantID)

	if f.EntityType != nil {
		q = q.Where("entity_type = ?", *f.EntityType)
	}
	if f.EntityID != nil {
		q = q.Where("entity_id = ?", *f.EntityID)
	}
	if f.ActorID != nil {
		q = q.Where("actor_id = ?", *f.ActorID)
	}
	if f.Action != nil {
		q = q.Where("action = ?", *f.Action)
	}
	if f.DateFrom != nil {
		q = q.Where("created_at >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		q = q.Where("created_at <= ?", *f.DateTo)
	}

	q.Count(&total)
	err := pagination.Apply(q, p).Order("created_at DESC").Find(&items).Error
	return items, total, err
}
