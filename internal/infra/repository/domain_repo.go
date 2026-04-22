package repository

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"gorm.io/gorm"
)

type DomainRepository struct{ db *gorm.DB }

func NewDomainRepository(db *gorm.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

func (r *DomainRepository) ListMaterials() ([]entity.Material, error) {
	var items []entity.Material
	return items, r.db.Find(&items).Error
}

func (r *DomainRepository) ListPackagings() ([]entity.Packaging, error) {
	var items []entity.Packaging
	return items, r.db.Find(&items).Error
}

func (r *DomainRepository) ListTreatments() ([]entity.Treatment, error) {
	var items []entity.Treatment
	return items, r.db.Find(&items).Error
}

func (r *DomainRepository) ListUFs() ([]entity.UF, error) {
	var items []entity.UF
	return items, r.db.Find(&items).Error
}

func (r *DomainRepository) ListCities(ufID *uint) ([]entity.City, error) {
	var items []entity.City
	q := r.db.Preload("UF")
	if ufID != nil {
		q = q.Where("uf_id = ?", *ufID)
	}
	return items, q.Find(&items).Error
}
