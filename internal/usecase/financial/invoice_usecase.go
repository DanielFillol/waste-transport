package financial

import (
	"errors"
	"fmt"
	"time"

	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
)

type FinancialSummary struct {
	Revenue        float64 `json:"revenue"`
	TruckCosts     float64 `json:"truck_costs"`
	PersonnelCosts float64 `json:"personnel_costs"`
	GrossMargin    float64 `json:"gross_margin"`
}

func (uc *UseCase) FinancialSummary(tenantID uuid.UUID, from, to time.Time) (*FinancialSummary, error) {
	revenue, err := uc.repo.SumInvoiceRevenue(tenantID, from, to)
	if err != nil {
		return nil, err
	}
	truckCosts, err := uc.repo.SumTruckCosts(tenantID, from, to)
	if err != nil {
		return nil, err
	}
	personnelCosts, err := uc.repo.SumPersonnelCosts(tenantID, from, to)
	if err != nil {
		return nil, err
	}
	return &FinancialSummary{
		Revenue:        revenue,
		TruckCosts:     truckCosts,
		PersonnelCosts: personnelCosts,
		GrossMargin:    revenue - truckCosts - personnelCosts,
	}, nil
}

func (uc *UseCase) ListInvoices(tenantID uuid.UUID, generatorID *uuid.UUID, status *entity.InvoiceStatus, p pagination.Params) ([]entity.Invoice, int64, error) {
	return uc.repo.ListInvoices(tenantID, generatorID, status, p)
}

func (uc *UseCase) GetInvoice(id, tenantID uuid.UUID) (*entity.Invoice, error) {
	return uc.repo.FindInvoiceByID(id, tenantID)
}

// GenerateInvoice creates an invoice for a generator in a given period.
// It finds all COLLECTED collects in the period, applies the matching pricing rule to each,
// and creates invoice items. The entire operation is atomic.
func (uc *UseCase) GenerateInvoice(tenantID, generatorID uuid.UUID, from, to time.Time, notes string) (*entity.Invoice, error) {
	exists, err := uc.repo.ExistsOverlappingInvoice(tenantID, generatorID, from, to)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("invoice already exists for this generator and period")
	}

	collects, err := uc.collectRepo.FindByPeriodAndGenerator(tenantID, generatorID, from, to)
	if err != nil {
		return nil, err
	}
	if len(collects) == 0 {
		return nil, errors.New("no collected items found for this period")
	}

	var invoiceID uuid.UUID
	err = uc.repo.RunInTransaction(func(txRepo *repository.FinancialRepository) error {
		invoiceNum, err := txRepo.NextInvoiceNumber(tenantID, from.Year())
		if err != nil {
			return err
		}

		invoice := &entity.Invoice{
			TenantID:      tenantID,
			GeneratorID:   generatorID,
			InvoiceNumber: invoiceNum,
			PeriodStart:   from,
			PeriodEnd:     to,
			Status:        entity.InvoiceStatusDraft,
			Notes:         notes,
		}
		if err := txRepo.CreateInvoice(invoice); err != nil {
			return err
		}
		invoiceID = invoice.ID

		var items []entity.InvoiceItem
		var totalAmount float64

		for _, c := range collects {
			if c.CollectedQuantity == nil || c.CollectedUnit == nil {
				continue
			}
			rule, err := txRepo.FindMatchingPricingRule(tenantID, c.CollectType, c.MaterialID, c.PackagingID)
			if err != nil || rule == nil {
				continue
			}
			quantity := *c.CollectedQuantity
			unitPrice := rule.PricePerUnit
			items = append(items, entity.InvoiceItem{
				InvoiceID:   invoice.ID,
				CollectID:   c.ID,
				Description: buildItemDescription(c),
				Quantity:    quantity,
				Unit:        *c.CollectedUnit,
				UnitPrice:   unitPrice,
				TotalPrice:  quantity * unitPrice,
			})
			totalAmount += quantity * unitPrice
		}

		if len(items) > 0 {
			if err := txRepo.CreateInvoiceItems(items); err != nil {
				return err
			}
		}
		invoice.TotalAmount = totalAmount
		return txRepo.UpdateInvoice(invoice)
	})
	if err != nil {
		return nil, err
	}

	return uc.repo.FindInvoiceByID(invoiceID, tenantID)
}

func (uc *UseCase) IssueInvoice(id, tenantID uuid.UUID, dueDays int) (*entity.Invoice, error) {
	inv, err := uc.repo.FindInvoiceByID(id, tenantID)
	if err != nil {
		return nil, err
	}
	if inv.Status != entity.InvoiceStatusDraft {
		return nil, errors.New("only draft invoices can be issued")
	}
	now := time.Now()
	dueDate := now.AddDate(0, 0, dueDays)
	inv.Status = entity.InvoiceStatusIssued
	inv.IssuedAt = &now
	inv.DueDate = &dueDate
	return inv, uc.repo.UpdateInvoice(inv)
}

func (uc *UseCase) MarkInvoicePaid(id, tenantID uuid.UUID) (*entity.Invoice, error) {
	inv, err := uc.repo.FindInvoiceByID(id, tenantID)
	if err != nil {
		return nil, err
	}
	if inv.Status != entity.InvoiceStatusIssued {
		return nil, errors.New("only issued invoices can be marked as paid")
	}
	now := time.Now()
	inv.Status = entity.InvoiceStatusPaid
	inv.PaidAt = &now
	return inv, uc.repo.UpdateInvoice(inv)
}

func buildItemDescription(c entity.Collect) string {
	mat := ""
	if c.Material != nil {
		mat = c.Material.Name
	}
	pkg := ""
	if c.Packaging != nil {
		pkg = c.Packaging.Name
	}
	return fmt.Sprintf("Coleta %s - %s / %s - %s",
		string(c.CollectType),
		mat, pkg,
		c.PlannedDate.Format("02/01/2006"),
	)
}
