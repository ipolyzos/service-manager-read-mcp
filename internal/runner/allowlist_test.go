package runner_test

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestRegistry_buildsExpectedArgv(t *testing.T) {
	t.Parallel()
	truePath, err := exec.LookPath("true")
	if err != nil {
		t.Skip("true not in PATH:", err)
	}
	reg, err := runner.NewRegistry(runner.Policy{
		ID:        "demo_true",
		Binary:    "true",
		FixedArgs: []string{},
		Params:    nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := reg.Build(runner.Invocation{ID: "demo_true"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Path != truePath || len(spec.Args) != 0 {
		t.Fatalf("spec: path=%q args=%q", spec.Path, spec.Args)
	}
	if spec.MaxStdoutBytes != runner.DefaultMaxStreamBytes {
		t.Fatalf("stdout cap: %d", spec.MaxStdoutBytes)
	}
}

func TestRegistry_argumentConstruction(t *testing.T) {
	t.Parallel()
	echoPath, err := exec.LookPath("echo")
	if err != nil {
		t.Skip("echo not in PATH:", err)
	}
	reg, err := runner.NewRegistry(runner.Policy{
		ID:        "echo_message",
		Binary:    "echo",
		FixedArgs: []string{"-n"},
		Params: []runner.ParamRule{
			runner.NonEmptyParam("message"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := reg.Build(runner.Invocation{
		ID:     "echo_message",
		Params: []string{"hello"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Path != echoPath {
		t.Fatalf("path: %q", spec.Path)
	}
	want := []string{"-n", "hello"}
	if len(spec.Args) != len(want) {
		t.Fatalf("args=%v want=%v", spec.Args, want)
	}
	for i := range want {
		if spec.Args[i] != want[i] {
			t.Fatalf("args=%v want=%v", spec.Args, want)
		}
	}
}

func TestRegistry_blockedUnknownID(t *testing.T) {
	t.Parallel()
	reg, err := runner.NewRegistry(runner.Policy{
		ID:     "only_one",
		Binary: "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = reg.Build(runner.Invocation{ID: "nope"})
	if !errors.Is(err, svcerr.ErrNotAllowlisted) {
		t.Fatalf("expected ErrNotAllowlisted, got %v", err)
	}
}

func TestRegistry_blockedArityAndValidation(t *testing.T) {
	t.Parallel()
	reg, err := runner.NewRegistry(runner.Policy{
		ID:     "need_param",
		Binary: "true",
		Params: []runner.ParamRule{runner.NonEmptyParam("unit")},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = reg.Build(runner.Invocation{ID: "need_param"})
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for arity, got %v", err)
	}
	_, err = reg.Build(runner.Invocation{ID: "need_param", Params: []string{"  "}})
	if !errors.Is(err, svcerr.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for validation, got %v", err)
	}
}
