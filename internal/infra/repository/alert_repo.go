package repository

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlertRepository struct{ db *gorm.DB }

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) Create(a *entity.Alert) error {
	return r.db.Create(a).Error
}

func (r *AlertRepository) FindByID(id, tenantID uuid.UUID) (*entity.Alert, error) {
	var a entity.Alert
	if err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AlertRepository) List(tenantID uuid.UUID, onlyUnread bool, p pagination.Params) ([]entity.Alert, int64, error) {
	var items []entity.Alert
	var total int64
	q := r.db.Model(&entity.Alert{}).Where("tenant_id = ?", tenantID)
	if onlyUnread {
		q = q.Where("read = false")
	}
	q.Count(&total)
	err := pagination.Apply(q, p).Order("created_at DESC").Find(&items).Error
	return items, total, err
}

func (r *AlertRepository) MarkRead(id, tenantID uuid.UUID) error {
	return r.db.Model(&entity.Alert{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update("read", true).Error
}

func (r *AlertRepository) CountUnread(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Alert{}).
		Where("tenant_id = ? AND read = false", tenantID).
		Count(&count).Error
	return count, err
}

func (r *AlertRepository) MarkAllRead(tenantID uuid.UUID) error {
	return r.db.Model(&entity.Alert{}).
		Where("tenant_id = ? AND read = false", tenantID).
		Update("read", true).Error
}

func (r *AlertRepository) DeleteExisting(tenantID uuid.UUID, alertType entity.AlertType, ref string) error {
	return r.db.Where("tenant_id = ? AND type = ? AND message LIKE ?", tenantID, alertType, "%"+ref+"%").
		Delete(&entity.Alert{}).Error
}
