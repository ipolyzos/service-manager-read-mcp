package systemd

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// Backend implements backend.ServiceManagerBackend using only allowlisted systemctl/journalctl argv.
type Backend struct {
	safe *runner.SafeRunner
}

var _ backend.ServiceManagerBackend = (*Backend)(nil)

// New resolves systemctl and journalctl on PATH and constructs a SafeRunner over Exec.
func New() (*Backend, error) {
	return NewWithBinaries("systemctl", "journalctl")
}

// NewWithBinaries allows tests or packagers to override binary names.
func NewWithBinaries(systemctlBin, journalctlBin string) (*Backend, error) {
	if strings.TrimSpace(systemctlBin) == "" || strings.TrimSpace(journalctlBin) == "" {
		return nil, errors.New("systemd: binary name required")
	}
	reg, err := runner.NewRegistry(Policies(systemctlBin, journalctlBin)...)
	if err != nil {
		return nil, err
	}
	safe := &runner.SafeRunner{
		Inner:    runner.Exec{},
		Registry: reg,
	}
	return &Backend{safe: safe}, nil
}

// NewWithSafeRunner wires an existing SafeRunner whose registry already contains systemd Policies().
func NewWithSafeRunner(safe *runner.SafeRunner) (*Backend, error) {
	if safe == nil {
		return nil, errors.New("systemd: nil safe runner")
	}
	return &Backend{safe: safe}, nil
}

// Capabilities implements backend.ServiceManagerBackend.
func (b *Backend) Capabilities() domain.BackendCapabilities {
	return domain.BackendCapabilities{
		Manager:                  domain.ManagerSystemd,
		SupportsDetect:           true,
		SupportsListUnits:        true,
		SupportsGetServiceStatus: true,
		SupportsListFailed:       true,
		SupportsRecentLogs:       true,
	}
}

// DetectServiceManager implements backend.ServiceManagerBackend.
func (b *Backend) DetectServiceManager(ctx context.Context) (domain.ServiceManager, error) {
	if err := ctx.Err(); err != nil {
		return domain.ManagerUnknown, err
	}
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdDetect})
	if err != nil {
		return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
	}
	if strings.Contains(string(res.Stdout), "Id=systemd") {
		return domain.ManagerSystemd, nil
	}
	return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
}

// ListUnits implements backend.ServiceManagerBackend.
func (b *Backend) ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdListUnits})
	if err != nil {
		return nil, err
	}
	units, err := ParseListUnitsJSON(res.Stdout)
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
	if err := ctx.Err(); err != nil {
		return domain.ServiceStatus{}, err
	}
	if ref.Manager != domain.ManagerSystemd && ref.Manager != domain.ManagerUnknown && ref.Manager != "" {
		return domain.ServiceStatus{}, svcerr.ErrUnsupported
	}
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{
		ID:     cmdShowUnit,
		Params: []string{ref.ID},
	})
	if err != nil {
		return domain.ServiceStatus{}, err
	}
	props, err := ParseShowProperties(res.Stdout)
	if err != nil {
		return domain.ServiceStatus{}, err
	}
	ref = domain.UnitRef{Manager: domain.ManagerSystemd, ID: ref.ID}
	return StatusFromShow(ref, props)
}

// ListFailedServices implements backend.ServiceManagerBackend.
func (b *Backend) ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdListFailed})
	if err != nil {
		return nil, err
	}
	return ParseFailedServicesJSON(res.Stdout)
}

// GetRecentServiceLogs implements backend.ServiceManagerBackend.
func (b *Backend) GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if ref.Manager != domain.ManagerSystemd && ref.Manager != domain.ManagerUnknown && ref.Manager != "" {
		return nil, svcerr.ErrUnsupported
	}
	if in.Since != nil {
		return nil, svcerr.ErrUnsupported
	}
	n := in.MaxLines
	if n <= 0 {
		n = 100
	}
	if n > 500 {
		n = 500
	}
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{
		ID:     cmdJournalTail,
		Params: []string{strconv.Itoa(n), ref.ID},
	})
	if err != nil {
		return nil, err
	}
	ref = domain.UnitRef{Manager: domain.ManagerSystemd, ID: ref.ID}
	return ParseJournalShortISO(res.Stdout, ref)
}
