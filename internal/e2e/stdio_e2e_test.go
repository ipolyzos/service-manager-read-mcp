package e2e

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/servicemcp"
	"github.com/ipolyzos/service-manager-read-mcp/internal/servicemcp/e2etest"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestStdioMCPServer_toolFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	bin := buildE2EMCPServer(t)

	cmd := exec.Command(bin)
	client := mcp.NewClient(&mcp.Implementation{Name: "e2e-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })

	wantReport := e2etest.DeterministicReport()
	t.Run("list_service_managers", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      servicemcp.ToolListServiceManagers,
			Arguments: map[string]any{},
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.IsError {
			t.Fatalf("tool error: %#v", res)
		}
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		var out servicemcp.ListServiceManagersResult
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		if out.Report.GOOS != wantReport.GOOS || len(out.Report.Findings) != len(wantReport.Findings) {
			t.Fatalf("report mismatch: %#v", out.Report)
		}
		got := out.Report.Findings[0]
		wf := wantReport.Findings[0]
		if got.Manager != wf.Manager || !got.Applicable || !got.Present || got.ResolvedPath != wf.ResolvedPath {
			t.Fatalf("finding mismatch: %#v", got)
		}
		c := got.Capabilities
		if !c.SupportsListUnits || !c.SupportsGetServiceStatus || !c.SupportsListFailed || !c.SupportsRecentLogs {
			t.Fatalf("capabilities: %#v", c)
		}
	})

	t.Run("list_units", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: servicemcp.ToolListUnits,
			Arguments: map[string]any{
				"manager": "systemd",
				"limit":   0,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.IsError {
			t.Fatalf("tool error: %#v", res)
		}
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		var out servicemcp.ListUnitsResult
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		if len(out.Units) != 2 {
			t.Fatalf("units: %#v", out.Units)
		}
		if out.Units[0].Ref.ID != "nginx.service" || out.Units[0].State != domain.StateRunning {
			t.Fatalf("unit0: %#v", out.Units[0])
		}
		if out.Units[1].Ref.ID != "ssh.service" || out.Units[1].State != domain.StateStopped {
			t.Fatalf("unit1: %#v", out.Units[1])
		}
	})

	t.Run("list_units_limit", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: servicemcp.ToolListUnits,
			Arguments: map[string]any{
				"manager": "systemd",
				"limit":   1,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.IsError {
			t.Fatalf("tool error: %#v", res)
		}
		raw, _ := json.Marshal(res.StructuredContent)
		var out servicemcp.ListUnitsResult
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		if len(out.Units) != 1 || out.Units[0].Ref.ID != "nginx.service" {
			t.Fatalf("got %#v", out.Units)
		}
	})

	t.Run("get_service_status", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: servicemcp.ToolGetServiceStatus,
			Arguments: map[string]any{
				"manager": "systemd",
				"id":      "nginx.service",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.IsError {
			t.Fatalf("tool error: %#v", res)
		}
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		var out servicemcp.GetServiceStatusResult
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		if out.Status.Ref.ID != "nginx.service" || out.Status.Active != domain.ActiveStateActive {
			t.Fatalf("status: %#v", out.Status)
		}
		if out.Status.MainPID == nil || *out.Status.MainPID != 4242 {
			t.Fatalf("main pid: %#v", out.Status.MainPID)
		}
	})

	t.Run("list_failed_services", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      servicemcp.ToolListFailedServices,
			Arguments: map[string]any{"manager": "systemd"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.IsError {
			t.Fatalf("tool error: %#v", res)
		}
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		var out servicemcp.ListFailedServicesResult
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		if len(out.Failed) != 2 {
			t.Fatalf("failed: %#v", out.Failed)
		}
		if out.Failed[0].Ref.ID != "nginx.service" || out.Failed[0].Reason != "e2e-failed" {
			t.Fatalf("failed[0]: %#v", out.Failed[0])
		}
		if out.Failed[1].Ref.ID != "z-failed.service" {
			t.Fatalf("failed[1]: %#v", out.Failed[1])
		}
	})

	t.Run("get_recent_service_logs", func(t *testing.T) {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: servicemcp.ToolGetRecentServiceLogs,
			Arguments: map[string]any{
				"manager":   "systemd",
				"service":   "nginx.service",
				"max_lines": 2,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.IsError {
			t.Fatalf("tool error: %#v", res)
		}
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		var out servicemcp.GetRecentServiceLogsResult
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		if out.EffectiveMaxLines != 2 || len(out.Logs) != 2 {
			t.Fatalf("logs out: %#v", out)
		}
		if out.Logs[0].Line != "[e2e] log line 2" || out.Logs[1].Line != "[e2e] log line 3" {
			t.Fatalf("tail mismatch: %#v", out.Logs)
		}
	})
}

// Guard: e2e fixture report stays aligned with what tests assert.
func TestDeterministicReport_shape(t *testing.T) {
	t.Parallel()
	rep := e2etest.DeterministicReport()
	if rep.GOOS != "linux" || len(rep.Findings) != 1 {
		t.Fatalf("unexpected report: %#v", rep)
	}
}
