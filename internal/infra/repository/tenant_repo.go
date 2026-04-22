package repository

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantRepository struct{ db *gorm.DB }

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(t *entity.Tenant) error {
	return r.db.Create(t).Error
}

func (r *TenantRepository) FindBySlug(slug string) (*entity.Tenant, error) {
	var t entity.Tenant
	if err := r.db.Where("slug = ?", slug).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepository) FindByID(id uuid.UUID) (*entity.Tenant, error) {
	var t entity.Tenant
	if err := r.db.First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}
