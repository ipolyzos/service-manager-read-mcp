package servicemcp

import (
	"context"
	"errors"
	"os/exec"
	"testing"

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
