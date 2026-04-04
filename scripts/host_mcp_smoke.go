// Command host_mcp_smoke exercises the installed MCP binary against the real host (journalctl/systemd).
//
// Usage (from repo root):
//
//	go run ./scripts/host_mcp_smoke.go [-q] [/path/to/service-manager-read-mcp]
//
// Default binary: SERVICE_MANAGER_READ_MCP_BIN, else service-manager-read-mcp on PATH, else $HOME/go/bin/service-manager-read-mcp.
// Use -q for one OK/FAIL line per tool (chat-friendly).
//
//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func defaultBinaryPath() string {
	if e := os.Getenv("SERVICE_MANAGER_READ_MCP_BIN"); e != "" {
		return e
	}
	if p, err := exec.LookPath("service-manager-read-mcp"); err == nil {
		return p
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, "go", "bin", "service-manager-read-mcp")
	}
	return "service-manager-read-mcp"
}

func main() {
	bin := defaultBinaryPath()
	quiet := false
	switch {
	case len(os.Args) > 1 && os.Args[1] == "-q":
		quiet = true
		if len(os.Args) > 2 {
			bin = os.Args[2]
		}
	case len(os.Args) > 1:
		bin = os.Args[1]
	}
	ctx := context.Background()
	cmd := exec.Command(bin)
	client := mcp.NewClient(&mcp.Implementation{Name: "host-smoke", Version: "0"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	defer func() { _ = session.Close() }()

	tests := []struct {
		name string
		tool string
		args map[string]any
	}{
		{"list_service_managers", "list_service_managers", map[string]any{}},
		{"list_units (systemd, limit 3)", "list_units", map[string]any{
			"manager": "systemd", "limit": 3,
		}},
		{"list_failed_services (systemd)", "list_failed_services", map[string]any{
			"manager": "systemd",
		}},
		{"get_service_status (chronyd)", "get_service_status", map[string]any{
			"manager": "systemd", "id": "chronyd.service",
		}},
		{"sshd logs (empty ok)", "get_recent_service_logs", map[string]any{
			"manager": "systemd", "service": "sshd.service", "max_lines": 25,
		}},
		{"chronyd max_lines 0", "get_recent_service_logs", map[string]any{
			"manager": "systemd", "service": "chronyd.service", "max_lines": 0,
		}},
		{"systemd-journald tail", "get_recent_service_logs", map[string]any{
			"manager": "systemd", "service": "systemd-journald.service", "max_lines": 15,
		}},
	}

	ok := true
	for _, tc := range tests {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tc.tool, Arguments: tc.args})
		if err != nil {
			fmt.Printf("FAIL %s: %v\n", tc.name, err)
			ok = false
			continue
		}
		if res.IsError {
			fmt.Printf("FAIL %s: tool error %#v\n", tc.name, res)
			ok = false
			continue
		}
		if quiet {
			fmt.Printf("OK %s\n", tc.name)
			continue
		}
		b, _ := json.MarshalIndent(res.StructuredContent, "", "  ")
		fmt.Printf("OK %s\n%s\n\n", tc.name, string(b))
	}
	if !ok {
		os.Exit(1)
	}
}
