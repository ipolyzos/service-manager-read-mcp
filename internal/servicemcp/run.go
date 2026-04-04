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
		Instructions: "Read-only access to host service supervisors. Tools never start, stop, or reconfigure services.",
		Logger:       slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelError})),
	})
	RegisterTools(s, h)
	return s.Run(ctx, &mcp.StdioTransport{})
}
