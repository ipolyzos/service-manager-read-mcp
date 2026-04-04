package runit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// Backend implements backend.ServiceManagerBackend using allowlisted `sv` and a filesystem
// scan of the runit service directory (SVDIR or common defaults).
type Backend struct {
	safe *runner.SafeRunner
	// Getenv overrides environment lookup (tests). Nil uses os.Getenv.
	Getenv func(string) string

	invMu     sync.Mutex
	invokeDir string
}

var _ backend.ServiceManagerBackend = (*Backend)(nil)

// New resolves `sv` on PATH.
func New() (*Backend, error) {
	return NewWithBinary("sv")
}

// NewWithBinary allows tests or non-standard install paths.
func NewWithBinary(svBin string) (*Backend, error) {
	if strings.TrimSpace(svBin) == "" {
		return nil, errors.New("runit: binary name required")
	}
	reg, err := runner.NewRegistry(Policies(svBin)...)
	if err != nil {
		return nil, err
	}
	b := &Backend{}
	dr := &dirRunner{b: b, inner: runner.Exec{}}
	b.safe = &runner.SafeRunner{Inner: dr, Registry: reg}
	return b, nil
}

// NewWithSafeRunner uses a SafeRunner whose registry already includes Policies() for the same sv binary.
// It wraps the runner’s Inner with a working-directory injector so `sv status` runs with Dir set to the resolved service tree.
func NewWithSafeRunner(safe *runner.SafeRunner) (*Backend, error) {
	if safe == nil {
		return nil, errors.New("runit: nil safe runner")
	}
	b := &Backend{}
	b.safe = &runner.SafeRunner{
		Inner:    &dirRunner{b: b, inner: safe.Inner},
		Registry: safe.Registry,
	}
	return b, nil
}

func (b *Backend) getenv() func(string) string {
	if b != nil && b.Getenv != nil {
		return b.Getenv
	}
	return os.Getenv
}

// Capabilities implements backend.ServiceManagerBackend.
func (b *Backend) Capabilities() domain.BackendCapabilities {
	return domain.BackendCapabilities{
		Manager:                  domain.ManagerRunit,
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
	if strings.Contains(out, "busybox") {
		return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
	}
	if strings.Contains(out, "runit") || strings.Contains(out, "runsv") {
		return domain.ManagerRunit, nil
	}
	return domain.ManagerUnknown, svcerr.ErrBackendUnavailable
}

// ListUnits implements backend.ServiceManagerBackend.
func (b *Backend) ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dir, err := ResolveServiceDir(b.getenv())
	if err != nil {
		return nil, err
	}
	names, err := ListServiceNames(dir)
	if err != nil {
		return nil, err
	}
	if in.Limit > 0 && len(names) > in.Limit {
		names = names[:in.Limit]
	}
	units := make([]domain.Unit, 0, len(names))
	for _, id := range names {
		units = append(units, domain.Unit{
			Ref:   domain.UnitRef{Manager: domain.ManagerRunit, ID: id},
			State: domain.StateUnknown,
		})
	}
	return units, nil
}

// GetServiceStatus implements backend.ServiceManagerBackend.
func (b *Backend) GetServiceStatus(ctx context.Context, ref domain.UnitRef) (domain.ServiceStatus, error) {
	if ref.Manager != domain.ManagerRunit && ref.Manager != domain.ManagerUnknown && ref.Manager != "" {
		return domain.ServiceStatus{}, svcerr.ErrUnsupported
	}
	if err := validateServiceName(ref.ID); err != nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: %v", svcerr.ErrInvalidArgument, err)
	}
	dir, err := ResolveServiceDir(b.getenv())
	if err != nil {
		return domain.ServiceStatus{}, err
	}
	refN := domain.UnitRef{Manager: domain.ManagerRunit, ID: ref.ID}
	res, err := b.runInDir(ctx, dir, runner.Invocation{
		ID:     cmdStatus,
		Params: []string{ref.ID},
	})
	st, perr := ParseSvStatus(refN, res.Stdout, res.Stderr)
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
		// sv often exits non-zero for down / want-up while still printing status.
		return st, nil
	}
	return st, nil
}

// ListFailedServices implements backend.ServiceManagerBackend.
// It treats services whose `sv status` output includes `down: <name>: …, want up` as failed
// (supervise wants the service up but it is down). This is not equivalent to systemd’s failed units.
func (b *Backend) ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dir, err := ResolveServiceDir(b.getenv())
	if err != nil {
		return nil, err
	}
	names, err := ListServiceNames(dir)
	if err != nil {
		return nil, err
	}
	var out []domain.FailedServiceEntry
	for _, name := range names {
		res, err := b.runInDir(ctx, dir, runner.Invocation{
			ID:     cmdStatus,
			Params: []string{name},
		})
		if err != nil {
			var inv *svcerr.InvocationError
			if errors.As(err, &inv) {
				res = runner.Result{Stdout: inv.Stdout, Stderr: inv.Stderr, ExitCode: inv.ExitCode}
			} else {
				return nil, err
			}
		}
		if SvStatusIndicatesWantUpDown(name, res.Stdout, res.Stderr) {
			out = append(out, domain.FailedServiceEntry{
				Ref:    domain.UnitRef{Manager: domain.ManagerRunit, ID: name},
				Reason: "down with want up (supervise is trying to bring the service up)",
			})
		}
	}
	return out, nil
}

// GetRecentServiceLogs implements backend.ServiceManagerBackend.
func (b *Backend) GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error) {
	_ = ctx
	_ = ref
	_ = in
	return nil, svcerr.ErrUnsupported
}

type dirRunner struct {
	b     *Backend
	inner runner.Runner
}

func (d *dirRunner) Run(ctx context.Context, s runner.Spec) (runner.Result, error) {
	if d != nil && d.b != nil && d.b.invokeDir != "" {
		s.Dir = d.b.invokeDir
	}
	return d.inner.Run(ctx, s)
}

func (b *Backend) runInDir(ctx context.Context, dir string, inv runner.Invocation) (runner.Result, error) {
	b.invMu.Lock()
	defer b.invMu.Unlock()
	b.invokeDir = dir
	defer func() { b.invokeDir = "" }()
	return b.safe.RunInvocation(ctx, inv)
}
