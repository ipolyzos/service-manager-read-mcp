package openrc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// Backend implements backend.ServiceManagerBackend using allowlisted rc-service / rc-status only.
type Backend struct {
	safe *runner.SafeRunner
}

var _ backend.ServiceManagerBackend = (*Backend)(nil)

// New resolves rc-service and rc-status on PATH.
func New() (*Backend, error) {
	return NewWithBinaries("rc-service", "rc-status")
}

// NewWithBinaries allows tests or custom install paths.
func NewWithBinaries(rcServiceBin, rcStatusBin string) (*Backend, error) {
	if strings.TrimSpace(rcServiceBin) == "" || strings.TrimSpace(rcStatusBin) == "" {
		return nil, errors.New("openrc: binary names required")
	}
	reg, err := runner.NewRegistry(Policies(rcServiceBin, rcStatusBin)...)
	if err != nil {
		return nil, err
	}
	safe := &runner.SafeRunner{
		Inner:    runner.Exec{},
		Registry: reg,
	}
	return &Backend{safe: safe}, nil
}

// NewWithSafeRunner uses a SafeRunner whose registry already includes OpenRC Policies().
func NewWithSafeRunner(safe *runner.SafeRunner) (*Backend, error) {
	if safe == nil {
		return nil, errors.New("openrc: nil safe runner")
	}
	return &Backend{safe: safe}, nil
}

// Capabilities implements backend.ServiceManagerBackend.
// ListFailedServices uses rc-status --crashed (OpenRC ≥ ~0.45). Logs are not safely bound to a service.
func (b *Backend) Capabilities() domain.BackendCapabilities {
	return domain.BackendCapabilities{
		Manager:                  domain.ManagerOpenRC,
		SupportsDetect:           true,
		SupportsListUnits:        true,
		SupportsGetServiceStatus: true,
		SupportsListFailed:       true,
		SupportsRecentLogs:       false,
	}
}

// DetectServiceManager implements backend.ServiceManagerBackend.
func (b *Backend) DetectServiceManager(ctx context.Context) (domain.ServiceManager, error) {
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdDetect})
	if err != nil {
		return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
	}
	out := strings.ToLower(string(res.Stdout))
	if strings.Contains(out, "openrc") {
		return domain.ManagerOpenRC, nil
	}
	return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
}

// ListUnits implements backend.ServiceManagerBackend.
func (b *Backend) ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error) {
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdList})
	if err != nil {
		return nil, err
	}
	units, err := ParseRcServiceList(res.Stdout)
	if err != nil {
		return nil, err
	}
	if in.Limit > 0 && len(units) > in.Limit {
		units = units[:in.Limit]
	}
	return units, nil
}

// GetServiceStatus implements backend.ServiceManagerBackend.
func (b *Backend) GetServiceStatus(ctx context.Context, ref domain.UnitRef) (domain.ServiceStatus, error) {
	if ref.Manager != domain.ManagerOpenRC && ref.Manager != domain.ManagerUnknown && ref.Manager != "" {
		return domain.ServiceStatus{}, svcerr.ErrUnsupported
	}
	if err := validateServiceName(ref.ID); err != nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: %v", svcerr.ErrInvalidArgument, err)
	}
	refN := domain.UnitRef{Manager: domain.ManagerOpenRC, ID: ref.ID}
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{
		ID:     cmdStatus,
		Params: []string{ref.ID},
	})
	st, perr := ParseRcServiceStatus(refN, res.Stdout, res.Stderr)
	if errors.Is(perr, svcerr.ErrUnitNotFound) {
		return domain.ServiceStatus{}, perr
	}
	if perr != nil {
		if err != nil {
			return domain.ServiceStatus{}, err
		}
		return domain.ServiceStatus{}, perr
	}
	if err != nil {
		// OpenRC often uses non-zero exit for "stopped"; still parseable output.
		return st, nil
	}
	return st, nil
}

// ListFailedServices implements backend.ServiceManagerBackend.
func (b *Backend) ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error) {
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdCrashed})
	if err != nil {
		return nil, err
	}
	return ParseRcStatusCrashed(res.Stdout)
}

// GetRecentServiceLogs implements backend.ServiceManagerBackend.
func (b *Backend) GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error) {
	_ = ctx
	_ = ref
	_ = in
	return nil, svcerr.ErrUnsupported
}
