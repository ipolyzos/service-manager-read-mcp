package servicemcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// Bounds for MCP get_recent_service_logs (line-bounded tail; no streaming/follow).
const (
	MaxRecentLogLines     = 500
	DefaultRecentLogLines = 100
)

// MaxRecentLogWindow caps parsed recent_window duration strings (e.g. 168h = 7d).
const MaxRecentLogWindow = 168 * time.Hour

// ParsedRecentLogsBounds is validated input ready for the backend after parsing.
type ParsedRecentLogsBounds struct {
	Service  string
	MaxLines int
	Window   time.Duration // 0 means no time filter at the MCP layer
}

// ParseRecentLogsBounds validates manager-specific service id, max_lines, and recent_window.
// recent_window uses Go duration syntax (e.g. 1h, 30m); empty or "0" means no window.
func ParseRecentLogsBounds(manager domain.ServiceManager, service string, maxLines int, recentWindow string) (ParsedRecentLogsBounds, error) {
	service = strings.TrimSpace(service)
	if err := validateServiceIdentifier(manager, service); err != nil {
		return ParsedRecentLogsBounds{}, err
	}
	if maxLines < 0 {
		return ParsedRecentLogsBounds{}, fmt.Errorf("%w: max_lines must be >= 0", svcerr.ErrInvalidArgument)
	}
	if maxLines > MaxRecentLogLines {
		return ParsedRecentLogsBounds{}, fmt.Errorf("%w: max_lines must be <= %d", svcerr.ErrInvalidArgument, MaxRecentLogLines)
	}
	effective := maxLines
	if effective == 0 {
		effective = DefaultRecentLogLines
	}

	s := strings.TrimSpace(recentWindow)
	if s == "" || s == "0" {
		return ParsedRecentLogsBounds{Service: service, MaxLines: effective, Window: 0}, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return ParsedRecentLogsBounds{}, fmt.Errorf("%w: recent_window: invalid duration: %v", svcerr.ErrInvalidArgument, err)
	}
	if d <= 0 {
		return ParsedRecentLogsBounds{}, fmt.Errorf("%w: recent_window must be positive when set", svcerr.ErrInvalidArgument)
	}
	if d > MaxRecentLogWindow {
		return ParsedRecentLogsBounds{}, fmt.Errorf("%w: recent_window exceeds maximum allowed window", svcerr.ErrInvalidArgument)
	}
	return ParsedRecentLogsBounds{Service: service, MaxLines: effective, Window: d}, nil
}
