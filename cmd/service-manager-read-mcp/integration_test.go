package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCLI_stdioSmoke builds the binary and runs it with an immediate EOF on stdin.
// The process should terminate (not hang) once stdio closes; the MCP handshake may
// not complete, so exit errors are ignored. This guards the stdio entrypoint wiring.
func TestCLI_stdioSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo := findRepoRoot(t)
	bin := filepath.Join(t.TempDir(), "service-manager-read-mcp")
	build := exec.Command("go", "build", "-o", bin, "./cmd/service-manager-read-mcp")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	run := exec.CommandContext(ctx, bin)
	run.Stdin = strings.NewReader("")
	run.Stdout = io.Discard
	run.Stderr = io.Discard
	_ = run.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("server did not exit after stdin closed")
	}
}

func findRepoRoot(t *testing.T) string {
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
			t.Fatal("go.mod not found from ", dir)
		}
		dir = parent
	}
}
