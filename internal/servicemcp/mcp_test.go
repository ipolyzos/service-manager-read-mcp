package servicemcp

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRegisterTools_toolMetadata(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewHandler(Options{
		DetectEnv: detect.Environment{
			GOOS: "linux",
			LookPath: func(string) (string, error) {
				return "", exec.ErrNotFound
			},
		},
	})
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	RegisterTools(srv, h)

	st, ct := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "v0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	want := map[string]bool{ToolListServiceManagers: false, ToolListUnits: false}
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := want[tool.Name]; !ok {
			t.Fatalf("unexpected tool %q", tool.Name)
		}
		want[tool.Name] = true
		if tool.InputSchema == nil {
			t.Fatalf("tool %q missing input schema", tool.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("missing tool %q", name)
		}
	}
}

func TestRegisterTools_callListServiceManagers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewHandler(Options{
		DetectEnv: detect.Environment{
			GOOS: "linux",
			LookPath: func(string) (string, error) {
				return "", exec.ErrNotFound
			},
		},
	})
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	RegisterTools(srv, h)

	st, ct := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "v0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolListServiceManagers,
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
	var out ListServiceManagersResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.Report.GOOS != "linux" || len(out.Report.Findings) == 0 {
		t.Fatalf("unexpected report: %#v", out.Report)
	}
}

func TestRegisterTools_listUnitsValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewHandler(Options{})
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	RegisterTools(srv, h)

	st, ct := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "v0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	_, err = session.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolListUnits,
		Arguments: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected JSON-RPC error: input schema requires manager")
	}
}

func TestRegisterTools_listUnitsUnavailableBackend(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewHandler(Options{
		DetectEnv: detect.Environment{
			GOOS: "linux",
			LookPath: func(string) (string, error) {
				return "", exec.ErrNotFound
			},
		},
	})
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	RegisterTools(srv, h)

	st, ct := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "v0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolListUnits,
		Arguments: map[string]any{"manager": "systemd"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error when systemd not on PATH")
	}
}
