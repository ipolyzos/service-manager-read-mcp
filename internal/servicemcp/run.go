package servicemcp

import (
	"context"
	"io"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultServerName    = "service-manager-read-mcp"
	defaultServerVersion = "0.1.0"
)

// RunStdio serves MCP over stdin/stdout using the official SDK stdio transport.
// Logs go to logWriter (typically os.Stderr); protocol traffic uses stdout only.
func RunStdio(ctx context.Context, logWriter io.Writer, opt Options) error {
	h := NewHandler(opt)
	s := mcp.NewServer(&mcp.Implementation{Name: defaultServerName, Version: defaultServerVersion}, &mcp.ServerOptions{
		Instructions: "Read-only host supervisor tools: discovery first (list_service_managers), then per-manager calls. Errors are typed (invalid argument, unavailable, unsupported, parse, invocation). No start/stop/reconfigure.",
		Logger:       slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelError})),
	})
	RegisterTools(s, h)
	return s.Run(ctx, &mcp.StdioTransport{})
}
