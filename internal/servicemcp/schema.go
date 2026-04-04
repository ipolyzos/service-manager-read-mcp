package servicemcp

import (
	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

// MCP tool names (underscore form per MCP naming conventions).
const (
	ToolListServiceManagers   = "list_service_managers"
	ToolListUnits             = "list_units"
	ToolListFailedServices    = "list_failed_services"
	ToolGetServiceStatus      = "get_service_status"
	ToolGetRecentServiceLogs  = "get_recent_service_logs"
)

// ListServiceManagersParams is the input schema for [ToolListServiceManagers].
type ListServiceManagersParams struct{}

// ListServiceManagersResult is the structured output for [ToolListServiceManagers].
type ListServiceManagersResult struct {
	Report detect.Report `json:"report"`
}

// ListUnitsParams is the input schema for [ToolListUnits].
type ListUnitsParams struct {
	Manager string `json:"manager" jsonschema:"Supervisor: systemd, launchd, openrc, runit, brew_services (or brew)"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Optional maximum units to return; 1–10000, or 0 for backend default"`
}

// ListUnitsResult is the structured output for [ToolListUnits].
type ListUnitsResult struct {
	Units []domain.Unit `json:"units"`
}

// ListFailedServicesParams is the input schema for [ToolListFailedServices].
type ListFailedServicesParams struct {
	Manager string `json:"manager" jsonschema:"Supervisor: systemd, launchd, openrc, runit, brew_services (or brew)"`
}

// ListFailedServicesResult is the structured output for [ToolListFailedServices].
type ListFailedServicesResult struct {
	Failed []domain.FailedServiceEntry `json:"failed"`
}

// GetServiceStatusParams is the input schema for [ToolGetServiceStatus].
type GetServiceStatusParams struct {
	Manager string `json:"manager" jsonschema:"Supervisor: systemd, launchd, openrc, runit, brew_services (or brew)"`
	ID      string `json:"id" jsonschema:"Unit/service id for that supervisor (validated per backend argv rules)"`
}

// GetServiceStatusResult is the structured output for [ToolGetServiceStatus].
type GetServiceStatusResult struct {
	Status domain.ServiceStatus `json:"status"`
}

// GetRecentServiceLogsParams is the input schema for [ToolGetRecentServiceLogs].
type GetRecentServiceLogsParams struct {
	Manager      string `json:"manager" jsonschema:"Supervisor: systemd, launchd, openrc, runit, brew_services (or brew)"`
	Service      string `json:"service" jsonschema:"Unit/service id for that supervisor (same rules as backend argv)"`
	MaxLines     int    `json:"max_lines" jsonschema:"Max lines to return; 0 uses default (100), max 500"`
	RecentWindow string `json:"recent_window,omitempty" jsonschema:"Optional Go duration (e.g. 1h, 30m); empty or 0 = line tail only. Not implemented yet—non-zero returns unsupported."`
}

// GetRecentServiceLogsResult is the structured output for [ToolGetRecentServiceLogs].
type GetRecentServiceLogsResult struct {
	Logs              []domain.RecentLogEntry `json:"logs"`
	EffectiveMaxLines int                     `json:"effective_max_lines"`
	Truncated         bool                    `json:"truncated,omitempty"`
}
