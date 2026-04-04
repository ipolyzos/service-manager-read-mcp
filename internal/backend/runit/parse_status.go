package runit

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

var pidRE = regexp.MustCompile(`(?i)\(pid\s+(\d+)\)`)

// ParseSvStatus maps combined `sv status <name>` output into ServiceStatus.
// runit lines look like: run: name: (pid 1) 10s; down: name: 5s, want up; down: name: 1s, normally up
func ParseSvStatus(ref domain.UnitRef, stdout, stderr []byte) (domain.ServiceStatus, error) {
	combined := bytes.TrimSpace(append(append([]byte{}, stdout...), stderr...))
	if err := ValidateUTF8Output(combined); err != nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: %v", svcerr.ErrParse, err)
	}
	low := strings.ToLower(string(combined))
	if low == "" {
		return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
	}
	if looksSvMissingService(low) {
		return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
	}

	st := domain.ServiceStatus{
		Ref:    ref,
		Active: domain.ActiveStateUnknown,
		Load:   domain.LoadStateLoaded,
	}
	if m := pidRE.FindStringSubmatch(string(combined)); len(m) == 2 {
		if pid, err := strconv.ParseInt(m[1], 10, 64); err == nil && pid > 0 {
			st.MainPID = &pid
		}
	}

	// Prefer failure / want-up over plain down when multiple clauses appear on one line.
	switch {
	case strings.Contains(low, "want up") && strings.Contains(low, "down:"):
		st.Active = domain.ActiveStateFailed
		st.SubState = "want_up"
	case strings.Contains(low, "run:"):
		st.Active = domain.ActiveStateActive
		st.SubState = "run"
	case strings.Contains(low, "down:") && strings.Contains(low, "normally up"):
		st.Active = domain.ActiveStateInactive
		st.SubState = "normally_down"
	case strings.Contains(low, "down:"):
		st.Active = domain.ActiveStateInactive
		st.SubState = "down"
	default:
		st.SubState = strings.TrimSpace(string(combined))
	}
	return st, nil
}

func looksSvMissingService(s string) bool {
	switch {
	case strings.Contains(s, "unable to change to service directory"),
		strings.Contains(s, "file does not exist"),
		strings.Contains(s, "no such file"),
		strings.Contains(s, "cannot stat"),
		strings.Contains(s, "does not exist"):
		return true
	default:
		return false
	}
}

// SvStatusIndicatesWantUpDown reports whether output for a given service name includes
// a down stanza with "want up" for that exact service label (safe failed-service heuristic).
func SvStatusIndicatesWantUpDown(serviceName string, stdout, stderr []byte) bool {
	combined := bytes.TrimSpace(append(append([]byte{}, stdout...), stderr...))
	if err := ValidateUTF8Output(combined); err != nil {
		return false
	}
	pat := `(?i)down:\s*` + regexp.QuoteMeta(serviceName) + `:\s*[^\n]*want\s+up`
	re, err := regexp.Compile(pat)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(combined), "\n") {
		line = strings.TrimSpace(line)
		if re.MatchString(line) {
			return true
		}
	}
	return false
}
