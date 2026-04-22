package operations

import (
	"fmt"
	"time"

	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/danielfillol/waste/pkg/pagination"
	"github.com/google/uuid"
)

const expiryWarningDays = 30

type AlertUseCase struct {
	repo *repository.AlertRepository
}

func NewAlertUseCase(repo *repository.AlertRepository) *AlertUseCase {
	return &AlertUseCase{repo: repo}
}

func (uc *AlertUseCase) ListAlerts(tenantID uuid.UUID, onlyUnread bool, p pagination.Params) ([]entity.Alert, int64, error) {
	return uc.repo.List(tenantID, onlyUnread, p)
}

func (uc *AlertUseCase) MarkRead(id, tenantID uuid.UUID) error {
	return uc.repo.MarkRead(id, tenantID)
}

func (uc *AlertUseCase) MarkAllRead(tenantID uuid.UUID) error {
	return uc.repo.MarkAllRead(tenantID)
}

// CheckDriverAlerts generates or removes a CNH expiry alert for the driver.
func (uc *AlertUseCase) CheckDriverAlerts(driver *entity.Driver) error {
	if driver.CNHExpiry == nil {
		return nil
	}

	daysUntil := int(time.Until(*driver.CNHExpiry).Hours() / 24)

	// Remove existing alert for this driver
	_ = uc.repo.DeleteExisting(driver.TenantID, entity.AlertTypeCNHExpiry, driver.ID.String())

	if daysUntil <= expiryWarningDays && daysUntil >= 0 {
		alert := &entity.Alert{
			TenantID: driver.TenantID,
			Type:     entity.AlertTypeCNHExpiry,
			Title:    fmt.Sprintf("CNH vencendo em %d dias — %s", daysUntil, driver.Name),
			Message:  fmt.Sprintf("Motorista %s (ID: %s) tem CNH com vencimento em %s.", driver.Name, driver.ID.String(), driver.CNHExpiry.Format("02/01/2006")),
		}
		return uc.repo.Create(alert)
	}
	return nil
}

// CheckReceiverAlerts generates or removes a license expiry alert for the receiver.
func (uc *AlertUseCase) CheckReceiverAlerts(receiver *entity.Receiver) error {
	if receiver.LicenseExpiry == nil {
		return nil
	}

	daysUntil := int(time.Until(*receiver.LicenseExpiry).Hours() / 24)

	_ = uc.repo.DeleteExisting(receiver.TenantID, entity.AlertTypeLicenseExpiry, receiver.ID.String())

	if daysUntil <= expiryWarningDays && daysUntil >= 0 {
		alert := &entity.Alert{
			TenantID: receiver.TenantID,
			Type:     entity.AlertTypeLicenseExpiry,
			Title:    fmt.Sprintf("Licença vencendo em %d dias — %s", daysUntil, receiver.Name),
			Message:  fmt.Sprintf("Receptor %s (ID: %s) tem licença com vencimento em %s.", receiver.Name, receiver.ID.String(), receiver.LicenseExpiry.Format("02/01/2006")),
		}
		return uc.repo.Create(alert)
	}
	return nil
}
