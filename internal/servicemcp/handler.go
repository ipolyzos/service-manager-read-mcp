package servicemcp

import (
	"context"
	"fmt"
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
}

// Handler bridges MCP tools to [detect] and [backend.ServiceManagerBackend] without executing ad-hoc shell.
type Handler struct {
	env detect.Environment
}

// NewHandler constructs a Handler with the given options.
func NewHandler(opt Options) *Handler {
	return &Handler{env: opt.DetectEnv}
}

// ListServiceManagers runs PATH-based supervisor discovery (read-only, no subprocess for detection).
func (h *Handler) ListServiceManagers(ctx context.Context) (detect.Report, error) {
	if err := ctx.Err(); err != nil {
		return detect.Report{}, err
	}
	return detect.Detect(ctx, h.env), nil
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
	rep := detect.Detect(ctx, h.env)
	var found *detect.Finding
	for i := range rep.Findings {
		if rep.Findings[i].Manager == manager {
			found = &rep.Findings[i]
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("%w: unknown manager %q", svcerr.ErrInvalidArgument, manager)
	}
	if !found.Applicable || !found.Present || !found.Capabilities.SupportsListUnits {
		return nil, fmt.Errorf("%w: manager %q is not available on this host", svcerr.ErrBackendUnavailable, manager)
	}
	be, err := openBackend(manager)
	if err != nil {
		return nil, err
	}
	return be.ListUnits(ctx, domain.ListUnitsInput{Limit: limit})
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
