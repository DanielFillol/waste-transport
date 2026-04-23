package auth

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/danielfillol/waste/internal/config"
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/danielfillol/waste/pkg/jwtutil"
	"github.com/danielfillol/waste/pkg/password"
	"github.com/golang-jwt/jwt/v5"
)

type UseCase struct {
	tenantRepo *repository.TenantRepository
	userRepo   *repository.UserRepository
	cfg        *config.Config
}

func NewUseCase(tenantRepo *repository.TenantRepository, userRepo *repository.UserRepository, cfg *config.Config) *UseCase {
	return &UseCase{tenantRepo: tenantRepo, userRepo: userRepo, cfg: cfg}
}

func (uc *UseCase) RegisterTenant(name, username, pwd string) (*entity.Tenant, *entity.User, string, error) {
	slug := toSlug(name)
	if _, err := uc.tenantRepo.FindBySlug(slug); err == nil {
		return nil, nil, "", errors.New("tenant already exists")
	}

	tenant := &entity.Tenant{Name: name, Slug: slug}
	if err := uc.tenantRepo.Create(tenant); err != nil {
		return nil, nil, "", err
	}

	hashed, err := password.Hash(pwd)
	if err != nil {
		return nil, nil, "", err
	}

	admin := &entity.User{
		TenantID: tenant.ID,
		Name:     name + " Admin",
		Username: username,
		Password: hashed,
		Role:     entity.UserRoleAdmin,
	}
	if err := uc.userRepo.Create(admin); err != nil {
		return nil, nil, "", err
	}

	token, err := uc.generateToken(admin, tenant)
	if err != nil {
		return nil, nil, "", err
	}

	return tenant, admin, token, nil
}

func (uc *UseCase) RefreshToken(tokenStr string) (string, *entity.User, error) {
	claims := &jwtutil.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(uc.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", nil, errors.New("invalid or expired token")
	}

	user, err := uc.userRepo.FindByID(claims.UserID, claims.TenantID)
	if err != nil {
		return "", nil, errors.New("user not found")
	}

	tenant, err := uc.tenantRepo.FindByID(claims.TenantID)
	if err != nil {
		return "", nil, errors.New("tenant not found")
	}

	newToken, err := uc.generateToken(user, tenant)
	return newToken, user, err
}

func (uc *UseCase) Login(slug, username, pwd string) (string, *entity.User, error) {
	tenant, err := uc.tenantRepo.FindBySlug(slug)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	user, err := uc.userRepo.FindByUsername(username, tenant.ID)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if !password.Check(pwd, user.Password) {
		return "", nil, errors.New("invalid credentials")
	}

	token, err := uc.generateToken(user, tenant)
	return token, user, err
}

func (uc *UseCase) generateToken(user *entity.User, tenant *entity.Tenant) (string, error) {
	exp := time.Now().Add(time.Duration(uc.cfg.JWTExpirationHours) * time.Hour)
	claims := jwtutil.Claims{
		UserID:   user.ID,
		TenantID: tenant.ID,
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.cfg.JWTSecret))
}

func toSlug(name string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else if r == ' ' || r == '-' {
			sb.WriteRune('-')
		}
	}
	return sb.String()
}
