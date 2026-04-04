package launchd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// Backend implements backend.ServiceManagerBackend using allowlisted launchctl only.
type Backend struct {
	safe *runner.SafeRunner
}

var _ backend.ServiceManagerBackend = (*Backend)(nil)

// New resolves launchctl on PATH and wires SafeRunner over Exec.
func New() (*Backend, error) {
	return NewWithBinary("launchctl")
}

// NewWithBinary allows tests or packagers to override the binary name.
func NewWithBinary(launchctlBin string) (*Backend, error) {
	if strings.TrimSpace(launchctlBin) == "" {
		return nil, errors.New("launchd: binary name required")
	}
	reg, err := runner.NewRegistry(Policies(launchctlBin)...)
	if err != nil {
		return nil, err
	}
	safe := &runner.SafeRunner{
		Inner:    runner.Exec{},
		Registry: reg,
	}
	return &Backend{safe: safe}, nil
}

// NewWithSafeRunner uses an existing SafeRunner whose registry includes launchd Policies().
func NewWithSafeRunner(safe *runner.SafeRunner) (*Backend, error) {
	if safe == nil {
		return nil, errors.New("launchd: nil safe runner")
	}
	return &Backend{safe: safe}, nil
}

// Capabilities implements backend.ServiceManagerBackend.
// Logs are unsupported: unified logging has no safe allowlisted equivalent to journalctl -u.
func (b *Backend) Capabilities() domain.BackendCapabilities {
	return domain.BackendCapabilities{
		Manager:                  domain.ManagerLaunchd,
		SupportsDetect:           true,
		SupportsListUnits:        true,
		SupportsGetServiceStatus: true,
		SupportsListFailed:       true,
		SupportsRecentLogs:       false,
	}
}

// DetectServiceManager implements backend.ServiceManagerBackend.
func (b *Backend) DetectServiceManager(ctx context.Context) (domain.ServiceManager, error) {
	if runtime.GOOS != "darwin" {
		return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
	}
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdVersion})
	if err != nil {
		return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
	}
	out := strings.ToLower(string(res.Stdout))
	if strings.Contains(out, "launchd") {
		return domain.ManagerLaunchd, nil
	}
	return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
}

// ListUnits implements backend.ServiceManagerBackend.
func (b *Backend) ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error) {
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdList})
	if err != nil {
		return nil, err
	}
	rows, err := ParseLaunchctlList(res.Stdout)
	if err != nil {
		return nil, err
	}
	units := ListRowsToUnits(rows)
	if in.Limit > 0 && len(units) > in.Limit {
		units = units[:in.Limit]
	}
	return units, nil
}

// GetServiceStatus implements backend.ServiceManagerBackend.
func (b *Backend) GetServiceStatus(ctx context.Context, ref domain.UnitRef) (domain.ServiceStatus, error) {
	if ref.Manager != domain.ManagerLaunchd && ref.Manager != domain.ManagerUnknown && ref.Manager != "" {
		return domain.ServiceStatus{}, svcerr.ErrUnsupported
	}
	if err := validateLabel(ref.ID); err != nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: %v", svcerr.ErrInvalidArgument, err)
	}
	ref = domain.UnitRef{Manager: domain.ManagerLaunchd, ID: ref.ID}
	for _, path := range printCandidates(ref.ID) {
		res, err := b.safe.RunInvocation(ctx, runner.Invocation{
			ID:     cmdPrint,
			Params: []string{path},
		})
		if err != nil {
			continue
		}
		st, perr := ParseLaunchctlPrint(ref, res.Stdout)
		if errors.Is(perr, svcerr.ErrUnitNotFound) {
			continue
		}
		if perr != nil {
			return domain.ServiceStatus{}, perr
		}
		return st, nil
	}
	return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
}

// ListFailedServices implements backend.ServiceManagerBackend.
// This is a read-only heuristic: rows from launchctl list whose last exit status is non-zero.
func (b *Backend) ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error) {
	res, err := b.safe.RunInvocation(ctx, runner.Invocation{ID: cmdList})
	if err != nil {
		return nil, err
	}
	rows, err := ParseLaunchctlList(res.Stdout)
	if err != nil {
		return nil, err
	}
	return ListRowsToFailedApproximation(rows), nil
}

// GetRecentServiceLogs implements backend.ServiceManagerBackend.
func (b *Backend) GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error) {
	_ = ctx
	_ = ref
	_ = in
	return nil, svcerr.ErrUnsupported
}

func printCandidates(label string) []string {
	uid := os.Geteuid()
	return []string{
		fmt.Sprintf("gui/%d/%s", uid, label),
		fmt.Sprintf("user/%d/%s", uid, label),
		fmt.Sprintf("system/%s", label),
	}
}
