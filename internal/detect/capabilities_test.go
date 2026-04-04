package detect

import (
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

// capabilityTemplate must stay aligned with concrete backend packages under internal/backend/*.
func TestCapabilityTemplate_presentMatchesBackendImplementations(t *testing.T) {
	t.Parallel()
	cases := []struct {
		mgr               domain.ServiceManager
		wantListFailed    bool
		wantRecentLogs    bool
		wantList, wantGet bool
	}{
		{domain.ManagerSystemd, true, true, true, true},
		{domain.ManagerLaunchd, true, false, true, true},
		{domain.ManagerOpenRC, true, false, true, true},
		{domain.ManagerRunit, true, false, true, true},
		{domain.ManagerBrew, true, false, true, true},
	}
	for _, tc := range cases {
		c := capabilityTemplate(tc.mgr, true)
		if c.SupportsListFailed != tc.wantListFailed {
			t.Fatalf("%s SupportsListFailed: got %v want %v", tc.mgr, c.SupportsListFailed, tc.wantListFailed)
		}
		if c.SupportsRecentLogs != tc.wantRecentLogs {
			t.Fatalf("%s SupportsRecentLogs: got %v want %v", tc.mgr, c.SupportsRecentLogs, tc.wantRecentLogs)
		}
		if c.SupportsListUnits != tc.wantList || c.SupportsGetServiceStatus != tc.wantGet {
			t.Fatalf("%s: %#v", tc.mgr, c)
		}
	}
}
