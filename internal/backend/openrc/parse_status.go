package openrc

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// ParseRcServiceStatus maps `rc-service <name> status` combined output into ServiceStatus.
func ParseRcServiceStatus(ref domain.UnitRef, stdout, stderr []byte) (domain.ServiceStatus, error) {
	combined := bytes.TrimSpace(append(append([]byte{}, stdout...), stderr...))
	if err := ValidateUTF8Output(combined); err != nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: %v", svcerr.ErrParse, err)
	}
	s := strings.ToLower(string(combined))
	if s == "" {
		return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
	}
	if strings.Contains(s, "does not exist") || strings.Contains(s, "not found") || strings.Contains(s, "no such file") {
		return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
	}

	st := domain.ServiceStatus{
		Ref:    ref,
		Active: domain.ActiveStateUnknown,
		Load:   domain.LoadStateLoaded,
	}

	switch {
	case strings.Contains(s, "crashed"):
		st.Active = domain.ActiveStateFailed
		st.SubState = "crashed"
	case strings.Contains(s, "stopped") || strings.Contains(s, "not running") || strings.Contains(s, "is stopped"):
		st.Active = domain.ActiveStateInactive
		st.SubState = "stopped"
	case strings.Contains(s, "started") || strings.Contains(s, "is running") || strings.Contains(s, "running"):
		st.Active = domain.ActiveStateActive
		st.SubState = "started"
	default:
		st.SubState = strings.TrimSpace(string(combined))
	}
	return st, nil
}
