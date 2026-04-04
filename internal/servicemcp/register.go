package servicemcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTools adds read-only supervisor tools to the MCP server.
func RegisterTools(server *mcp.Server, h *Handler) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolListServiceManagers,
		Description: "Discover supervisor CLIs on PATH for this host (read-only; no subprocess). Returns goos and findings with capability flags—call before other tools.",
	}, listServiceManagersHandler(h))

	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolListUnits,
		Description: "List units for one supervisor (read-only). Manager must be on PATH with supports_list_units; use list_service_managers first.",
	}, listUnitsHandler(h))

	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolListFailedServices,
		Description: "List failed or unhealthy service rows for one supervisor (read-only). Needs supports_list_failed_services; semantics differ by backend—see docs/failed-services.md.",
	}, listFailedServicesHandler(h))

	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolGetServiceStatus,
		Description: "Read normalized status for a single unit (read-only). Requires supports_get_service_status; id is validated per supervisor before any backend call.",
	}, getServiceStatusHandler(h))

	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolGetRecentServiceLogs,
		Description: "Bounded, non-streaming service log tail (read-only). Requires supports_get_recent_logs; max_lines capped at 500. No follow mode. recent_window is reserved and currently unsupported—use line tail only. See docs/recent-logs.md.",
	}, getRecentServiceLogsHandler(h))
}

// parseRequiredManager normalizes the manager tool argument. Empty or whitespace-only
// input returns [svcerr.ErrInvalidArgument] with a stable message for every tool.
func parseRequiredManager(raw string) (domain.ServiceManager, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("%w: manager is required", svcerr.ErrInvalidArgument)
	}
	return ParseServiceManager(raw)
}

func listServiceManagersHandler(h *Handler) mcp.ToolHandlerFor[ListServiceManagersParams, ListServiceManagersResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ ListServiceManagersParams) (*mcp.CallToolResult, ListServiceManagersResult, error) {
		rep, err := h.ListServiceManagers(ctx)
		if err != nil {
			return nil, ListServiceManagersResult{}, err
		}
		return nil, ListServiceManagersResult{Report: rep}, nil
	}
}

func listUnitsHandler(h *Handler) mcp.ToolHandlerFor[ListUnitsParams, ListUnitsResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in ListUnitsParams) (*mcp.CallToolResult, ListUnitsResult, error) {
		mgr, err := parseRequiredManager(in.Manager)
		if err != nil {
			return nil, ListUnitsResult{}, err
		}
		units, err := h.ListUnits(ctx, mgr, in.Limit)
		if err != nil {
			return nil, ListUnitsResult{}, err
		}
		return nil, ListUnitsResult{Units: units}, nil
	}
}

func listFailedServicesHandler(h *Handler) mcp.ToolHandlerFor[ListFailedServicesParams, ListFailedServicesResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in ListFailedServicesParams) (*mcp.CallToolResult, ListFailedServicesResult, error) {
		mgr, err := parseRequiredManager(in.Manager)
		if err != nil {
			return nil, ListFailedServicesResult{}, err
		}
		failed, err := h.ListFailedServices(ctx, mgr)
		if err != nil {
			return nil, ListFailedServicesResult{}, err
		}
		return nil, ListFailedServicesResult{Failed: failed}, nil
	}
}

func getServiceStatusHandler(h *Handler) mcp.ToolHandlerFor[GetServiceStatusParams, GetServiceStatusResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in GetServiceStatusParams) (*mcp.CallToolResult, GetServiceStatusResult, error) {
		if strings.TrimSpace(in.Manager) == "" {
			return nil, GetServiceStatusResult{}, fmt.Errorf("%w: manager is required", svcerr.ErrInvalidArgument)
		}
		if strings.TrimSpace(in.ID) == "" {
			return nil, GetServiceStatusResult{}, fmt.Errorf("%w: id is required", svcerr.ErrInvalidArgument)
		}
		mgr, err := ParseServiceManager(in.Manager)
		if err != nil {
			return nil, GetServiceStatusResult{}, err
		}
		st, err := h.GetServiceStatus(ctx, mgr, in.ID)
		if err != nil {
			return nil, GetServiceStatusResult{}, err
		}
		return nil, GetServiceStatusResult{Status: st}, nil
	}
}

func getRecentServiceLogsHandler(h *Handler) mcp.ToolHandlerFor[GetRecentServiceLogsParams, GetRecentServiceLogsResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in GetRecentServiceLogsParams) (*mcp.CallToolResult, GetRecentServiceLogsResult, error) {
		if strings.TrimSpace(in.Manager) == "" {
			return nil, GetRecentServiceLogsResult{}, fmt.Errorf("%w: manager is required", svcerr.ErrInvalidArgument)
		}
		mgr, err := ParseServiceManager(in.Manager)
		if err != nil {
			return nil, GetRecentServiceLogsResult{}, err
		}
		out, err := h.GetRecentServiceLogs(ctx, mgr, in.Service, in.MaxLines, in.RecentWindow)
		if err != nil {
			return nil, GetRecentServiceLogsResult{}, err
		}
		return nil, out, nil
	}
}
