package runit

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

func TestBackend_ListUnits(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_ = os.Mkdir(filepath.Join(dir, "nginx"), 0o755)
	_ = os.Mkdir(filepath.Join(dir, "cron"), 0o755)
	b := newBackendWithFunc(t, func(k string) string {
		if k == "SVDIR" {
			return dir
		}
		return ""
	}, nil)
	units, err := b.ListUnits(context.Background(), domain.ListUnitsInput{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 || units[0].Ref.ID != "cron" || units[0].State != domain.StateUnknown {
		t.Fatalf("%#v", units)
	}
}

func TestBackend_GetServiceStatus_func(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	raw, err := os.ReadFile(filepath.Join("testdata", "sv_status_run.txt"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(k string) string {
		if k == "SVDIR" {
			return dir
		}
		return ""
	}, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if filepath.Clean(s.Dir) != filepath.Clean(dir) {
			t.Errorf("Spec.Dir=%q want %q", s.Dir, dir)
		}
		if len(s.Args) >= 2 && s.Args[0] == "status" && s.Args[1] == "sshd" {
			return runner.Result{Stdout: raw}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	st, err := b.GetServiceStatus(context.Background(), domain.UnitRef{Manager: domain.ManagerRunit, ID: "sshd"})
	if err != nil || st.Active != domain.ActiveStateActive {
		t.Fatalf("err=%v %#v", err, st)
	}
}

func TestBackend_ListFailed_wantUp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_ = os.Mkdir(filepath.Join(dir, "nginx"), 0o755)
	_ = os.Mkdir(filepath.Join(dir, "cron"), 0o755)
	wantUp, err := os.ReadFile(filepath.Join("testdata", "sv_status_want_up.txt"))
	if err != nil {
		t.Fatal(err)
	}
	ok, err := os.ReadFile(filepath.Join("testdata", "sv_status_run.txt"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(k string) string {
		if k == "SVDIR" {
			return dir
		}
		return ""
	}, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if filepath.Clean(s.Dir) != filepath.Clean(dir) {
			t.Errorf("Spec.Dir=%q want %q", s.Dir, dir)
		}
		if len(s.Args) >= 2 && s.Args[0] == "status" {
			switch s.Args[1] {
			case "nginx":
				return runner.Result{Stdout: wantUp}, nil
			case "cron":
				return runner.Result{Stdout: ok}, nil
			}
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	f, err := b.ListFailedServices(context.Background())
	if err != nil || len(f) != 1 || f[0].Ref.ID != "nginx" {
		t.Fatalf("err=%v %#v", err, f)
	}
}

func TestBackend_detect_func(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, nil, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) >= 1 && s.Args[0] == "-v" {
			return runner.Result{Stdout: []byte("runit version 2.1.2\n")}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	m, err := b.DetectServiceManager(context.Background())
	if err != nil || m != domain.ManagerRunit {
		t.Fatalf("m=%v err=%v", m, err)
	}
}

func TestBackend_GetServiceStatus_wrongManager(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, nil, nil)
	_, err := b.GetServiceStatus(context.Background(), domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx"})
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestBackend_GetRecentServiceLogs_unsupported(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, nil, nil)
	_, err := b.GetRecentServiceLogs(context.Background(), domain.UnitRef{Manager: domain.ManagerRunit, ID: "nginx"}, domain.RecentLogsInput{})
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func newBackendWithFunc(t *testing.T, getenv func(string) string, fn runner.Func) *Backend {
	t.Helper()
	reg, err := runner.NewRegistry(Policies("true")...)
	if err != nil {
		t.Skip("registry:", err)
	}
	b := &Backend{Getenv: getenv}
	inner := runner.Runner(runner.Exec{})
	if fn != nil {
		inner = fn
	}
	b.safe = &runner.SafeRunner{
		Inner:    &dirRunner{b: b, inner: inner},
		Registry: reg,
	}
	return b
}
