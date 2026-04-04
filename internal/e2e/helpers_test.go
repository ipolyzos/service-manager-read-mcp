package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found from test working directory")
		}
		dir = parent
	}
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// buildE2EMCPServer compiles cmd/service-manager-read-mcp-e2e into a temp directory.
func buildE2EMCPServer(t *testing.T) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "service-manager-read-mcp-e2e"+exeSuffix())
	cmd := exec.Command("go", "build", "-o", out, "./cmd/service-manager-read-mcp-e2e")
	cmd.Dir = repoRoot(t)
	cmd.Env = os.Environ()
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./cmd/service-manager-read-mcp-e2e: %v\n%s", err, b)
	}
	return out
}
