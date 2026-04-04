package stub

import (
	"context"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// Backend is a test double with no real subprocess integration.
type Backend struct {
	Manager   domain.ServiceManager
	Available bool
}

var _ backend.ServiceManagerBackend = (*Backend)(nil)

// Capabilities implements backend.ServiceManagerBackend.
func (b *Backend) Capabilities() domain.BackendCapabilities {
	return domain.BackendCapabilities{
		Manager:                  b.Manager,
		SupportsDetect:           true,
		SupportsListUnits:        b.Available,
		SupportsGetServiceStatus: false,
		SupportsListFailed:       false,
		SupportsRecentLogs:       false,
	}
}

// DetectServiceManager implements backend.ServiceManagerBackend.
func (b *Backend) DetectServiceManager(ctx context.Context) (domain.ServiceManager, error) {
	if !b.Available {
		return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
	}
	if b.Manager == "" {
		return domain.ManagerUnknown, nil
	}
	return b.Manager, nil
}

// ListUnits implements backend.ServiceManagerBackend.
func (b *Backend) ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error) {
	if !b.Available {
		return nil, svcerr.ErrBackendUnavailable
	}
	return nil, nil
}

// GetServiceStatus implements backend.ServiceManagerBackend.
func (b *Backend) GetServiceStatus(ctx context.Context, ref domain.UnitRef) (domain.ServiceStatus, error) {
	if !b.Available {
		return domain.ServiceStatus{}, svcerr.ErrBackendUnavailable
	}
	return domain.ServiceStatus{}, svcerr.ErrUnsupported
}

// ListFailedServices implements backend.ServiceManagerBackend.
func (b *Backend) ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error) {
	if !b.Available {
		return nil, svcerr.ErrBackendUnavailable
	}
	return nil, svcerr.ErrUnsupported
}

// GetRecentServiceLogs implements backend.ServiceManagerBackend.
func (b *Backend) GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error) {
	if !b.Available {
		return nil, svcerr.ErrBackendUnavailable
	}
	return nil, svcerr.ErrUnsupported
}
