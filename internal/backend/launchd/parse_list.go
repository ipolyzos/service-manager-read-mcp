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

// ListRow is one parsed row from launchctl list.
type ListRow struct {
	PID        string
	LastStatus int
	Label      string
	RawStatus  string
}

// ParseLaunchctlList parses tab-separated launchctl list output (header row optional).
func ParseLaunchctlList(stdout []byte) ([]ListRow, error) {
	if len(stdout) == 0 {
		return nil, nil
	}
	var rows []ListRow
	sc := bufio.NewScanner(bytes.NewReader(stdout))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "pid") && strings.Contains(line, "Label") {
			continue // header
		}
		row, err := parseListLine(line)
		if err != nil {
			return nil, fmt.Errorf("%w: list line %q: %v", svcerr.ErrParse, line, err)
		}
		if row.Label == "" {
			continue
		}
		rows = append(rows, row)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

func parseListLine(line string) (ListRow, error) {
	parts := strings.Split(line, "\t")
	if len(parts) < 3 {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return ListRow{}, fmt.Errorf("expected at least 3 columns")
		}
		parts = []string{fields[0], fields[1], strings.Join(fields[2:], " ")}
	}
	pid := strings.TrimSpace(parts[0])
	statusStr := strings.TrimSpace(parts[1])
	label := strings.TrimSpace(parts[2])
	st, err := strconv.Atoi(statusStr)
	if err != nil {
		return ListRow{}, fmt.Errorf("status %q: %w", statusStr, err)
	}
	return ListRow{PID: pid, LastStatus: st, Label: label, RawStatus: statusStr}, nil
}

// ListRowsToUnits maps list rows to domain units (coarse state from PID only).
func ListRowsToUnits(rows []ListRow) []domain.Unit {
	out := make([]domain.Unit, 0, len(rows))
	for _, r := range rows {
		ref := domain.UnitRef{
			Manager: domain.ManagerLaunchd,
			ID:      r.Label,
		}
		st := domain.StateUnknown
		if r.PID != "-" && r.PID != "" {
			if _, err := strconv.Atoi(r.PID); err == nil {
				st = domain.StateRunning
			}
		} else {
			st = domain.StateStopped
		}
		out = append(out, domain.Unit{Ref: ref, State: st})
	}
	return out
}

// ListRowsToFailedApproximation returns rows whose last exit status is non-zero (read-only heuristic).
func ListRowsToFailedApproximation(rows []ListRow) []domain.FailedServiceEntry {
	var out []domain.FailedServiceEntry
	for _, r := range rows {
		if r.LastStatus == 0 {
			continue
		}
		reason := fmt.Sprintf("last_exit_status=%d", r.LastStatus)
		out = append(out, domain.FailedServiceEntry{
			Ref: domain.UnitRef{
				Manager: domain.ManagerLaunchd,
				ID:      r.Label,
			},
			Reason: reason,
		})
	}
	return out
}
