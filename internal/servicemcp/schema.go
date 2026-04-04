package servicemcp

import (
	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

// MCP tool names (underscore form per MCP naming conventions).
const (
	ToolListServiceManagers = "list_service_managers"
	ToolListUnits           = "list_units"
	ToolGetServiceStatus    = "get_service_status"
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

// GetServiceStatusParams is the input schema for [ToolGetServiceStatus].
type GetServiceStatusParams struct {
	Manager string `json:"manager" jsonschema:"Supervisor: systemd, launchd, openrc, runit, brew_services (or brew)"`
	ID      string `json:"id" jsonschema:"Unit/service id for that supervisor (systemd unit name, launchd label, OpenRC/runit service name, brew formula)"`
}

// GetServiceStatusResult is the structured output for [ToolGetServiceStatus].
type GetServiceStatusResult struct {
	Status domain.ServiceStatus `json:"status"`
}
