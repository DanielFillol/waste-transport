package financial

import (
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
)

type UseCase struct {
	repo        *repository.FinancialRepository
	collectRepo *repository.CollectRepository
}

func NewUseCase(repo *repository.FinancialRepository, collectRepo *repository.CollectRepository) *UseCase {
	return &UseCase{repo: repo, collectRepo: collectRepo}
}

func (uc *UseCase) ListPricingRules(tenantID uuid.UUID, onlyActive bool, p pagination.Params) ([]entity.PricingRule, int64, error) {
	return uc.repo.ListPricingRules(tenantID, onlyActive, p)
}

func (uc *UseCase) GetPricingRule(id, tenantID uuid.UUID) (*entity.PricingRule, error) {
	return uc.repo.FindPricingRuleByID(id, tenantID)
}

func (uc *UseCase) CreatePricingRule(tenantID uuid.UUID, input CreatePricingRuleInput) (*entity.PricingRule, error) {
	rule := &entity.PricingRule{
		TenantID:     tenantID,
		CollectType:  input.CollectType,
		MaterialID:   input.MaterialID,
		PackagingID:  input.PackagingID,
		PricePerUnit: input.PricePerUnit,
		Unit:         input.Unit,
		Active:       true,
	}
	return rule, uc.repo.CreatePricingRule(rule)
}

func (uc *UseCase) UpdatePricingRule(id, tenantID uuid.UUID, input UpdatePricingRuleInput) (*entity.PricingRule, error) {
	rule, err := uc.repo.FindPricingRuleByID(id, tenantID)
	if err != nil {
		return nil, err
	}
	if input.CollectType != nil {
		rule.CollectType = input.CollectType
	}
	if input.MaterialID != nil {
		rule.MaterialID = input.MaterialID
	}
	if input.PackagingID != nil {
		rule.PackagingID = input.PackagingID
	}
	if input.PricePerUnit != nil {
		rule.PricePerUnit = *input.PricePerUnit
	}
	if input.Unit != nil {
		rule.Unit = *input.Unit
	}
	if input.Active != nil {
		rule.Active = *input.Active
	}
	return rule, uc.repo.UpdatePricingRule(rule)
}

func (uc *UseCase) DeletePricingRule(id, tenantID uuid.UUID) error {
	return uc.repo.DeletePricingRule(id, tenantID)
}

type CreatePricingRuleInput struct {
	CollectType  *entity.CollectType
	MaterialID   *uint
	PackagingID  *uint
	PricePerUnit float64
	Unit         entity.MeasurementUnit
}

type UpdatePricingRuleInput struct {
	CollectType  *entity.CollectType
	MaterialID   *uint
	PackagingID  *uint
	PricePerUnit *float64
	Unit         *entity.MeasurementUnit
	Active       *bool
}
