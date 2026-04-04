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
		Name:        ToolGetServiceStatus,
		Description: "Read normalized status for a single unit (read-only). Requires manager on PATH and supports_get_service_status in discovery; id is validated per supervisor.",
	}, getServiceStatusHandler(h))
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
