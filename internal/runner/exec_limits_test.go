package runner

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

func TestReadLimited_truncates(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(strings.Repeat("a", 50))
	b, err := readLimited(r, 10)
	if !errors.Is(err, svcerr.ErrOutputLimit) || len(b) != 10 {
		t.Fatalf("b=%q err=%v", b, err)
	}
}

func TestExec_echoStdout(t *testing.T) {
	t.Parallel()
	echo, err := exec.LookPath("echo")
	if err != nil {
		t.Skip("echo not in PATH:", err)
	}
	var ex Exec
	res, err := ex.Run(context.Background(), Spec{Path: echo, Args: []string{"hello"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Stdout), "hello") {
		t.Fatalf("stdout=%q", res.Stdout)
	}
}

func TestExec_yesOutputLimit(t *testing.T) {
	t.Parallel()
	yes, err := exec.LookPath("yes")
	if err != nil {
		t.Skip("yes not in PATH:", err)
	}
	var ex Exec
	_, err = ex.Run(context.Background(), Spec{
		Path:           yes,
		MaxStdoutBytes: 32,
		MaxStderrBytes: 32,
		Timeout:        500 * time.Millisecond,
	})
	if !errors.Is(err, svcerr.ErrOutputLimit) {
		t.Fatalf("expected ErrOutputLimit, got %v", err)
	}
}

func TestExec_sleepTimeout(t *testing.T) {
	t.Parallel()
	sleep, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not in PATH:", err)
	}
	var ex Exec
	_, err = ex.Run(context.Background(), Spec{
		Path:    sleep,
		Args:    []string{"10"},
		Timeout: 25 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return
	}
	var inv *svcerr.InvocationError
	if errors.As(err, &inv) && inv.ExitCode != 0 {
		return
	}
	t.Fatalf("expected timeout or non-zero exit, got %v", err)
}
