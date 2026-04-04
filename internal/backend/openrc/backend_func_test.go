package openrc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestBackend_listUnits_func(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "rc_service_list.txt"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) >= 1 && s.Args[0] == "-l" {
			return runner.Result{Stdout: raw}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	units, err := b.ListUnits(context.Background(), domain.ListUnitsInput{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 2 || units[0].Ref.ID != "acpid" {
		t.Fatalf("%#v", units)
	}
}

func TestBackend_listFailed_func(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "rc_status_crashed.txt"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) >= 1 && s.Args[0] == "--crashed" {
			return runner.Result{Stdout: raw}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	f, err := b.ListFailedServices(context.Background())
	if err != nil || len(f) != 2 {
		t.Fatalf("err=%v n=%d", err, len(f))
	}
}

func TestBackend_GetServiceStatus_func(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "rc_service_status_started.txt"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) >= 2 && s.Args[len(s.Args)-1] == "status" {
			return runner.Result{Stdout: raw}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	st, err := b.GetServiceStatus(context.Background(), domain.UnitRef{Manager: domain.ManagerOpenRC, ID: "nginx"})
	if err != nil || st.Active != domain.ActiveStateActive {
		t.Fatalf("err=%v %#v", err, st)
	}
}

func TestBackend_GetServiceStatus_wrongManager(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, nil)
	_, err := b.GetServiceStatus(context.Background(), domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx"})
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestBackend_GetRecentServiceLogs_unsupported(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, nil)
	_, err := b.GetRecentServiceLogs(context.Background(), domain.UnitRef{Manager: domain.ManagerOpenRC, ID: "nginx"}, domain.RecentLogsInput{})
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestBackend_detect_func(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) >= 1 && s.Args[0] == "-V" {
			return runner.Result{Stdout: []byte("OpenRC 0.52")}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	m, err := b.DetectServiceManager(context.Background())
	if err != nil || m != domain.ManagerOpenRC {
		t.Fatalf("m=%v err=%v", m, err)
	}
}

func newBackendWithFunc(t *testing.T, fn runner.Func) *Backend {
	t.Helper()
	reg, err := runner.NewRegistry(Policies("true", "true")...)
	if err != nil {
		t.Skip("registry:", err)
	}
	var inner runner.Runner = runner.Exec{}
	if fn != nil {
		inner = fn
	}
	safe := &runner.SafeRunner{Inner: inner, Registry: reg}
	be, err := NewWithSafeRunner(safe)
	if err != nil {
		t.Fatal(err)
	}
	return be
}
