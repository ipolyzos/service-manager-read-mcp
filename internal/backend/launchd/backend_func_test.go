package launchd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestBackend_listUnits_func(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "list_sample.txt"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) > 0 && s.Args[0] == "list" {
			return runner.Result{Stdout: raw}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	units, err := b.ListUnits(context.Background(), domain.ListUnitsInput{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 2 || units[0].Ref.ID != "com.apple.coreservicesd" {
		t.Fatalf("%#v", units)
	}
}

func TestBackend_GetServiceStatus_func(t *testing.T) {
	t.Parallel()
	printOut, err := os.ReadFile(filepath.Join("testdata", "print_running.txt"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) > 0 && s.Args[0] == "print" && strings.Contains(strings.Join(s.Args, " "), "gui/") {
			return runner.Result{Stdout: printOut}, nil
		}
		if len(s.Args) > 0 && s.Args[0] == "print" {
			return runner.Result{Stdout: []byte("Could not find service\n")}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	st, err := b.GetServiceStatus(context.Background(), domain.UnitRef{Manager: domain.ManagerLaunchd, ID: "com.example.app"})
	if err != nil {
		t.Fatal(err)
	}
	if st.Active != domain.ActiveStateActive {
		t.Fatalf("%#v", st)
	}
}

func TestBackend_GetServiceStatus_wrongManager(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		return runner.Result{}, nil
	})
	_, err := b.GetServiceStatus(context.Background(), domain.UnitRef{Manager: domain.ManagerSystemd, ID: "x.service"})
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestBackend_GetRecentServiceLogs_unsupported(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, nil)
	_, err := b.GetRecentServiceLogs(context.Background(), domain.UnitRef{Manager: domain.ManagerLaunchd, ID: "com.foo"}, domain.RecentLogsInput{MaxLines: 10})
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestBackend_Detect_nonDarwin(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "darwin" {
		t.Skip("darwin uses real version path")
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		t.Fatal("should not invoke launchctl off darwin")
		return runner.Result{}, nil
	})
	_, err := b.DetectServiceManager(context.Background())
	if !errors.Is(err, svcerr.ErrBackendUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestCapabilities_logsDisabled(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, nil)
	c := b.Capabilities()
	if c.SupportsRecentLogs {
		t.Fatal("expected logs unsupported")
	}
	if !c.SupportsListFailed {
		t.Fatal("failed heuristic should be advertised")
	}
}

func newBackendWithFunc(t *testing.T, fn runner.Func) *Backend {
	t.Helper()
	reg, err := runner.NewRegistry(Policies("true")...)
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
