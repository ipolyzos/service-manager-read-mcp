package openrc

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// ParseRcServiceList parses `rc-service -l` output (one service name per line).
func ParseRcServiceList(stdout []byte) ([]domain.Unit, error) {
	trim := bytes.TrimSpace(stdout)
	if len(trim) == 0 {
		return nil, nil
	}
	if err := ValidateUTF8Output(stdout); err != nil {
		return nil, fmt.Errorf("%w: %v", svcerr.ErrParse, err)
	}
	var out []domain.Unit
	sc := bufio.NewScanner(bytes.NewReader(trim))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Some versions prefix with noise; keep conservative token.
		name := strings.Fields(line)[0]
		if name == "" {
			continue
		}
		out = append(out, domain.Unit{
			Ref: domain.UnitRef{
				Manager: domain.ManagerOpenRC,
				ID:      name,
			},
			// rc-service -l does not include runtime state; use detail call for truth.
			State: domain.StateUnknown,
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, svcerr.ErrParse
	}
	return out, nil
}
