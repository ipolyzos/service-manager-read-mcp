package detect_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

func lookPathMap(paths map[string]string) func(string) (string, error) {
	return func(name string) (string, error) {
		p, ok := paths[name]
		if !ok || p == "" {
			return "", exec.ErrNotFound
		}
		return p, nil
	}
}

func TestDetect_linux(t *testing.T) {
	t.Parallel()
	paths := map[string]string{
		"systemctl":  "/usr/bin/systemctl",
		"rc-service": "/sbin/rc-service",
		"sv":         "/bin/sv",
		"brew":       "",
	}
	rep := detect.Detect(context.Background(), detect.Environment{
		GOOS:     "linux",
		LookPath: lookPathMap(paths),
	})
	if rep.GOOS != "linux" {
		t.Fatalf("goos: %q", rep.GOOS)
	}
	by := findingsByManager(rep)
	if f := by[domain.ManagerSystemd]; !f.Applicable || !f.Present || f.ResolvedPath != paths["systemctl"] {
		t.Fatalf("systemd: %#v", f)
	}
	if f := by[domain.ManagerLaunchd]; f.Applicable || f.Present {
		t.Fatalf("launchd should be skipped on linux: %#v", f)
	}
	if f := by[domain.ManagerOpenRC]; !f.Applicable || !f.Present {
		t.Fatalf("openrc: %#v", f)
	}
	if f := by[domain.ManagerRunit]; !f.Applicable || !f.Present {
		t.Fatalf("runit: %#v", f)
	}
	if f := by[domain.ManagerBrew]; !f.Applicable || f.Present {
		t.Fatalf("brew: want applicable without binary, got %#v", f)
	}
}

func TestDetect_darwinBrew(t *testing.T) {
	t.Parallel()
	paths := map[string]string{
		"launchctl": "/bin/launchctl",
		"brew":      "/opt/homebrew/bin/brew",
	}
	rep := detect.Detect(context.Background(), detect.Environment{
		GOOS:     "darwin",
		LookPath: lookPathMap(paths),
	})
	by := findingsByManager(rep)
	if f := by[domain.ManagerLaunchd]; !f.Applicable || !f.Present {
		t.Fatalf("launchd: %#v", f)
	}
	if f := by[domain.ManagerBrew]; !f.Applicable || !f.Present {
		t.Fatalf("brew: %#v", f)
	}
	if f := by[domain.ManagerSystemd]; f.Applicable {
		t.Fatalf("systemd should be skipped on darwin: %#v", f)
	}
}

func TestDetect_contextCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rep := detect.Detect(ctx, detect.Environment{
		GOOS: "linux",
		LookPath: func(string) (string, error) {
			t.Fatal("lookpath should not run when ctx canceled")
			return "", nil
		},
	})
	if len(rep.Findings) != 0 || rep.GOOS != "" {
		t.Fatalf("expected empty report, got %#v", rep)
	}
}

func findingsByManager(rep detect.Report) map[domain.ServiceManager]detect.Finding {
	out := make(map[domain.ServiceManager]detect.Finding, len(rep.Findings))
	for _, f := range rep.Findings {
		out[f.Manager] = f
	}
	return out
}
