package repository

import (
	"time"

	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CollectRepository struct{ db *gorm.DB }

func NewCollectRepository(db *gorm.DB) *CollectRepository {
	return &CollectRepository{db: db}
}

type CollectFilters struct {
	GeneratorID    *uuid.UUID
	ReceiverID     *uuid.UUID
	RouteID        *uuid.UUID
	TruckID        *uuid.UUID
	MaterialID     *uint
	PackagingID    *uint
	Status         *entity.CollectStatus
	CollectType    *entity.CollectType
	DateFrom       *time.Time
	DateTo         *time.Time
	IncludeDeleted bool
}

func (r *CollectRepository) Create(c *entity.Collect) error {
	return r.db.Create(c).Error
}

func (r *CollectRepository) FindByID(id, tenantID uuid.UUID) (*entity.Collect, error) {
	var c entity.Collect
	err := r.db.
		Preload("Generator").Preload("Receiver").
		Preload("Material").Preload("Packaging").Preload("Treatment").
		Preload("Route").Preload("Truck").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CollectRepository) List(tenantID uuid.UUID, f CollectFilters, p pagination.Params) ([]entity.Collect, int64, error) {
	var items []entity.Collect
	var total int64

	db := r.db
	if f.IncludeDeleted {
		db = db.Unscoped()
	}
	q := db.Model(&entity.Collect{}).Where("tenant_id = ?", tenantID)

	if f.GeneratorID != nil {
		q = q.Where("generator_id = ?", *f.GeneratorID)
	}
	if f.ReceiverID != nil {
		q = q.Where("receiver_id = ?", *f.ReceiverID)
	}
	if f.RouteID != nil {
		q = q.Where("route_id = ?", *f.RouteID)
	}
	if f.TruckID != nil {
		q = q.Where("truck_id = ?", *f.TruckID)
	}
	if f.MaterialID != nil {
		q = q.Where("material_id = ?", *f.MaterialID)
	}
	if f.PackagingID != nil {
		q = q.Where("packaging_id = ?", *f.PackagingID)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.CollectType != nil {
		q = q.Where("collect_type = ?", *f.CollectType)
	}
	if f.DateFrom != nil {
		q = q.Where("planned_date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil {
		q = q.Where("planned_date <= ?", *f.DateTo)
	}

	q.Count(&total)
	err := pagination.Apply(q, p).
		Preload("Generator").Preload("Receiver").
		Preload("Material").Preload("Packaging").Preload("Treatment").
		Preload("Route").Preload("Truck").
		Order("planned_date DESC").
		Find(&items).Error
	return items, total, err
}

func (r *CollectRepository) Update(c *entity.Collect) error {
	return r.db.Save(c).Error
}

func (r *CollectRepository) BulkUpdateStatus(ids []uuid.UUID, tenantID uuid.UUID, status entity.CollectStatus) error {
	return r.db.Model(&entity.Collect{}).
		Where("id IN ? AND tenant_id = ?", ids, tenantID).
		Update("status", status).Error
}

func (r *CollectRepository) BulkAssignRoute(ids []uuid.UUID, tenantID uuid.UUID, routeID *uuid.UUID) error {
	return r.db.Model(&entity.Collect{}).
		Where("id IN ? AND tenant_id = ?", ids, tenantID).
		Update("route_id", routeID).Error
}

func (r *CollectRepository) BulkCreate(items []entity.Collect) error {
	return r.db.CreateInBatches(items, 100).Error
}

func (r *CollectRepository) BulkDelete(ids []uuid.UUID, tenantID uuid.UUID) (int64, error) {
	res := r.db.Where("id IN ? AND tenant_id = ?", ids, tenantID).Delete(&entity.Collect{})
	return res.RowsAffected, res.Error
}

func (r *CollectRepository) CountByStatus(tenantID uuid.UUID) (map[entity.CollectStatus]int64, error) {
	type row struct {
		Status entity.CollectStatus
		Count  int64
	}
	var rows []row
	err := r.db.Model(&entity.Collect{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[entity.CollectStatus]int64)
	for _, r := range rows {
		result[r.Status] = r.Count
	}
	return result, nil
}

func (r *CollectRepository) FindByPeriodAndGenerator(tenantID, generatorID uuid.UUID, from, to time.Time) ([]entity.Collect, error) {
	var items []entity.Collect
	err := r.db.
		Where("tenant_id = ? AND generator_id = ? AND planned_date BETWEEN ? AND ? AND status = ?",
			tenantID, generatorID, from, to, entity.CollectStatusCollected).
		Preload("Material").Preload("Packaging").Preload("Treatment").
		Find(&items).Error
	return items, err
}
