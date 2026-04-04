package runner

import (
	"context"
	"time"
)

// Spec describes an argv-only subprocess (no shell).
type Spec struct {
	Path    string
	Args    []string
	Dir     string
	Env     []string
	Timeout time.Duration
}

// Result captures bounded process output and exit status.
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Runner executes read-only diagnostic commands on behalf of backends.
type Runner interface {
	Run(ctx context.Context, s Spec) (Result, error)
}
