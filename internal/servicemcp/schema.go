package servicemcp

import (
	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

// MCP tool names (underscore form per MCP naming conventions).
const (
	ToolListServiceManagers = "list_service_managers"
	ToolListUnits           = "list_units"
	ToolListFailedServices  = "list_failed_services"
)

// ListServiceManagersParams is the input schema for [ToolListServiceManagers].
type ListServiceManagersParams struct{}

// ListServiceManagersResult is the structured output for [ToolListServiceManagers].
type ListServiceManagersResult struct {
	Report detect.Report `json:"report"`
}

// ListUnitsParams is the input schema for [ToolListUnits].
type ListUnitsParams struct {
	Manager string `json:"manager" jsonschema:"Required. Supervisor key: systemd | launchd | openrc | runit | brew_services | brew. Call list_service_managers first."`
	Limit   int    `json:"limit,omitempty" jsonschema:"Optional. Max units (0 = backend default). Range 0–10000."`
}

// ListUnitsResult is the structured output for [ToolListUnits].
type ListUnitsResult struct {
	Units []domain.Unit `json:"units"`
}

// ListFailedServicesParams is the input schema for [ToolListFailedServices].
type ListFailedServicesParams struct {
	Manager string `json:"manager" jsonschema:"Required. Supervisor key: systemd | launchd | openrc | runit | brew_services | brew. Call list_service_managers first."`
}

// ListFailedServicesResult is the structured output for [ToolListFailedServices].
type ListFailedServicesResult struct {
	Failed []domain.FailedServiceEntry `json:"failed"`
}
