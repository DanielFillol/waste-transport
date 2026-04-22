package operations

import (
	"time"

	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/repository"
	financialUC "github.com/danielfillol/waste/internal/usecase/financial"
	"github.com/google/uuid"
)

type DashboardUseCase struct {
	collectRepo   *repository.CollectRepository
	alertRepo     *repository.AlertRepository
	financialRepo *repository.FinancialRepository
	financialUC   *financialUC.UseCase
}

func NewDashboardUseCase(
	collectRepo *repository.CollectRepository,
	alertRepo *repository.AlertRepository,
	financialRepo *repository.FinancialRepository,
	financialUC *financialUC.UseCase,
) *DashboardUseCase {
	return &DashboardUseCase{
		collectRepo:   collectRepo,
		alertRepo:     alertRepo,
		financialRepo: financialRepo,
		financialUC:   financialUC,
	}
}

type CollectCounts struct {
	Planned   int64 `json:"planned"`
	Collected int64 `json:"collected"`
	Cancelled int64 `json:"cancelled"`
}

type InvoiceCounts struct {
	Draft  int64 `json:"draft"`
	Issued int64 `json:"issued"`
	Paid   int64 `json:"paid"`
}

type ThisMonth struct {
	Revenue        float64 `json:"revenue"`
	TruckCosts     float64 `json:"truck_costs"`
	PersonnelCosts float64 `json:"personnel_costs"`
}

type DashboardData struct {
	Collects     CollectCounts `json:"collects"`
	AlertsUnread int64         `json:"alerts_unread"`
	Invoices     InvoiceCounts `json:"invoices"`
	ThisMonth    ThisMonth     `json:"this_month"`
}

func (uc *DashboardUseCase) Get(tenantID uuid.UUID) (*DashboardData, error) {
	collectCounts, err := uc.collectRepo.CountByStatus(tenantID)
	if err != nil {
		return nil, err
	}

	alertsUnread, err := uc.alertRepo.CountUnread(tenantID)
	if err != nil {
		return nil, err
	}

	draftCount, _ := uc.financialRepo.CountInvoicesByStatus(tenantID, entity.InvoiceStatusDraft)
	issuedCount, _ := uc.financialRepo.CountInvoicesByStatus(tenantID, entity.InvoiceStatusIssued)
	paidCount, _ := uc.financialRepo.CountInvoicesByStatus(tenantID, entity.InvoiceStatusPaid)

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	revenue, _ := uc.financialRepo.SumInvoiceRevenue(tenantID, monthStart, monthEnd)
	truckCosts, _ := uc.financialRepo.SumTruckCosts(tenantID, monthStart, monthEnd)
	personnelCosts, _ := uc.financialRepo.SumPersonnelCosts(tenantID, monthStart, monthEnd)

	return &DashboardData{
		Collects: CollectCounts{
			Planned:   collectCounts[entity.CollectStatusPlanned],
			Collected: collectCounts[entity.CollectStatusCollected],
			Cancelled: collectCounts[entity.CollectStatusCancelled],
		},
		AlertsUnread: alertsUnread,
		Invoices: InvoiceCounts{
			Draft:  draftCount,
			Issued: issuedCount,
			Paid:   paidCount,
		},
		ThisMonth: ThisMonth{
			Revenue:        revenue,
			TruckCosts:     truckCosts,
			PersonnelCosts: personnelCosts,
		},
	}, nil
}
