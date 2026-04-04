package openrc

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

// ParseRcStatusCrashed parses `rc-status --crashed` output (one service per line).
func ParseRcStatusCrashed(stdout []byte) ([]domain.FailedServiceEntry, error) {
	if err := ValidateUTF8Output(stdout); err != nil {
		return nil, err
	}
	var out []domain.FailedServiceEntry
	sc := bufio.NewScanner(bytes.NewReader(bytes.TrimSpace(stdout)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name := strings.Fields(line)[0]
		if name == "" {
			continue
		}
		out = append(out, domain.FailedServiceEntry{
			Ref: domain.UnitRef{
				Manager: domain.ManagerOpenRC,
				ID:      name,
			},
			Reason: "crashed",
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
