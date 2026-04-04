package brew

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// brewListEntry matches `brew services list --json` rows.
type brewListEntry struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	User     any    `json:"user"`
	File     string `json:"file"`
	ExitCode *int64 `json:"exit_code"`
}

// ParseBrewServicesList decodes list JSON into domain units.
func ParseBrewServicesList(stdout []byte) ([]domain.Unit, error) {
	if err := ValidateUTF8Output(stdout); err != nil {
		return nil, fmt.Errorf("%w: %v", svcerr.ErrParse, err)
	}
	rows, err := decodeListEntries(stdout)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Unit, 0, len(rows))
	for _, row := range rows {
		if row.Name == "" {
			continue
		}
		out = append(out, domain.Unit{
			Ref:   domain.UnitRef{Manager: domain.ManagerBrew, ID: row.Name},
			State: coarseStateFromBrewStatus(row.Status, row.ExitCode),
		})
	}
	return out, nil
}

// ParseBrewServicesListFailed returns entries that look unhealthy from the same list JSON.
// Heuristic: status is "error" (case-insensitive) or exit_code is non-zero.
func ParseBrewServicesListFailed(stdout []byte) ([]domain.FailedServiceEntry, error) {
	if err := ValidateUTF8Output(stdout); err != nil {
		return nil, fmt.Errorf("%w: %v", svcerr.ErrParse, err)
	}
	rows, err := decodeListEntries(stdout)
	if err != nil {
		return nil, err
	}
	var out []domain.FailedServiceEntry
	for _, row := range rows {
		if row.Name == "" {
			continue
		}
		if !rowLooksFailed(row) {
			continue
		}
		out = append(out, domain.FailedServiceEntry{
			Ref:    domain.UnitRef{Manager: domain.ManagerBrew, ID: row.Name},
			Reason: failedReason(row),
		})
	}
	return out, nil
}

func decodeListEntries(stdout []byte) ([]brewListEntry, error) {
	var rows []brewListEntry
	if err := json.Unmarshal(stdout, &rows); err != nil {
		return nil, fmt.Errorf("%w: list json: %v", svcerr.ErrParse, err)
	}
	return rows, nil
}

func coarseStateFromBrewStatus(status string, exitCode *int64) domain.CoarseState {
	// Prefer non-zero exit_code as failed even when status text still says "started"
	// (brew JSON can be internally inconsistent).
	if exitCode != nil && *exitCode != 0 {
		return domain.StateFailed
	}
	s := strings.ToLower(strings.TrimSpace(status))
	switch s {
	case "started":
		return domain.StateRunning
	case "stopped", "none":
		return domain.StateStopped
	case "error":
		return domain.StateFailed
	default:
		if s == "" {
			return domain.StateUnknown
		}
		return domain.StateUnknown
	}
}

func rowLooksFailed(row brewListEntry) bool {
	if strings.EqualFold(strings.TrimSpace(row.Status), "error") {
		return true
	}
	if row.ExitCode != nil && *row.ExitCode != 0 {
		return true
	}
	return false
}

func failedReason(row brewListEntry) string {
	var parts []string
	if st := strings.TrimSpace(row.Status); st != "" {
		parts = append(parts, "status="+st)
	}
	if row.ExitCode != nil {
		parts = append(parts, fmt.Sprintf("exit_code=%d", *row.ExitCode))
	}
	if len(parts) == 0 {
		return "brew services list reported failure"
	}
	return strings.Join(parts, " ")
}
