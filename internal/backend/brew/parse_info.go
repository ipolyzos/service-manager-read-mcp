package brew

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// brewInfoEntry matches `brew services info <formula> --json` array elements.
type brewInfoEntry struct {
	Name        string  `json:"name"`
	ServiceName string  `json:"service_name"`
	Running     bool    `json:"running"`
	Loaded      bool    `json:"loaded"`
	Schedulable bool    `json:"schedulable"`
	Pid         *int64  `json:"pid"`
	ExitCode    *int64  `json:"exit_code"`
	User        any     `json:"user"`
	Status      string  `json:"status"`
	File        string  `json:"file"`
	Registered  bool    `json:"registered"`
	LoadedFile  *string `json:"loaded_file"`
}

// ParseBrewServicesInfo decodes info JSON into a single ServiceStatus.
func ParseBrewServicesInfo(stdout []byte, wantID string) (domain.ServiceStatus, error) {
	if err := ValidateUTF8Output(stdout); err != nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: %v", svcerr.ErrParse, err)
	}
	var rows []brewInfoEntry
	if err := json.Unmarshal(stdout, &rows); err != nil {
		return domain.ServiceStatus{}, fmt.Errorf("%w: info json: %v", svcerr.ErrParse, err)
	}
	if len(rows) == 0 {
		return domain.ServiceStatus{}, svcerr.ErrUnitNotFound
	}
	row := rows[0]
	id := row.Name
	if id == "" {
		id = wantID
	}
	ref := domain.UnitRef{Manager: domain.ManagerBrew, ID: id}
	if row.ServiceName != "" {
		ref.DisplayName = row.ServiceName
	}
	st := domain.ServiceStatus{
		Ref:  ref,
		Load: domain.LoadStateUnknown,
	}
	if row.Loaded {
		st.Load = domain.LoadStateLoaded
	}
	switch {
	case row.Running:
		st.Active = domain.ActiveStateActive
	case strings.EqualFold(strings.TrimSpace(row.Status), "error"):
		st.Active = domain.ActiveStateFailed
	case !row.Running:
		st.Active = domain.ActiveStateInactive
	default:
		st.Active = domain.ActiveStateUnknown
	}
	st.SubState = strings.TrimSpace(row.Status)
	if row.Pid != nil && *row.Pid > 0 {
		p := *row.Pid
		st.MainPID = &p
	}
	if row.File != "" || row.ServiceName != "" || row.LoadedFile != nil {
		st.Metadata = map[string]string{}
		if row.File != "" {
			st.Metadata["file"] = row.File
		}
		if row.ServiceName != "" {
			st.Metadata["service_name"] = row.ServiceName
		}
		if row.LoadedFile != nil && *row.LoadedFile != "" {
			st.Metadata["loaded_file"] = *row.LoadedFile
		}
		if row.ExitCode != nil {
			st.Metadata["exit_code"] = fmt.Sprintf("%d", *row.ExitCode)
		}
	}
	return st, nil
}

// FormulaMissingInBrewServicesInfo stderr/stdout heuristic for unknown formula.
func FormulaMissingInBrewServicesInfo(stderr, stdout []byte) bool {
	s := strings.ToLower(string(stderr) + "\n" + string(stdout))
	return strings.Contains(s, "no available formula") || strings.Contains(s, "no such formula")
}
