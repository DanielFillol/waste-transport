package repository

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TruckRepository struct{ db *gorm.DB }

func NewTruckRepository(db *gorm.DB) *TruckRepository {
	return &TruckRepository{db: db}
}

func (r *TruckRepository) Create(t *entity.Truck) error {
	return r.db.Create(t).Error
}

func (r *TruckRepository) FindByID(id, tenantID uuid.UUID) (*entity.Truck, error) {
	var t entity.Truck
	if err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TruckRepository) List(tenantID uuid.UUID, onlyActive bool, search string, includeDeleted bool, p pagination.Params) ([]entity.Truck, int64, error) {
	var items []entity.Truck
	var total int64

	db := r.db
	if includeDeleted {
		db = db.Unscoped()
	}
	q := db.Model(&entity.Truck{}).Where("tenant_id = ?", tenantID)
	if onlyActive {
		q = q.Where("active = true")
	}
	if search != "" {
		q = q.Where("plate ILIKE ? OR model ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	q.Count(&total)
	if err := pagination.Apply(q, p).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *TruckRepository) Update(t *entity.Truck) error {
	return r.db.Save(t).Error
}

func (r *TruckRepository) Delete(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.Truck{}).Error
}
