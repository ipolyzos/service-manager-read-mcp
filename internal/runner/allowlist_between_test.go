package runner_test

import (
	"os/exec"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

func TestRegistry_betweenParams(t *testing.T) {
	t.Parallel()
	truePath, err := exec.LookPath("true")
	if err != nil {
		t.Skip("true not in PATH:", err)
	}
	reg, err := runner.NewRegistry(runner.Policy{
		ID:        "two_slot",
		Binary:    "true",
		FixedArgs: []string{"a"},
		Params: []runner.ParamRule{
			runner.NonEmptyParam("first"),
			runner.NonEmptyParam("second"),
		},
		BetweenParams: [][]string{{"between"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := reg.Build(runner.Invocation{ID: "two_slot", Params: []string{"one", "two"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a", "one", "between", "two"}
	if len(spec.Args) != len(want) {
		t.Fatalf("args=%v want=%v", spec.Args, want)
	}
	for i := range want {
		if spec.Args[i] != want[i] {
			t.Fatalf("args=%v want=%v", spec.Args, want)
		}
	}
	if spec.Path != truePath {
		t.Fatalf("path %q", spec.Path)
	}
}
