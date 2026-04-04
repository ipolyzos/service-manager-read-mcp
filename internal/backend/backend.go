package backend

import (
	"context"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

// ServiceManagerBackend maps host supervisor CLIs into the shared domain model.
//
// Contract (implementations should follow this so MCP wiring stays predictable):
//   - Context: every method that accepts [context.Context] must return [context.Canceled] /
//     [context.DeadlineExceeded] via ctx.Err() when the context is already done, before
//     doing other work (including returning [svcerr.ErrUnsupported] for unsupported ops).
//   - UnitRef.Manager: treat [domain.ManagerUnknown] or empty as “this backend’s namespace”.
//     If the manager is set to a different backend, return [svcerr.ErrUnsupported] for
//     operations that are namespaced (typically [ServiceManagerBackend.GetServiceStatus] and
//     [ServiceManagerBackend.GetRecentServiceLogs]).
//   - Typed errors: use [svcerr] sentinels where applicable ([svcerr.ErrBackendUnavailable],
//     [svcerr.ErrUnsupported], [svcerr.ErrInvalidArgument], [svcerr.ErrUnitNotFound],
//     [svcerr.ErrInvocation], parse failures as documented per backend).
//   - Capabilities: [ServiceManagerBackend.Capabilities] must match what the implementation
//     actually does on success paths (detection templates in [internal/detect] should match).
//   - Responses: populate [domain.UnitRef.Manager] with this backend’s [domain.ServiceManager]
//     on normalized rows and status objects.
//
// Implementations must be safe for concurrent use or document otherwise.
type ServiceManagerBackend interface {
	DetectServiceManager(ctx context.Context) (domain.ServiceManager, error)
	ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error)
	GetServiceStatus(ctx context.Context, ref domain.UnitRef) (domain.ServiceStatus, error)
	ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error)
	GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error)
	Capabilities() domain.BackendCapabilities
}
