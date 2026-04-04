package servicemcp

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend"
	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
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
	t.Cleanup(func() { _ = session.Close() })

	want := map[string]bool{
		ToolListServiceManagers:  false,
		ToolListUnits:            false,
		ToolListFailedServices:   false,
		ToolGetServiceStatus:     false,
		ToolGetRecentServiceLogs: false,
	}
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
	t.Cleanup(func() { _ = session.Close() })

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
	t.Cleanup(func() { _ = session.Close() })

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
	t.Cleanup(func() { _ = session.Close() })

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

func TestRegisterTools_listFailedServicesValidation(t *testing.T) {
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
	t.Cleanup(func() { _ = session.Close() })

	_, err = session.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolListFailedServices,
		Arguments: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected JSON-RPC error: input schema requires manager")
	}
}

func TestRegisterTools_listFailedServicesSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:  true,
				SupportsListFailed: true,
			})
		},
		NewBackend: func(domain.ServiceManager) (backend.ServiceManagerBackend, error) {
			return &recordingBackend{
				caps: domain.BackendCapabilities{
					Manager:            domain.ManagerSystemd,
					SupportsListFailed: true,
				},
				failed: []domain.FailedServiceEntry{
					{Ref: domain.UnitRef{Manager: domain.ManagerSystemd, ID: "bad.service"}, Reason: "failed"},
				},
			}, nil
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
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolListFailedServices,
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
	var out ListFailedServicesResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Failed) != 1 || out.Failed[0].Ref.ID != "bad.service" {
		t.Fatalf("unexpected: %#v", out.Failed)
	}
}

func TestRegisterTools_listFailedServicesUnsupported(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:  true,
				SupportsListFailed: false,
			})
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
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolListFailedServices,
		Arguments: map[string]any{"manager": "systemd"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error when list_failed unsupported")
	}
}

func TestRegisterTools_listFailedServicesUnavailable(t *testing.T) {
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
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolListFailedServices,
		Arguments: map[string]any{"manager": "systemd"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error when manager not on PATH")
	}
}

func TestRegisterTools_listFailedServicesNotFoundManagerString(t *testing.T) {
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
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      ToolListFailedServices,
		Arguments: map[string]any{"manager": "not_a_manager"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for invalid manager")
	}
}

func TestRegisterTools_getRecentServiceLogsUnsupportedCapability(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:  true,
				SupportsRecentLogs: false,
			})
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
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: ToolGetRecentServiceLogs,
		Arguments: map[string]any{
			"manager":   "systemd",
			"service":   "nginx.service",
			"max_lines": 10,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error when logs unsupported")
	}
}

func TestRegisterTools_getRecentServiceLogsSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:  true,
				SupportsRecentLogs: true,
			})
		},
		NewBackend: func(domain.ServiceManager) (backend.ServiceManagerBackend, error) {
			return &recordingBackend{
				caps: domain.BackendCapabilities{
					Manager:            domain.ManagerSystemd,
					SupportsRecentLogs: true,
				},
				logs: []domain.RecentLogEntry{{Line: "hello"}},
			}, nil
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
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: ToolGetRecentServiceLogs,
		Arguments: map[string]any{
			"manager":   "systemd",
			"service":   "nginx.service",
			"max_lines": 25,
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
	var out GetRecentServiceLogsResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Logs) != 1 || out.Logs[0].Line != "hello" || out.EffectiveMaxLines != 25 {
		t.Fatalf("unexpected out: %#v", out)
	}
}

func TestRegisterTools_getRecentServiceLogsRecentWindowRejected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:  true,
				SupportsRecentLogs: true,
			})
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
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: ToolGetRecentServiceLogs,
		Arguments: map[string]any{
			"manager":       "systemd",
			"service":       "nginx.service",
			"max_lines":     10,
			"recent_window": "30m",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected unsupported error for recent_window")
	}
}
