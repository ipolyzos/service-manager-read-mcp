package detect

import "github.com/ipolyzos/service-manager-read-mcp/internal/domain"

// capabilityTemplate returns planned backend features once a CLI is visible on PATH.
// Absent managers return a zero capability block with SupportsDetect false.
func capabilityTemplate(m domain.ServiceManager, present bool) domain.BackendCapabilities {
	if !present {
		return domain.BackendCapabilities{Manager: m}
	}
	caps := domain.BackendCapabilities{
		Manager:        m,
		SupportsDetect: true,
	}
	switch m {
	case domain.ManagerSystemd:
		caps.SupportsListUnits = true
		caps.SupportsGetServiceStatus = true
		caps.SupportsListFailed = true
		caps.SupportsRecentLogs = true
	case domain.ManagerLaunchd:
		caps.SupportsListUnits = true
		caps.SupportsGetServiceStatus = true
		caps.SupportsListFailed = true
		caps.SupportsRecentLogs = false
	case domain.ManagerRunit:
		caps.SupportsListUnits = true
		caps.SupportsGetServiceStatus = true
		caps.SupportsListFailed = true
		caps.SupportsRecentLogs = false
	case domain.ManagerOpenRC, domain.ManagerBrew:
		caps.SupportsListUnits = true
		caps.SupportsGetServiceStatus = true
		caps.SupportsListFailed = true
		caps.SupportsRecentLogs = false
	default:
		caps.SupportsDetect = false
	}
	return caps
}
