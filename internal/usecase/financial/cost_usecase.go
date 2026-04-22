package financial

import (
	"time"

	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
)

// TruckCost

func (uc *UseCase) ListTruckCosts(tenantID uuid.UUID, truckID *uuid.UUID, from, to *time.Time, p pagination.Params) ([]entity.TruckCost, int64, error) {
	return uc.repo.ListTruckCosts(tenantID, truckID, from, to, p)
}

func (uc *UseCase) GetTruckCost(id, tenantID uuid.UUID) (*entity.TruckCost, error) {
	return uc.repo.FindTruckCostByID(id, tenantID)
}

func (uc *UseCase) CreateTruckCost(tenantID uuid.UUID, input CreateTruckCostInput) (*entity.TruckCost, error) {
	costPerKM := 0.0
	if input.TotalKM > 0 {
		costPerKM = input.TotalAmount / input.TotalKM
	}

	tc := &entity.TruckCost{
		TenantID:    tenantID,
		TruckID:     input.TruckID,
		Type:        input.Type,
		PeriodStart: input.PeriodStart,
		PeriodEnd:   input.PeriodEnd,
		TotalAmount: input.TotalAmount,
		TotalKM:     input.TotalKM,
		CostPerKM:   costPerKM,
		Notes:       input.Notes,
	}
	return tc, uc.repo.CreateTruckCost(tc)
}

func (uc *UseCase) UpdateTruckCost(id, tenantID uuid.UUID, input UpdateTruckCostInput) (*entity.TruckCost, error) {
	tc, err := uc.repo.FindTruckCostByID(id, tenantID)
	if err != nil {
		return nil, err
	}

	if input.TotalAmount != nil {
		tc.TotalAmount = *input.TotalAmount
	}
	if input.TotalKM != nil {
		tc.TotalKM = *input.TotalKM
	}
	if input.Notes != nil {
		tc.Notes = *input.Notes
	}
	if tc.TotalKM > 0 {
		tc.CostPerKM = tc.TotalAmount / tc.TotalKM
	}

	return tc, uc.repo.UpdateTruckCost(tc)
}

func (uc *UseCase) DeleteTruckCost(id, tenantID uuid.UUID) error {
	return uc.repo.DeleteTruckCost(id, tenantID)
}

// PersonnelCost

func (uc *UseCase) ListPersonnelCosts(tenantID uuid.UUID, driverID *uuid.UUID, month *time.Time, p pagination.Params) ([]entity.PersonnelCost, int64, error) {
	return uc.repo.ListPersonnelCosts(tenantID, driverID, month, p)
}

func (uc *UseCase) GetPersonnelCost(id, tenantID uuid.UUID) (*entity.PersonnelCost, error) {
	return uc.repo.FindPersonnelCostByID(id, tenantID)
}

func (uc *UseCase) CreatePersonnelCost(tenantID uuid.UUID, input CreatePersonnelCostInput) (*entity.PersonnelCost, error) {
	pc := &entity.PersonnelCost{
		TenantID:    tenantID,
		DriverID:    input.DriverID,
		Role:        input.Role,
		PeriodMonth: input.PeriodMonth,
		BaseSalary:  input.BaseSalary,
		Benefits:    input.Benefits,
		TotalCost:   input.BaseSalary + input.Benefits,
		Notes:       input.Notes,
	}
	return pc, uc.repo.CreatePersonnelCost(pc)
}

func (uc *UseCase) UpdatePersonnelCost(id, tenantID uuid.UUID, input UpdatePersonnelCostInput) (*entity.PersonnelCost, error) {
	pc, err := uc.repo.FindPersonnelCostByID(id, tenantID)
	if err != nil {
		return nil, err
	}

	if input.BaseSalary != nil {
		pc.BaseSalary = *input.BaseSalary
	}
	if input.Benefits != nil {
		pc.Benefits = *input.Benefits
	}
	if input.Notes != nil {
		pc.Notes = *input.Notes
	}
	pc.TotalCost = pc.BaseSalary + pc.Benefits

	return pc, uc.repo.UpdatePersonnelCost(pc)
}

func (uc *UseCase) DeletePersonnelCost(id, tenantID uuid.UUID) error {
	return uc.repo.DeletePersonnelCost(id, tenantID)
}

type CreateTruckCostInput struct {
	TruckID     uuid.UUID
	Type        entity.TruckCostType
	PeriodStart time.Time
	PeriodEnd   time.Time
	TotalAmount float64
	TotalKM     float64
	Notes       string
}

type UpdateTruckCostInput struct {
	TotalAmount *float64
	TotalKM     *float64
	Notes       *string
}

type CreatePersonnelCostInput struct {
	DriverID    uuid.UUID
	Role        entity.PersonnelCostRole
	PeriodMonth time.Time
	BaseSalary  float64
	Benefits    float64
	Notes       string
}

type UpdatePersonnelCostInput struct {
	BaseSalary *float64
	Benefits   *float64
	Notes      *string
}
