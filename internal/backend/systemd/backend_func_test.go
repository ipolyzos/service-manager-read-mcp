package systemd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestBackend_listUnits_viaFunc(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("testdata", "list_units.json"))
	if err != nil {
		t.Fatal(err)
	}
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) > 0 && s.Args[0] == "list-units" && strings.Contains(strings.Join(s.Args, " "), "-o json") {
			return runner.Result{Stdout: raw}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	units, err := b.ListUnits(context.Background(), domain.ListUnitsInput{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 || units[0].Ref.ID != "nginx.service" {
		t.Fatalf("%#v", units)
	}
}

func TestBackend_listUnits_parseError(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		if len(s.Args) > 0 && s.Args[0] == "list-units" {
			return runner.Result{Stdout: []byte(`{`)}, nil
		}
		return runner.Result{}, svcerr.ErrInvocation
	})
	_, err := b.ListUnits(context.Background(), domain.ListUnitsInput{})
	if !errors.Is(err, svcerr.ErrParse) {
		t.Fatalf("expected ErrParse, got %v", err)
	}
}

func TestBackend_GetServiceStatus_wrongManager(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		return runner.Result{}, nil
	})
	_, err := b.GetServiceStatus(context.Background(), domain.UnitRef{Manager: domain.ManagerLaunchd, ID: "x.service"})
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
}

func TestBackend_GetRecentServiceLogs_sinceUnsupported(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		return runner.Result{}, nil
	})
	ts := time.Now()
	_, err := b.GetRecentServiceLogs(context.Background(), domain.UnitRef{Manager: domain.ManagerSystemd, ID: "nginx.service"}, domain.RecentLogsInput{
		MaxLines: 10,
		Since:    &ts,
	})
	if !errors.Is(err, svcerr.ErrUnsupported) {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
}

func TestBackend_invocationErrorPropagates(t *testing.T) {
	t.Parallel()
	b := newBackendWithFunc(t, func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		return runner.Result{}, svcerr.ErrInvocation
	})
	_, err := b.ListFailedServices(context.Background())
	if !errors.Is(err, svcerr.ErrInvocation) {
		t.Fatalf("expected ErrInvocation, got %v", err)
	}
}

func newBackendWithFunc(t *testing.T, fn runner.Func) *Backend {
	t.Helper()
	reg, err := runner.NewRegistry(Policies("true", "true")...)
	if err != nil {
		t.Skip("registry:", err)
	}
	safe := &runner.SafeRunner{Inner: fn, Registry: reg}
	be, err := NewWithSafeRunner(safe)
	if err != nil {
		t.Fatal(err)
	}
	return be
}
