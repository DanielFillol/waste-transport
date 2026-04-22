package repository

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *entity.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepository) FindByID(id, tenantID uuid.UUID) (*entity.User, error) {
	var u entity.User
	if err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByUsername(username string, tenantID uuid.UUID) (*entity.User, error) {
	var u entity.User
	if err := r.db.Where("username = ? AND tenant_id = ?", username, tenantID).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) List(tenantID uuid.UUID, p pagination.Params) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64
	q := r.db.Model(&entity.User{}).Where("tenant_id = ?", tenantID)
	q.Count(&total)
	if err := pagination.Apply(q, p).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) Update(u *entity.User) error {
	return r.db.Save(u).Error
}

func (r *UserRepository) Delete(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.User{}).Error
}
