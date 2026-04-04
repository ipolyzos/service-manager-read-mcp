package e2etest

import (
	"context"
	"fmt"
	"sort"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// FakeBackend is a deterministic [backend.ServiceManagerBackend] for MCP stdio e2e tests.
type FakeBackend struct{}

var _ backend.ServiceManagerBackend = (*FakeBackend)(nil)

// NewStubBackend returns [FakeBackend] for systemd only (e2e harness).
func NewStubBackend(m domain.ServiceManager) (backend.ServiceManagerBackend, error) {
	if m != domain.ManagerSystemd {
		return nil, fmt.Errorf("%w: e2e stub only supports systemd", svcerr.ErrInvalidArgument)
	}
	return &FakeBackend{}, nil
}

func (b *FakeBackend) Capabilities() domain.BackendCapabilities {
	return DeterministicReport().Findings[0].Capabilities
}

func (b *FakeBackend) DetectServiceManager(ctx context.Context) (domain.ServiceManager, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return domain.ManagerSystemd, nil
}

func (b *FakeBackend) ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	all := []domain.Unit{
		{
			Ref:   domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx.service"},
			State: domain.StateRunning,
		},
		{
			Ref:   domain.UnitRef{Manager: domain.ManagerSystemd, ID: "ssh.service"},
			State: domain.StateStopped,
		},
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Ref.ID < all[j].Ref.ID })
	if in.Limit > 0 && len(all) > in.Limit {
		all = all[:in.Limit]
	}
	return all, nil
}

func (b *FakeBackend) GetServiceStatus(ctx context.Context, ref domain.UnitRef) (domain.ServiceStatus, error) {
	if err := ctx.Err(); err != nil {
		return domain.ServiceStatus{}, err
	}
	if ref.Manager != "" && ref.Manager != domain.ManagerSystemd {
		return domain.ServiceStatus{}, svcerr.ErrUnsupported
	}
	if ref.ID != "nginx.service" {
		return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
	}
	enabled := true
	pid := int64(4242)
	return domain.ServiceStatus{
		Ref:         domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx.service"},
		Active:      domain.ActiveStateActive,
		SubState:    "running",
		Load:        domain.LoadStateLoaded,
		Enabled:     &enabled,
		MainPID:     &pid,
		Description: "e2e stub nginx",
		Metadata:    map[string]string{"e2e": "true"},
	}, nil
}

func (b *FakeBackend) ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Unsorted on purpose; [servicemcp.Handler] sorts for stable MCP output.
	return []domain.FailedServiceEntry{
		{Ref: domain.UnitRef{Manager: domain.ManagerSystemd, ID: "z-failed.service"}, Reason: "z-reason"},
		{Ref: domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx.service"}, Reason: "e2e-failed"},
	}, nil
}

func (b *FakeBackend) GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if ref.Manager != "" && ref.Manager != domain.ManagerSystemd {
		return nil, svcerr.ErrUnsupported
	}
	if ref.ID != "nginx.service" {
		return nil, svcerr.ErrUnitNotFound
	}
	max := in.MaxLines
	if max <= 0 {
		max = 100
	}
	lines := []domain.RecentLogEntry{
		{Line: "[e2e] log line 1", Severity: domain.LogSeverityInfo},
		{Line: "[e2e] log line 2", Severity: domain.LogSeverityWarning},
		{Line: "[e2e] log line 3", Severity: domain.LogSeverityError},
	}
	if len(lines) > max {
		lines = lines[len(lines)-max:]
	}
	return lines, nil
}
