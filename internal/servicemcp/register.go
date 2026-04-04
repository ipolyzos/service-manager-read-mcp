package servicemcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTools adds read-only supervisor tools to the MCP server.
func RegisterTools(server *mcp.Server, h *Handler) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolListServiceManagers,
		Description: "List service supervisor integrations for this host (PATH-based discovery only; read-only). Each finding includes coarse capability flags.",
	}, listServiceManagersHandler(h))

	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolListUnits,
		Description: "List units for one supervisor (read-only). Use list_service_managers to see which managers are available; manager must be present on PATH.",
	}, listUnitsHandler(h))

	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolListFailedServices,
		Description: "List failed or unhealthy services for one supervisor (read-only). Requires supports_list_failed_services in discovery; meaning of 'failed' varies by manager—see project docs.",
	}, listFailedServicesHandler(h))

	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolGetRecentServiceLogs,
		Description: "Bounded, non-streaming service log tail (read-only). Requires supports_get_recent_logs; max_lines capped at 500. No follow mode. recent_window is reserved and currently unsupported—use line tail only. See docs/recent-logs.md.",
	}, getRecentServiceLogsHandler(h))
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
		if strings.TrimSpace(in.Manager) == "" {
			return nil, ListUnitsResult{}, fmt.Errorf("%w: manager is required", svcerr.ErrInvalidArgument)
		}
		mgr, err := ParseServiceManager(in.Manager)
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
		if strings.TrimSpace(in.Manager) == "" {
			return nil, ListFailedServicesResult{}, fmt.Errorf("%w: manager is required", svcerr.ErrInvalidArgument)
		}
		mgr, err := ParseServiceManager(in.Manager)
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
