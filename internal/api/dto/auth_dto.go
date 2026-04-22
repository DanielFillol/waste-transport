package dto

import "github.com/danielfillol/waste/internal/domain/entity"

type RegisterTenantRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

type LoginRequest struct {
	Slug     string `json:"slug" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *entity.User `json:"user"`
}

type CreateUserRequest struct {
	Name     string           `json:"name" binding:"required"`
	Username string           `json:"username" binding:"required"`
	Password string           `json:"password" binding:"required,min=6"`
	Role     entity.UserRole  `json:"role" binding:"required,oneof=admin user"`
}

type UpdateUserRequest struct {
	Name     string           `json:"name"`
	Role     *entity.UserRole `json:"role"`
	Password *string          `json:"password"`
}
