package repository

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RouteRepository struct{ db *gorm.DB }

func NewRouteRepository(db *gorm.DB) *RouteRepository {
	return &RouteRepository{db: db}
}

func (r *RouteRepository) Create(route *entity.Route) error {
	return r.db.Create(route).Error
}

func (r *RouteRepository) FindByID(id, tenantID uuid.UUID) (*entity.Route, error) {
	var route entity.Route
	err := r.db.
		Preload("Material").Preload("Packaging").Preload("Treatment").Preload("Drivers").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&route).Error
	if err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *RouteRepository) List(tenantID uuid.UUID, search string, includeDeleted bool, p pagination.Params) ([]entity.Route, int64, error) {
	var items []entity.Route
	var total int64

	db := r.db
	if includeDeleted {
		db = db.Unscoped()
	}
	q := db.Model(&entity.Route{}).Where("tenant_id = ?", tenantID)
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	q.Count(&total)
	err := pagination.Apply(q, p).
		Preload("Material").Preload("Packaging").Preload("Treatment").Preload("Drivers").
		Find(&items).Error
	return items, total, err
}

func (r *RouteRepository) Update(route *entity.Route) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(route).Error
}

func (r *RouteRepository) Delete(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.Route{}).Error
}

func (r *RouteRepository) SetDrivers(routeID uuid.UUID, drivers []*entity.Driver) error {
	var route entity.Route
	if err := r.db.First(&route, "id = ?", routeID).Error; err != nil {
		return err
	}
	return r.db.Model(&route).Association("Drivers").Replace(drivers)
}
