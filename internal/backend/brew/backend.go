package brew

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

// Backend implements backend.ServiceManagerBackend using allowlisted `brew` invocations only.
type Backend struct {
	safe *runner.SafeRunner
}

var _ backend.ServiceManagerBackend = (*Backend)(nil)

// New resolves `brew` on PATH.
func New() (*Backend, error) {
	return NewWithBinary("brew")
}

// NewWithBinary allows tests or custom install paths.
func NewWithBinary(brewBin string) (*Backend, error) {
	if strings.TrimSpace(brewBin) == "" {
		return nil, errors.New("brew: binary name required")
	}
	reg, err := runner.NewRegistry(Policies(brewBin)...)
	if err != nil {
		return nil, err
	}
	safe := &runner.SafeRunner{
		Inner:    runner.Exec{},
		Registry: reg,
	}
	return &Backend{safe: safe}, nil
}

// NewWithSafeRunner uses a SafeRunner whose registry already includes Policies() for the same brew binary.
func NewWithSafeRunner(safe *runner.SafeRunner) (*Backend, error) {
	if safe == nil {
		return nil, errors.New("brew: nil safe runner")
	}
	return &Backend{safe: safe}, nil
}

// Capabilities implements backend.ServiceManagerBackend.
func (b *Backend) Capabilities() domain.BackendCapabilities {
	return domain.BackendCapabilities{
		Manager:                  domain.ManagerBrew,
		SupportsDetect:           true,
		SupportsListUnits:        true,
		SupportsGetServiceStatus: true,
		SupportsListFailed:       true,
		SupportsRecentLogs:       false,
	}
}

// DetectServiceManager implements backend.ServiceManagerBackend.
func (b *Backend) DetectServiceManager(ctx context.Context) (domain.ServiceManager, error) {
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdVersion})
	if err != nil {
		return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
	}
	out := strings.ToLower(string(res.Stdout))
	if strings.Contains(out, "homebrew") {
		return domain.ManagerBrew, nil
	}
	return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
}

// ListUnits implements backend.ServiceManagerBackend.
func (b *Backend) ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error) {
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdList})
	if err != nil {
		return nil, err
	}
	units, err := ParseBrewServicesList(res.Stdout)
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
	if ref.Manager != domain.ManagerBrew && ref.Manager != domain.ManagerUnknown && ref.Manager != "" {
		return domain.ServiceStatus{}, svcerr.ErrUnsupported
	}
	if err := validateFormulaName(ref.ID); err != nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: %v", svcerr.ErrInvalidArgument, err)
	}
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{
		ID:     cmdInfo,
		Params: []string{ref.ID},
	})
	if err != nil {
		var inv *svcerr.InvocationError
		if errors.As(err, &inv) && FormulaMissingInBrewServicesInfo(inv.Stderr, inv.Stdout) {
			return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
		}
		return domain.ServiceStatus{}, err
	}
	return ParseBrewServicesInfo(res.Stdout, ref.ID)
}

// ListFailedServices implements backend.ServiceManagerBackend.
func (b *Backend) ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error) {
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdList})
	if err != nil {
		return nil, err
	}
	return ParseBrewServicesListFailed(res.Stdout)
}

// GetRecentServiceLogs implements backend.ServiceManagerBackend.
func (b *Backend) GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error) {
	_ = ctx
	_ = ref
	_ = in
	return nil, svcerr.ErrUnsupported
}
