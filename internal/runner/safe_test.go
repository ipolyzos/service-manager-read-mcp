package runner_test

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestSafeRunner_allowedExecution(t *testing.T) {
	t.Parallel()
	truePath, err := exec.LookPath("true")
	if err != nil {
		t.Skip("true not in PATH:", err)
	}
	reg, err := runner.NewRegistry(runner.Policy{
		ID:     "true_noop",
		Binary: "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	var seen runner.Spec
	inner := runner.Func(func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		seen = s
		return runner.Result{}, nil
	})
	safe := &runner.SafeRunner{Inner: inner, Registry: reg}
	_, err = safe.RunInvocation(context.Background(), runner.Invocation{ID: "true_noop"})
	if err != nil {
		t.Fatal(err)
	}
	if seen.Path != truePath || len(seen.Args) != 0 {
		t.Fatalf("unexpected spec: %#v", seen)
	}
}

func TestSafeRunner_propagatesAllowlistErrors(t *testing.T) {
	t.Parallel()
	reg, err := runner.NewRegistry(runner.Policy{
		ID:     "true_noop",
		Binary: "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	safe := &runner.SafeRunner{
		Inner: runner.Func(func(ctx context.Context, s runner.Spec) (runner.Result, error) {
			return runner.Result{}, nil
		}),
		Registry: reg,
	}
	_, err = safe.RunInvocation(context.Background(), runner.Invocation{ID: "missing"})
	if !errors.Is(err, svcerr.ErrNotAllowlisted) {
		t.Fatalf("expected ErrNotAllowlisted, got %v", err)
	}
}

func TestSafeRunner_respectsContextTimeout(t *testing.T) {
	t.Parallel()
	reg, err := runner.NewRegistry(runner.Policy{
		ID:     "block",
		Binary: "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	inner := runner.Func(func(ctx context.Context, s runner.Spec) (runner.Result, error) {
		<-ctx.Done()
		return runner.Result{}, ctx.Err()
	})
	safe := &runner.SafeRunner{Inner: inner, Registry: reg}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err = safe.RunInvocation(ctx, runner.Invocation{ID: "block"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
}
