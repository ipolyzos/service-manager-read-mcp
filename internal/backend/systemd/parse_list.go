package systemd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

type listUnitRow struct {
	Unit        string `json:"unit"`
	Load        string `json:"load"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
}

// ParseListUnitsJSON parses systemctl list-units --output=json (array form).
func ParseListUnitsJSON(data []byte) ([]domain.Unit, error) {
	if err := ValidateUTF8Lines(data, 0); err != nil {
		return nil, fmt.Errorf("%w: %v", svcerr.ErrParse, err)
	}
	trim := strings.TrimSpace(string(data))
	if trim == "" {
		return nil, nil
	}
	rows, err := decodeListUnitRows([]byte(trim))
	if err != nil {
		return nil, err
	}
	out := make([]domain.Unit, 0, len(rows))
	for _, row := range rows {
		id := strings.TrimSpace(row.Unit)
		if id == "" {
			continue
		}
		out = append(out, domain.Unit{
			Ref: domain.UnitRef{
				Manager:     domain.ManagerSystemd,
				ID:          id,
				DisplayName: strings.TrimSpace(row.Description),
			},
			State: coarseFromActiveSub(row.Active, row.Sub),
		})
	}
	return out, nil
}

// ParseFailedServicesJSON parses systemctl list-units --state=failed --output=json.
func ParseFailedServicesJSON(data []byte) ([]domain.FailedServiceEntry, error) {
	units, err := ParseListUnitsJSON(data)
	if err != nil {
		return nil, err
	}
	out := make([]domain.FailedServiceEntry, 0, len(units))
	for _, u := range units {
		out = append(out, domain.FailedServiceEntry{
			Ref:    u.Ref,
			Reason: "failed",
		})
	}
	return out, nil
}

func decodeListUnitRows(data []byte) ([]listUnitRow, error) {
	var direct []listUnitRow
	if err := json.Unmarshal(data, &direct); err == nil {
		return direct, nil
	}
	var wrapped struct {
		Units []listUnitRow `json:"units"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, fmt.Errorf("%w: list-units json: %v", svcerr.ErrParse, err)
	}
	return wrapped.Units, nil
}

func coarseFromActiveSub(active, sub string) domain.CoarseState {
	a := strings.ToLower(strings.TrimSpace(active))
	s := strings.ToLower(strings.TrimSpace(sub))
	switch {
	case a == "failed" || s == "failed":
		return domain.StateFailed
	case a == "active" && (s == "running" || s == "listening" || s == "mounted"):
		return domain.StateRunning
	case a == "active" && s == "exited":
		return domain.StateStopped
	case a == "inactive" || a == "dead":
		return domain.StateStopped
	case a == "activating":
		return domain.StateRunning
	default:
		return domain.StateUnknown
	}
}
