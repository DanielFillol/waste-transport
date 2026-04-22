package auth

import (
	"errors"

	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/danielfillol/waste/pkg/password"
	"github.com/google/uuid"
)

func (uc *UseCase) GetUser(tenantID, id uuid.UUID) (*entity.User, error) {
	return uc.userRepo.FindByID(id, tenantID)
}

func (uc *UseCase) ListUsers(tenantID uuid.UUID, p pagination.Params) ([]entity.User, int64, error) {
	return uc.userRepo.List(tenantID, p)
}

func (uc *UseCase) CreateUser(tenantID uuid.UUID, name, username, pwd string, role entity.UserRole) (*entity.User, error) {
	existing, _ := uc.userRepo.FindByUsername(username, tenantID)
	if existing != nil {
		return nil, errors.New("username already taken")
	}

	hashed, err := password.Hash(pwd)
	if err != nil {
		return nil, err
	}

	u := &entity.User{
		TenantID: tenantID,
		Name:     name,
		Username: username,
		Password: hashed,
		Role:     role,
	}
	return u, uc.userRepo.Create(u)
}

func (uc *UseCase) UpdateUser(tenantID, id uuid.UUID, name string, role *entity.UserRole, pwd *string) (*entity.User, error) {
	u, err := uc.userRepo.FindByID(id, tenantID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if name != "" {
		u.Name = name
	}
	if role != nil {
		u.Role = *role
	}
	if pwd != nil && *pwd != "" {
		hashed, err := password.Hash(*pwd)
		if err != nil {
			return nil, err
		}
		u.Password = hashed
	}

	return u, uc.userRepo.Update(u)
}

func (uc *UseCase) DeleteUser(tenantID, id uuid.UUID) error {
	return uc.userRepo.Delete(id, tenantID)
}
