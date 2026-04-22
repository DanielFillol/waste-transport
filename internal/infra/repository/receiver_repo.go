package repository

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReceiverRepository struct{ db *gorm.DB }

func NewReceiverRepository(db *gorm.DB) *ReceiverRepository {
	return &ReceiverRepository{db: db}
}

func (r *ReceiverRepository) Create(rec *entity.Receiver) error {
	return r.db.Create(rec).Error
}

func (r *ReceiverRepository) FindByID(id, tenantID uuid.UUID) (*entity.Receiver, error) {
	var rec entity.Receiver
	if err := r.db.Preload("City.UF").Where("id = ? AND tenant_id = ?", id, tenantID).First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *ReceiverRepository) List(tenantID uuid.UUID, search string, active *bool, includeDeleted bool, p pagination.Params) ([]entity.Receiver, int64, error) {
	var items []entity.Receiver
	var total int64

	db := r.db
	if includeDeleted {
		db = db.Unscoped()
	}
	q := db.Model(&entity.Receiver{}).Where("tenant_id = ?", tenantID)
	if search != "" {
		q = q.Where("name ILIKE ? OR cnpj ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if active != nil {
		q = q.Where("active = ?", *active)
	}

	q.Count(&total)
	if err := pagination.Apply(q, p).Preload("City").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *ReceiverRepository) Update(rec *entity.Receiver) error {
	return r.db.Save(rec).Error
}

func (r *ReceiverRepository) Delete(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.Receiver{}).Error
}

func (r *ReceiverRepository) BulkCreate(items []entity.Receiver) error {
	return r.db.CreateInBatches(items, 100).Error
}
