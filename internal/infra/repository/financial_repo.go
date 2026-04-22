package repository

import (
	"fmt"
	"time"

	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FinancialRepository struct{ db *gorm.DB }

func NewFinancialRepository(db *gorm.DB) *FinancialRepository {
	return &FinancialRepository{db: db}
}

// PricingRule

func (r *FinancialRepository) CreatePricingRule(p *entity.PricingRule) error {
	return r.db.Create(p).Error
}

func (r *FinancialRepository) FindPricingRuleByID(id, tenantID uuid.UUID) (*entity.PricingRule, error) {
	var p entity.PricingRule
	err := r.db.Preload("Material").Preload("Packaging").
		Where("id = ? AND tenant_id = ?", id, tenantID).First(&p).Error
	return &p, err
}

func (r *FinancialRepository) ListPricingRules(tenantID uuid.UUID, onlyActive bool, p pagination.Params) ([]entity.PricingRule, int64, error) {
	var items []entity.PricingRule
	var total int64
	q := r.db.Model(&entity.PricingRule{}).Where("tenant_id = ?", tenantID)
	if onlyActive {
		q = q.Where("active = true")
	}
	q.Count(&total)
	err := pagination.Apply(q, p).Preload("Material").Preload("Packaging").Find(&items).Error
	return items, total, err
}

func (r *FinancialRepository) UpdatePricingRule(p *entity.PricingRule) error {
	return r.db.Save(p).Error
}

func (r *FinancialRepository) DeletePricingRule(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.PricingRule{}).Error
}

// FindMatchingPricingRule returns the most specific rule for the given combination.
// Specificity = number of non-null discriminator fields (collect_type, material_id, packaging_id).
func (r *FinancialRepository) FindMatchingPricingRule(tenantID uuid.UUID, collectType entity.CollectType, materialID, packagingID *uint) (*entity.PricingRule, error) {
	var rules []entity.PricingRule
	err := r.db.Where("tenant_id = ? AND active = true", tenantID).Find(&rules).Error
	if err != nil {
		return nil, err
	}

	var best *entity.PricingRule
	bestScore := -1

	for i := range rules {
		rule := &rules[i]
		score := 0
		matches := true

		if rule.CollectType != nil {
			if *rule.CollectType != collectType {
				matches = false
			} else {
				score++
			}
		}
		if rule.MaterialID != nil {
			if materialID == nil || *rule.MaterialID != *materialID {
				matches = false
			} else {
				score++
			}
		}
		if rule.PackagingID != nil {
			if packagingID == nil || *rule.PackagingID != *packagingID {
				matches = false
			} else {
				score++
			}
		}

		if matches && score > bestScore {
			bestScore = score
			best = rule
		}
	}

	return best, nil
}

// Invoice

func (r *FinancialRepository) NextInvoiceNumber(tenantID uuid.UUID, year int) (string, error) {
	var count int64
	err := r.db.Model(&entity.Invoice{}).
		Where("tenant_id = ? AND EXTRACT(YEAR FROM created_at) = ?", tenantID, year).
		Count(&count).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d/%04d", year, count+1), nil
}

func (r *FinancialRepository) SumInvoiceRevenue(tenantID uuid.UUID, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.Model(&entity.Invoice{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("tenant_id = ? AND status IN ? AND period_start >= ? AND period_end <= ?",
			tenantID, []entity.InvoiceStatus{entity.InvoiceStatusIssued, entity.InvoiceStatusPaid}, from, to).
		Scan(&total).Error
	return total, err
}

func (r *FinancialRepository) SumTruckCosts(tenantID uuid.UUID, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.Model(&entity.TruckCost{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("tenant_id = ? AND period_start >= ? AND period_end <= ?", tenantID, from, to).
		Scan(&total).Error
	return total, err
}

func (r *FinancialRepository) SumPersonnelCosts(tenantID uuid.UUID, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.Model(&entity.PersonnelCost{}).
		Select("COALESCE(SUM(total_cost), 0)").
		Where("tenant_id = ? AND period_month >= ? AND period_month < ?", tenantID, from, to).
		Scan(&total).Error
	return total, err
}

func (r *FinancialRepository) CountInvoicesByStatus(tenantID uuid.UUID, status entity.InvoiceStatus) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Invoice{}).
		Where("tenant_id = ? AND status = ?", tenantID, status).
		Count(&count).Error
	return count, err
}

func (r *FinancialRepository) RunInTransaction(fn func(tx *FinancialRepository) error) error {
	return r.db.Transaction(func(db *gorm.DB) error {
		return fn(&FinancialRepository{db: db})
	})
}

func (r *FinancialRepository) ExistsOverlappingInvoice(tenantID, generatorID uuid.UUID, from, to time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Invoice{}).
		Where("tenant_id = ? AND generator_id = ? AND period_start <= ? AND period_end >= ?",
			tenantID, generatorID, to, from).
		Count(&count).Error
	return count > 0, err
}

func (r *FinancialRepository) CreateInvoice(inv *entity.Invoice) error {
	return r.db.Create(inv).Error
}

func (r *FinancialRepository) FindInvoiceByID(id, tenantID uuid.UUID) (*entity.Invoice, error) {
	var inv entity.Invoice
	err := r.db.
		Preload("Generator").
		Preload("Items.Collect.Material").
		Preload("Items.Collect.Packaging").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&inv).Error
	return &inv, err
}

func (r *FinancialRepository) ListInvoices(tenantID uuid.UUID, generatorID *uuid.UUID, status *entity.InvoiceStatus, p pagination.Params) ([]entity.Invoice, int64, error) {
	var items []entity.Invoice
	var total int64

	q := r.db.Model(&entity.Invoice{}).Where("tenant_id = ?", tenantID)
	if generatorID != nil {
		q = q.Where("generator_id = ?", *generatorID)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}

	q.Count(&total)
	err := pagination.Apply(q, p).Preload("Generator").Order("created_at DESC").Find(&items).Error
	return items, total, err
}

func (r *FinancialRepository) UpdateInvoice(inv *entity.Invoice) error {
	return r.db.Save(inv).Error
}

func (r *FinancialRepository) CreateInvoiceItems(items []entity.InvoiceItem) error {
	return r.db.CreateInBatches(items, 50).Error
}

// TruckCost

func (r *FinancialRepository) CreateTruckCost(tc *entity.TruckCost) error {
	return r.db.Create(tc).Error
}

func (r *FinancialRepository) FindTruckCostByID(id, tenantID uuid.UUID) (*entity.TruckCost, error) {
	var tc entity.TruckCost
	err := r.db.Preload("Truck").Where("id = ? AND tenant_id = ?", id, tenantID).First(&tc).Error
	return &tc, err
}

func (r *FinancialRepository) ListTruckCosts(tenantID uuid.UUID, truckID *uuid.UUID, from, to *time.Time, p pagination.Params) ([]entity.TruckCost, int64, error) {
	var items []entity.TruckCost
	var total int64

	q := r.db.Model(&entity.TruckCost{}).Where("tenant_id = ?", tenantID)
	if truckID != nil {
		q = q.Where("truck_id = ?", *truckID)
	}
	if from != nil {
		q = q.Where("period_start >= ?", *from)
	}
	if to != nil {
		q = q.Where("period_end <= ?", *to)
	}

	q.Count(&total)
	err := pagination.Apply(q, p).Preload("Truck").Order("period_start DESC").Find(&items).Error
	return items, total, err
}

func (r *FinancialRepository) UpdateTruckCost(tc *entity.TruckCost) error {
	return r.db.Save(tc).Error
}

func (r *FinancialRepository) DeleteTruckCost(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.TruckCost{}).Error
}

// PersonnelCost

func (r *FinancialRepository) CreatePersonnelCost(pc *entity.PersonnelCost) error {
	return r.db.Create(pc).Error
}

func (r *FinancialRepository) FindPersonnelCostByID(id, tenantID uuid.UUID) (*entity.PersonnelCost, error) {
	var pc entity.PersonnelCost
	err := r.db.Preload("Driver").Where("id = ? AND tenant_id = ?", id, tenantID).First(&pc).Error
	return &pc, err
}

func (r *FinancialRepository) ListPersonnelCosts(tenantID uuid.UUID, driverID *uuid.UUID, month *time.Time, p pagination.Params) ([]entity.PersonnelCost, int64, error) {
	var items []entity.PersonnelCost
	var total int64

	q := r.db.Model(&entity.PersonnelCost{}).Where("tenant_id = ?", tenantID)
	if driverID != nil {
		q = q.Where("driver_id = ?", *driverID)
	}
	if month != nil {
		start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		q = q.Where("period_month >= ? AND period_month < ?", start, end)
	}

	q.Count(&total)
	err := pagination.Apply(q, p).Preload("Driver").Order("period_month DESC").Find(&items).Error
	return items, total, err
}

func (r *FinancialRepository) UpdatePersonnelCost(pc *entity.PersonnelCost) error {
	return r.db.Save(pc).Error
}

func (r *FinancialRepository) DeletePersonnelCost(id, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.PersonnelCost{}).Error
}
