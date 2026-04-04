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
	status    domain.ServiceStatus
	statusErr error
	listErr   error
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
	return b.status, b.statusErr
}

func (b *recordingBackend) ListFailedServices(ctx context.Context) ([]domain.FailedServiceEntry, error) {
	_ = ctx
	return nil, svcerr.ErrUnsupported
}

func (b *recordingBackend) GetRecentServiceLogs(ctx context.Context, ref domain.UnitRef, in domain.RecentLogsInput) ([]domain.RecentLogEntry, error) {
	_ = ctx
	_ = ref
	_ = in
	return nil, svcerr.ErrUnsupported
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

func TestHandler_GetServiceStatus_unknownManager(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:        true,
				SupportsGetServiceStatus: true,
			})
		},
		NewBackend: func(domain.ServiceManager) (backend.ServiceManagerBackend, error) {
			return &recordingBackend{}, nil
		},
	})
	_, err := h.GetServiceStatus(context.Background(), domain.ServiceManager("nope"), "x.service")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_GetServiceStatus_notOnPath(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectEnv: detect.Environment{
			GOOS: "linux",
			LookPath: func(string) (string, error) {
				return "", exec.ErrNotFound
			},
		},
	})
	_, err := h.GetServiceStatus(context.Background(), domain.ManagerSystemd, "nginx.service")
	if !errors.Is(err, svcerr.ErrBackendUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_GetServiceStatus_capabilityUnsupported(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:        true,
				SupportsGetServiceStatus: false,
			})
		},
	})
	_, err := h.GetServiceStatus(context.Background(), domain.ManagerSystemd, "nginx.service")
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_GetServiceStatus_invalidID(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:        true,
				SupportsGetServiceStatus: true,
			})
		},
	})
	_, err := h.GetServiceStatus(context.Background(), domain.ManagerSystemd, "bad name!")
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_GetServiceStatus_unitNotFound(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:        true,
				SupportsGetServiceStatus: true,
			})
		},
		NewBackend: func(domain.ServiceManager) (backend.ServiceManagerBackend, error) {
			return &recordingBackend{
				caps: domain.BackendCapabilities{
					Manager:                  domain.ManagerSystemd,
					SupportsGetServiceStatus: true,
				},
				statusErr: svcerr.ErrUnitNotFound,
			}, nil
		},
	})
	_, err := h.GetServiceStatus(context.Background(), domain.ManagerSystemd, "missing.service")
	if !errors.Is(err, svcerr.ErrUnitNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_GetServiceStatus_backendUnsupported(t *testing.T) {
	t.Parallel()
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:        true,
				SupportsGetServiceStatus: true,
			})
		},
		NewBackend: func(domain.ServiceManager) (backend.ServiceManagerBackend, error) {
			return &recordingBackend{
				caps: domain.BackendCapabilities{
					Manager:                  domain.ManagerSystemd,
					SupportsGetServiceStatus: true,
				},
				statusErr: svcerr.ErrUnsupported,
			}, nil
		},
	})
	_, err := h.GetServiceStatus(context.Background(), domain.ManagerSystemd, "nginx.service")
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestHandler_GetServiceStatus_success(t *testing.T) {
	t.Parallel()
	want := domain.ServiceStatus{
		Ref:    domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx.service"},
		Active: domain.ActiveStateActive,
	}
	h := NewHandler(Options{
		DetectHook: func(context.Context, detect.Environment) detect.Report {
			return reportLinuxSystemd(domain.BackendCapabilities{
				SupportsListUnits:        true,
				SupportsGetServiceStatus: true,
			})
		},
		NewBackend: func(domain.ServiceManager) (backend.ServiceManagerBackend, error) {
			return &recordingBackend{
				caps: domain.BackendCapabilities{
					Manager:                  domain.ManagerSystemd,
					SupportsGetServiceStatus: true,
				},
				status: want,
			}, nil
		},
	})
	got, err := h.GetServiceStatus(context.Background(), domain.ManagerSystemd, "nginx.service")
	if err != nil {
		t.Fatal(err)
	}
	if got.Ref.ID != want.Ref.ID || got.Active != want.Active {
		t.Fatalf("got %#v", got)
	}
}
