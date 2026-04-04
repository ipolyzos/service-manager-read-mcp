package launchd

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// ParseLaunchctlPrint maps launchctl print stdout into ServiceStatus (best-effort).
func ParseLaunchctlPrint(ref domain.UnitRef, stdout []byte) (domain.ServiceStatus, error) {
	if len(bytes.TrimSpace(stdout)) == 0 {
		return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
	}
	s := string(stdout)
	lower := strings.ToLower(s)
	if strings.Contains(lower, "could not find") || strings.Contains(lower, "not found") {
		return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
	}

	st := domain.ServiceStatus{
		Ref:      ref,
		Active:   domain.ActiveStateUnknown,
		Load:     domain.LoadStateUnknown,
		Metadata: map[string]string{},
	}

	sc := bufio.NewScanner(bytes.NewReader(stdout))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		idx := strings.Index(line, "=")
		if idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			if strings.HasSuffix(val, "{") {
				continue
			}
			switch key {
			case "state":
				st.SubState = val
				switch strings.ToLower(val) {
				case "running":
					st.Active = domain.ActiveStateActive
				case "not running":
					st.Active = domain.ActiveStateInactive
				default:
					st.Active = domain.ActiveStateUnknown
				}
			case "pid":
				if val != "" && val != "-" {
					if pid, err := strconv.ParseInt(val, 10, 64); err == nil && pid > 0 {
						st.MainPID = &pid
					}
				}
			case "last exit code":
				st.Metadata["launchd.last_exit_code"] = val
			case "path", "program":
				if st.Description == "" {
					st.Description = val
				}
			}
		}
	}
	if err := sc.Err(); err != nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: %v", svcerr.ErrParse, err)
	}

	if st.Active == domain.ActiveStateActive {
		st.Load = domain.LoadStateLoaded
	} else if st.Active == domain.ActiveStateInactive {
		st.Load = domain.LoadStateLoaded
	}
	return st, nil
}
