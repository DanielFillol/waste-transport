package repository

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DriverRepository struct{ db *gorm.DB }

func NewDriverRepository(db *gorm.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

func (r *DriverRepository) Create(d *entity.Driver) error {
	return r.db.Create(d).Error
}

func (r *DriverRepository) FindByID(id, tenantID uuid.UUID) (*entity.Driver, error) {
	var d entity.Driver
	if err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DriverRepository) List(tenantID uuid.UUID, search string, active *bool, includeDeleted bool, p pagination.Params) ([]entity.Driver, int64, error) {
	var items []entity.Driver
	var total int64

	db := r.db
	if includeDeleted {
		db = db.Unscoped()
	}
	q := db.Model(&entity.Driver{}).Where("tenant_id = ?", tenantID)
	if search != "" {
		q = q.Where("name ILIKE ? OR cpf ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if active != nil {
		q = q.Where("active = ?", *active)
	}

	q.Count(&total)
	if err := pagination.Apply(q, p).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *DriverRepository) Update(d *entity.Driver) error {
	return r.db.Save(d).Error
}

func (r *DriverRepository) Delete(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.Driver{}).Error
}

func (r *DriverRepository) BulkCreate(items []entity.Driver) error {
	return r.db.Create(&items).Error
}

func (r *DriverRepository) BulkDelete(ids []uuid.UUID, tenantID uuid.UUID) (int64, error) {
	res := r.db.Where("id IN ? AND tenant_id = ?", ids, tenantID).Delete(&entity.Driver{})
	return res.RowsAffected, res.Error
}
