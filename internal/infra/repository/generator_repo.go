package repository

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GeneratorRepository struct{ db *gorm.DB }

func NewGeneratorRepository(db *gorm.DB) *GeneratorRepository {
	return &GeneratorRepository{db: db}
}

func (r *GeneratorRepository) Create(g *entity.Generator) error {
	return r.db.Create(g).Error
}

func (r *GeneratorRepository) FindByID(id, tenantID uuid.UUID) (*entity.Generator, error) {
	var g entity.Generator
	if err := r.db.Preload("City.UF").Where("id = ? AND tenant_id = ?", id, tenantID).First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GeneratorRepository) List(tenantID uuid.UUID, search string, active *bool, includeDeleted bool, p pagination.Params) ([]entity.Generator, int64, error) {
	var items []entity.Generator
	var total int64

	db := r.db
	if includeDeleted {
		db = db.Unscoped()
	}
	q := db.Model(&entity.Generator{}).Where("tenant_id = ?", tenantID)
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

func (r *GeneratorRepository) Update(g *entity.Generator) error {
	return r.db.Save(g).Error
}

func (r *GeneratorRepository) Delete(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.Generator{}).Error
}

func (r *GeneratorRepository) BulkCreate(items []entity.Generator) error {
	return r.db.CreateInBatches(items, 100).Error
}
