package systemd

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// ParseShowProperties parses systemctl show key=value lines.
func ParseShowProperties(stdout []byte) (map[string]string, error) {
	if err := ValidateUTF8Lines(stdout, 0); err != nil {
		return nil, fmt.Errorf("%w: %v", svcerr.ErrParse, err)
	}
	m := make(map[string]string)
	sc := bufio.NewScanner(bytes.NewReader(bytes.TrimSpace(stdout)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		m[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%w: scan show output: %v", svcerr.ErrParse, err)
	}
	return m, nil
}

// StatusFromShow builds a ServiceStatus from systemctl show properties for ref.
func StatusFromShow(ref domain.UnitRef, props map[string]string) (domain.ServiceStatus, error) {
	if props == nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: empty properties", svcerr.ErrParse)
	}
	load := strings.ToLower(props["LoadState"])
	if load == "not-found" {
		return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
	}
	st := domain.ServiceStatus{
		Ref:         ref,
		Active:      mapActiveState(props["ActiveState"]),
		SubState:    props["SubState"],
		Load:        mapLoadState(load),
		Description: props["Description"],
		Metadata:    map[string]string{},
	}
	if v, ok := props["UnitFileState"]; ok && v != "" {
		en := strings.EqualFold(v, "enabled") || strings.EqualFold(v, "enabled-runtime") ||
			strings.EqualFold(v, "static") || strings.EqualFold(v, "indirect")
		dis := strings.EqualFold(v, "disabled") || strings.EqualFold(v, "masked")
		switch {
		case en:
			t := true
			st.Enabled = &t
		case dis:
			f := false
			st.Enabled = &f
		}
	}
	if v := strings.TrimSpace(props["MainPID"]); v != "" && v != "0" {
		if pid, err := strconv.ParseInt(v, 10, 64); err == nil {
			st.MainPID = &pid
		}
	}
	if id := strings.TrimSpace(props["Id"]); id != "" {
		st.Metadata["systemd.id"] = id
	}
	if fr := strings.TrimSpace(props["FragmentPath"]); fr != "" {
		st.Metadata["systemd.fragment_path"] = fr
	}
	return st, nil
}

func mapActiveState(s string) domain.ActiveState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "active":
		return domain.ActiveStateActive
	case "inactive":
		return domain.ActiveStateInactive
	case "failed":
		return domain.ActiveStateFailed
	case "activating":
		return domain.ActiveStateActivating
	case "deactivating":
		return domain.ActiveStateDeactivating
	default:
		return domain.ActiveStateUnknown
	}
}

func mapLoadState(s string) domain.LoadState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "loaded":
		return domain.LoadStateLoaded
	case "not-found":
		return domain.LoadStateNotFound
	case "bad-setting":
		return domain.LoadStateBadSetting
	default:
		return domain.LoadStateUnknown
	}
}
