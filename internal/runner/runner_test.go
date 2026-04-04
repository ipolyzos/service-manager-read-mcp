package runner_test

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestFuncRunner(t *testing.T) {
	t.Parallel()
	var seen runner.Spec
	r := runner.Func(func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		seen = s
		return runner.Result{Stdout: []byte("ok")}, nil
	})
	res, err := r.Run(context.Background(), runner.Spec{Path: "/bin/echo", Args: []string{"hi"}})
	if err != nil {
		t.Fatal(err)
	}
	if string(res.Stdout) != "ok" {
		t.Fatalf("stdout: %q", res.Stdout)
	}
	if seen.Path != "/bin/echo" || len(seen.Args) != 1 || seen.Args[0] != "hi" {
		t.Fatalf("spec not forwarded: %#v", seen)
	}
}

func TestExecRun_true(t *testing.T) {
	t.Parallel()
	path, err := exec.LookPath("true")
	if err != nil {
		t.Skip("true not found:", err)
	}
	var ex runner.Exec
	res, err := ex.Run(context.Background(), runner.Spec{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit: %d", res.ExitCode)
	}
}

func TestExecRun_false(t *testing.T) {
	t.Parallel()
	path, err := exec.LookPath("false")
	if err != nil {
		t.Skip("false not found:", err)
	}
	var ex runner.Exec
	res, err := ex.Run(context.Background(), runner.Spec{Path: path})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, svcerr.ErrInvocation) {
		t.Fatalf("expected ErrInvocation, got %v", err)
	}
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit")
	}
}
