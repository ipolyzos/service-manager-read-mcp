package e2etest

import (
	"github.com/ipolyzos/service-manager-read-mcp/internal/servicemcp"
)

// StdioOptions returns [servicemcp.Options] for the e2e MCP server: fixed discovery report
// and [FakeBackend] for systemd.
func StdioOptions() servicemcp.Options {
	return servicemcp.Options{
		DetectHook: DetectHook,
		NewBackend: NewStubBackend,
	}
}
