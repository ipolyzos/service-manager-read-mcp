package runner

import (
	"context"
	"errors"
)

// SafeRunner executes only allowlisted invocations by building a Spec through Registry.Build.
type SafeRunner struct {
	Inner    Runner
	Registry *Registry
}

// RunInvocation resolves the invocation against the allowlist and delegates to Inner.
func (s *SafeRunner) RunInvocation(ctx context.Context, inv Invocation) (Result, error) {
	if s == nil || s.Inner == nil || s.Registry == nil {
		return Result{}, errors.New("runner: safe runner misconfigured")
	}
	spec, err := s.Registry.Build(inv)
	if err != nil {
		return Result{}, err
	}
	return s.Inner.Run(ctx, spec)
}
