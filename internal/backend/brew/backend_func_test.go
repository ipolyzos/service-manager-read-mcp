package brew

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

func TestBackend_ListUnits_func(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "brew_services_list.json"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) >= 3 && s.Args[0] == "services" && s.Args[1] == "list" && s.Args[2] == "--json" {
			return runner.Result{Stdout: raw}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	units, err := b.ListUnits(context.Background(), domain.ListUnitsInput{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 2 || units[0].Ref.ID != "nginx" {
		t.Fatalf("%#v", units)
	}
}

func TestBackend_ListFailed_func(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "brew_services_list.json"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) >= 3 && s.Args[0] == "services" && s.Args[1] == "list" {
			return runner.Result{Stdout: raw}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	f, err := b.ListFailedServices(context.Background())
	if err != nil || len(f) != 3 {
		t.Fatalf("err=%v n=%d %#v", err, len(f), f)
	}
}

func TestBackend_GetServiceStatus_func(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "brew_services_info_nginx.json"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) >= 4 && s.Args[0] == "services" && s.Args[1] == "info" && s.Args[2] == "nginx" && s.Args[3] == "--json" {
			return runner.Result{Stdout: raw}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	st, err := b.GetServiceStatus(context.Background(), domain.UnitRef{Manager: domain.ManagerBrew, ID: "nginx"})
	if err != nil || st.Active != domain.ActiveStateActive {
		t.Fatalf("err=%v %#v", err, st)
	}
}

func TestBackend_GetServiceStatus_formulaMissing(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		return runner.Result{}, &svcerr.InvocationError{
			ExitCode: 1,
			Stderr:   []byte("Error: No available formula with the name \"nope\".\n"),
		}
	})
	_, err := b.GetServiceStatus(context.Background(), domain.UnitRef{Manager: domain.ManagerBrew, ID: "nope"})
	if !errors.Is(err, svcerr.ErrUnitNotFound) {
		t.Fatalf("got %v", err)
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
	_, err := b.GetRecentServiceLogs(context.Background(), domain.UnitRef{Manager: domain.ManagerBrew, ID: "nginx"}, domain.RecentLogsInput{})
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestBackend_detect_func(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) >= 1 && s.Args[0] == "--version" {
			return runner.Result{Stdout: []byte("Homebrew 4.2.0\n")}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	m, err := b.DetectServiceManager(context.Background())
	if err != nil || m != domain.ManagerBrew {
		t.Fatalf("m=%v err=%v", m, err)
	}
}

func newBackendWithFunc(t *testing.T, fn runner.Func) *Backend {
	t.Helper()
	reg, err := runner.NewRegistry(Policies("true")...)
	if err != nil {
		t.Skip("registry:", err)
	}
	inner := runner.Runner(runner.Exec{})
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
