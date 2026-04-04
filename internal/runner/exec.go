package runner

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// Exec runs subprocesses with explicit argv and optional deadlines.
type Exec struct{}

// Run implements Runner using os/exec. Non-zero exits return Result populated
// together with *svcerr.InvocationError (errors.Is(..., svcerr.ErrInvocation)).
func (Exec) Run(ctx context.Context, s Spec) (Result, error) {
	if s.Path == "" {
		return Result{}, errors.New("runner: empty path")
	}
	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, s.Path, s.Args...)
	if s.Dir != "" {
		cmd.Dir = s.Dir
	}
	if len(s.Env) > 0 {
		cmd.Env = s.Env
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := Result{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}
	if err == nil {
		return res, nil
	}

	var exit *exec.ExitError
	if errors.As(err, &exit) {
		res.ExitCode = exit.ExitCode()
		return res, &svcerr.InvocationError{
			Path:     s.Path,
			Args:     s.Args,
			ExitCode: res.ExitCode,
			Stdout:   append([]byte(nil), res.Stdout...),
			Stderr:   append([]byte(nil), res.Stderr...),
			Cause:    err,
		}
	}
	return res, err
}

// WithTimeout returns a copy of s with Timeout set if t > 0.
func WithTimeout(s Spec, t time.Duration) Spec {
	if t <= 0 {
		return s
	}
	s.Timeout = t
	return s
}
