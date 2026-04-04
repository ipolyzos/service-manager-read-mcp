package backend

import (
	"context"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

// ServiceManagerBackend maps host supervisor CLIs into the shared domain model.
// Implementations must be safe for concurrent use or document otherwise.
type ServiceManagerBackend interface {
	DetectServiceManager(ctx context.Context) (domain.ServiceManager, error)
	ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error)
	GetServiceStatus(ctx context.Context, ref domain.UnitRef) (domain.ServiceStatus, error)
	ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error)
	GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error)
	Capabilities() domain.BackendCapabilities
}
