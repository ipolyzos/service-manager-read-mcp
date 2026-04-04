package servicemcp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend"
	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/brew"
	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/launchd"
	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/openrc"
	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/runit"
	"github.com/ipolyzos/service-manager-read-mcp/internal/backend/systemd"
	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// MaxListUnitsLimit is the upper bound accepted for [ListUnitsParams.Limit].
const MaxListUnitsLimit = 10000

// Options configures [Handler] behavior.
type Options struct {
	// DetectEnv is passed to [detect.Detect] (zero value uses runtime GOOS and exec.LookPath).
	DetectEnv detect.Environment
	// NewBackend overrides backend construction (tests). Nil uses the default registry.
	NewBackend func(domain.ServiceManager) (backend.ServiceManagerBackend, error)
	// DetectHook replaces [detect.Detect] for handler routing (tests). Nil uses [detect.Detect].
	DetectHook func(context.Context, detect.Environment) detect.Report
}

// Handler bridges MCP tools to [detect] and [backend.ServiceManagerBackend] without executing ad-hoc shell.
type Handler struct {
	env        detect.Environment
	newBackend func(domain.ServiceManager) (backend.ServiceManagerBackend, error)
	detectHook func(context.Context, detect.Environment) detect.Report
}

// NewHandler constructs a Handler with the given options.
func NewHandler(opt Options) *Handler {
	return &Handler{
		env:        opt.DetectEnv,
		newBackend: opt.NewBackend,
		detectHook: opt.DetectHook,
	}
}

func (h *Handler) detectReport(ctx context.Context) detect.Report {
	if h.detectHook != nil {
		return h.detectHook(ctx, h.env)
	}
	return detect.Detect(ctx, h.env)
}

// requireManager checks the detection report: known manager, available on host, and capability predicate.
func (h *Handler) requireManager(ctx context.Context, manager domain.ServiceManager, want func(domain.BackendCapabilities) bool, feature string) error {
	rep := h.detectReport(ctx)
	var found *detect.Finding
	for i := range rep.Findings {
		if rep.Findings[i].Manager == manager {
			found = &rep.Findings[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("%w: unknown manager %q", svcerr.ErrInvalidArgument, manager)
	}
	if !found.Applicable || !found.Present {
		return fmt.Errorf("%w: manager %q is not available on this host", svcerr.ErrBackendUnavailable, manager)
	}
	if !want(found.Capabilities) {
		return fmt.Errorf("%w: %q does not support %s on this host", svcerr.ErrUnsupported, manager, feature)
	}
	return nil
}

func (h *Handler) backendFor(m domain.ServiceManager) (backend.ServiceManagerBackend, error) {
	if h.newBackend != nil {
		return h.newBackend(m)
	}
	return openBackend(m)
}

// ListServiceManagers runs PATH-based supervisor discovery (read-only, no subprocess for detection).
func (h *Handler) ListServiceManagers(ctx context.Context) (detect.Report, error) {
	if err := ctx.Err(); err != nil {
		return detect.Report{}, err
	}
	return h.detectReport(ctx), nil
}

// ListUnits lists units using the library backend for the given manager.
// The manager must be applicable, present on PATH, and advertise list support in the detection report.
func (h *Handler) ListUnits(ctx context.Context, manager domain.ServiceManager, limit int) ([]domain.Unit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit < 0 || limit > MaxListUnitsLimit {
		return nil, fmt.Errorf("%w: limit must be between 0 and %d", svcerr.ErrInvalidArgument, MaxListUnitsLimit)
	}
	if err := h.requireManager(ctx, manager, func(c domain.BackendCapabilities) bool {
		return c.SupportsListUnits
	}, "list_units"); err != nil {
		return nil, err
	}
	be, err := h.backendFor(manager)
	if err != nil {
		return nil, err
	}
	return be.ListUnits(ctx, domain.ListUnitsInput{Limit: limit})
}

// ListFailedServices returns normalized failure rows for one manager. Entries are sorted by
// (ref.manager, ref.id) for stable MCP output; semantics of “failed” differ by backend (see docs/failed-services.md).
func (h *Handler) ListFailedServices(ctx context.Context, manager domain.ServiceManager) ([]domain.FailedServiceEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := h.requireManager(ctx, manager, func(c domain.BackendCapabilities) bool {
		return c.SupportsListFailed
	}, "list_failed_services"); err != nil {
		return nil, err
	}
	be, err := h.backendFor(manager)
	if err != nil {
		return nil, err
	}
	entries, err := be.ListFailedServices(ctx)
	if err != nil {
		return nil, err
	}
	return sortFailedServiceEntries(entries), nil
}

// GetRecentServiceLogs returns a bounded, non-streaming tail of log lines for one service.
// No follow mode and no unbounded reads: max_lines is clamp-validated by [ParseRecentLogsBounds];
// outputs may be truncated defensively if a backend returns more lines than requested.
// A non-zero recent_window is rejected with [svcerr.ErrUnsupported] until a safe backend mapping exists.
func (h *Handler) GetRecentServiceLogs(ctx context.Context, manager domain.ServiceManager, service string, maxLines int, recentWindow string) (GetRecentServiceLogsResult, error) {
	var empty GetRecentServiceLogsResult
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	bounds, err := ParseRecentLogsBounds(manager, service, maxLines, recentWindow)
	if err != nil {
		return empty, err
	}
	if bounds.Window > 0 {
		return empty, fmt.Errorf("%w: recent_window is not implemented; use empty or 0 for line-bounded tail only", svcerr.ErrUnsupported)
	}
	if err := h.requireManager(ctx, manager, func(c domain.BackendCapabilities) bool {
		return c.SupportsRecentLogs
	}, "get_recent_service_logs"); err != nil {
		return empty, err
	}
	be, err := h.backendFor(manager)
	if err != nil {
		return empty, err
	}
	ref := domain.UnitRef{Manager: manager, ID: bounds.Service}
	entries, err := be.GetRecentServiceLogs(ctx, ref, domain.RecentLogsInput{MaxLines: bounds.MaxLines})
	if err != nil {
		return empty, err
	}
	truncated := false
	if len(entries) > bounds.MaxLines {
		entries = entries[len(entries)-bounds.MaxLines:]
		truncated = true
	}
	return GetRecentServiceLogsResult{
		Logs:              entries,
		EffectiveMaxLines: bounds.MaxLines,
		Truncated:         truncated,
	}, nil
}

func sortFailedServiceEntries(entries []domain.FailedServiceEntry) []domain.FailedServiceEntry {
	if len(entries) < 2 {
		return entries
	}
	out := make([]domain.FailedServiceEntry, len(entries))
	copy(out, entries)
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Ref, out[j].Ref
		sa, sb := string(a.Manager), string(b.Manager)
		if sa != sb {
			return sa < sb
		}
		if a.ID != b.ID {
			return a.ID < b.ID
		}
		return a.DisplayName < b.DisplayName
	})
	return out
}

// ParseServiceManager normalizes user-supplied manager strings to [domain.ServiceManager].
func ParseServiceManager(s string) (domain.ServiceManager, error) {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "systemd":
		return domain.ManagerSystemd, nil
	case "launchd":
		return domain.ManagerLaunchd, nil
	case "openrc":
		return domain.ManagerOpenRC, nil
	case "runit":
		return domain.ManagerRunit, nil
	case "brew", "brew_services":
		return domain.ManagerBrew, nil
	default:
		return "", fmt.Errorf("%w: manager must be systemd, launchd, openrc, runit, or brew_services", svcerr.ErrInvalidArgument)
	}
}

func openBackend(m domain.ServiceManager) (backend.ServiceManagerBackend, error) {
	switch m {
	case domain.ManagerSystemd:
		return systemd.New()
	case domain.ManagerLaunchd:
		return launchd.New()
	case domain.ManagerOpenRC:
		return openrc.New()
	case domain.ManagerRunit:
		return runit.New()
	case domain.ManagerBrew:
		return brew.New()
	default:
		return nil, fmt.Errorf("%w: unsupported manager %q", svcerr.ErrInvalidArgument, m)
	}
}
