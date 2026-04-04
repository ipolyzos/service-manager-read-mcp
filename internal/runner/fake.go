package runner

import "context"

// Func adapts a function to Runner for tests.
type Func func(ctx context.Context, s Spec) (Result, error)

// Run implements Runner.
func (f Func) Run(ctx context.Context, s Spec) (Result, error) {
	return f(ctx, s)
}
