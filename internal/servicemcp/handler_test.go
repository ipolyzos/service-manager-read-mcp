package servicemcp

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/backend"
	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestParseServiceManager(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in   string
		want domain.ServiceManager
	}{
		{"systemd", domain.ManagerSystemd},
		{"brew_services", domain.ManagerBrew},
		{"brew", domain.ManagerBrew},
		{"  RUNit ", domain.ManagerRunit},
	} {
		got, err := ParseServiceManager(tc.in)
		if err != nil || got != tc.want {
			t.Fatalf("%q: got %q err=%v", tc.in, got, err)
		}
	}
	_, err := ParseServiceManager("nomanager")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("got %v", err)
	}
}

func reportLinuxSystemd(caps domain.BackendCapabilities) detect.Report {
	caps.Manager = domain.ManagerSystemd
	return detect.Report{
		GOOS: "linux",
		Findings: []detect.Finding{
			{
				Manager:      domain.ManagerSystemd,
				Applicable:   true,
				Present:      true,
				Method:       detect.MethodExecutablePath,
				ResolvedPath: "/usr/bin/systemctl",
				Capabilities: caps,
			},
		},
	}
}

type recordingBackend struct {
	caps      domain.BackendCapabilities
	listErr   error
	failed    []domain.FailedServiceEntry
	failedErr error
	status    domain.ServiceStatus
	statusErr error
	logs      []domain.RecentLogEntry
	logsErr   error
	lastRef   domain.UnitRef
	lastLogs  domain.RecentLogsInput
}

func (b *recordingBackend) Capabilities() domain.BackendCapabilities { return b.caps }

func (b *recordingBackend) DetectServiceManager(ctx context.Context) (domain.ServiceManager, error) {
	_ = ctx
	return b.caps.Manager, nil
}

func (b *recordingBackend) ListUnits(ctx context.Context, in domain.ListUnitsInput) ([]domain.Unit, error) {
	_ = ctx
	_ = in
	return nil, b.listErr
}

func (b *recordingBackend) GetServiceStatus(ctx context.Context, ref domain.UnitRef) (domain.ServiceStatus, error) {
	_ = ctx
	_ = ref
	if b.statusErr != nil {
		return domain.ServiceStatus{}, b.statusErr
	}
	return b.status, nil
}

func (b *recordingBackend) ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error) {
	_ = ctx
	if b.failedErr != nil {
		return nil, b.failedErr
	}
	return b.failed, nil
}

func (b *recordingBackend) GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error) {
	_ = ctx
	b.lastRef = ref
	b.lastLogs = in
	if b.logsErr != nil {
		return nil, b.logsErr
	}
	out := make([]domain.RecentLogEntry, len(b.logs))
	copy(out, b.logs)
	return out, nil
}

func TestSortFailedServiceEntries(t *testing.T) {
	t.Parallel()
	in := []domain.FailedServiceEntry{
		{Ref: domain.UnitRef{Manager: domain.ManagerSystemd, ID: "b.service"}},
		{Ref: domain.UnitRef{Manager: domain.ManagerSystemd, ID: "a.service"}},
	}
	got := sortFailedServiceEntries(in)
	if len(got) != 2 || got[0].Ref.ID != "a.service" || got[1].Ref.ID != "b.service" {
		t.Fatalf("got %#v", got)
	}
	if in[0].Ref.ID != "b.service" {
		t.Fatal("sort must not mutate the source slice")
	}
}

func TestHandler_ListUnits_limitOutOfRange(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{})
	_, err := h.ListUnits(context.Background(), domain.ManagerSystemd, MaxListUnitsLimit+1)
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_ListUnits_unknownManager(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{})
	_, err := h.ListUnits(context.Background(), domain.ServiceManager("notreal"), 0)
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_ListUnits_notOnPath(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectEnv: detect.Environment{
			GOOS: "linux",
			LookPath: func(string) (string, error) {
				return "", exec.ErrNotFound
			},
		},
	})
	_, err := h.ListUnits(context.Background(), domain.ManagerSystemd, 0)
	if !errors.Is(err, svcerr.ErrBackendUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_ListUnits_capabilityUnsupported(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				Manager:                  domain.ManagerSystemd,
				SupportsDetect:           true,
				SupportsListUnits:        false,
				SupportsGetServiceStatus: true,
				SupportsListFailed:       true,
			})
		},
	})
	_, err := h.ListUnits(context.Background(), domain.ManagerSystemd, 0)
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_ListServiceManagers_contextCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := NewHandler(Options{})
	_, err := h.ListServiceManagers(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_ListFailedServices_unknownManager(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:  true,
				SupportsListFailed: true,
			})
		},
		NewBackend: func(domain.ServiceManager) (backend.ServiceManagerBackend, error) {
			return &recordingBackend{}, nil
		},
	})
	_, err := h.ListFailedServices(context.Background(), domain.ServiceManager("nope"))
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_ListFailedServices_notOnPath(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectEnv: detect.Environment{
			GOOS: "linux",
			LookPath: func(string) (string, error) {
				return "", exec.ErrNotFound
			},
		},
	})
	_, err := h.ListFailedServices(context.Background(), domain.ManagerSystemd)
	if !errors.Is(err, svcerr.ErrBackendUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_ListFailedServices_capabilityUnsupported(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:        true,
				SupportsListFailed:       false,
				SupportsGetServiceStatus: true,
			})
		},
	})
	_, err := h.ListFailedServices(context.Background(), domain.ManagerSystemd)
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_ListFailedServices_propagatesBackendError(t *testing.T) {
	t.Parallel()
	want := errors.New("backend boom")
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
				failedErr: want,
			}, nil
		},
	})
	_, err := h.ListFailedServices(context.Background(), domain.ManagerSystemd)
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_ListFailedServices_successSorted(t *testing.T) {
	t.Parallel()
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
					{Ref: domain.UnitRef{Manager: domain.ManagerSystemd, ID: "z.service"}, Reason: "z"},
					{Ref: domain.UnitRef{Manager: domain.ManagerSystemd, ID: "a.service"}, Reason: "a"},
				},
			}, nil
		},
	})
	got, err := h.ListFailedServices(context.Background(), domain.ManagerSystemd)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Ref.ID != "a.service" || got[1].Ref.ID != "z.service" {
		t.Fatalf("got %#v", got)
	}
}

func TestHandler_GetRecentServiceLogs_recentWindowUnsupported(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:  true,
				SupportsRecentLogs: true,
			})
		},
	})
	_, err := h.GetRecentServiceLogs(context.Background(), domain.ManagerSystemd, "nginx.service", 10, "1h")
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_GetRecentServiceLogs_capabilityUnsupported(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:  true,
				SupportsRecentLogs: false,
			})
		},
	})
	_, err := h.GetRecentServiceLogs(context.Background(), domain.ManagerSystemd, "nginx.service", 10, "")
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_GetRecentServiceLogs_successAndPassesMaxLines(t *testing.T) {
	t.Parallel()
	rb := &recordingBackend{
		caps: domain.BackendCapabilities{
			Manager:            domain.ManagerSystemd,
			SupportsRecentLogs: true,
		},
		logs: []domain.RecentLogEntry{
			{Line: "a"},
			{Line: "b"},
		},
	}
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:  true,
				SupportsRecentLogs: true,
			})
		},
		NewBackend: func(domain.ServiceManager) (backend.ServiceManagerBackend, error) {
			return rb, nil
		},
	})
	out, err := h.GetRecentServiceLogs(context.Background(), domain.ManagerSystemd, "nginx.service", 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Logs) != 2 || out.EffectiveMaxLines != 50 || out.Truncated {
		t.Fatalf("got %#v", out)
	}
	if rb.lastLogs.MaxLines != 50 || rb.lastRef.ID != "nginx.service" {
		t.Fatalf("backend in: ref=%#v in=%#v", rb.lastRef, rb.lastLogs)
	}
}

func TestHandler_GetRecentServiceLogs_truncatesExcessLines(t *testing.T) {
	t.Parallel()
	var long []domain.RecentLogEntry
	for i := 0; i < 5; i++ {
		long = append(long, domain.RecentLogEntry{Line: string(rune('a' + i))})
	}
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
				logs: long,
			}, nil
		},
	})
	out, err := h.GetRecentServiceLogs(context.Background(), domain.ManagerSystemd, "nginx.service", 3, "")
	if err != nil {
		t.Fatal(err)
	}
	if !out.Truncated || len(out.Logs) != 3 || out.Logs[2].Line != "e" {
		t.Fatalf("got %#v", out)
	}
}
