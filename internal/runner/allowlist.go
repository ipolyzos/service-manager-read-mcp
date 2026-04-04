package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// DefaultMaxStreamBytes caps each of stdout and stderr when a policy leaves limits at zero.
const DefaultMaxStreamBytes = 1 << 20

// CommandID names a single allowlisted argv recipe. Callers must not treat arbitrary strings as IDs.
type CommandID string

// Invocation requests execution of one allowlisted command with validated parameter slots only.
// Flags and subcommands are never taken from free-form user input; they come from policy FixedArgs.
type Invocation struct {
	ID      CommandID
	Params  []string
	Timeout time.Duration
}

// ParamRule validates one positional parameter appended after FixedArgs in declaration order.
type ParamRule struct {
	Name     string
	Validate func(value string) error
}

// Policy describes one executable path, fixed argv tail, and validated parameter slots.
type Policy struct {
	ID        CommandID
	Binary    string
	FixedArgs []string
	Params    []ParamRule
	// SuffixArgs are appended after Params (e.g. trailing --property= flags for systemctl show).
	SuffixArgs []string
	// BetweenParams holds argv fragments emitted after each param except the last (e.g. "-u" between -n COUNT and UNIT).
	BetweenParams  [][]string
	DefaultTimeout time.Duration
	// MaxStdoutBytes and MaxStderrBytes bound each stream; zero applies DefaultMaxStreamBytes at build time.
	MaxStdoutBytes int
	MaxStderrBytes int

	// Path is populated by NewRegistry when resolving Binary and must not be set by callers.
	Path string
}

// Registry is an immutable allowlist of policies indexed by CommandID.
type Registry struct {
	byID map[CommandID]Policy
}

// NewRegistry validates policies, resolves binaries once, and returns a registry.
func NewRegistry(policies ...Policy) (*Registry, error) {
	if len(policies) == 0 {
		return nil, errors.New("runner: empty registry")
	}
	byID := make(map[CommandID]Policy, len(policies))
	for _, p := range policies {
		if p.ID == "" {
			return nil, errors.New("runner: policy id is empty")
		}
		if _, dup := byID[p.ID]; dup {
			return nil, fmt.Errorf("runner: duplicate command id %q", p.ID)
		}
		if strings.TrimSpace(p.Binary) == "" {
			return nil, fmt.Errorf("runner: policy %q: binary is empty", p.ID)
		}
		path, err := resolveExecutable(p.Binary)
		if err != nil {
			return nil, fmt.Errorf("runner: policy %q: resolve binary: %w", p.ID, err)
		}
		np := p
		np.Path = path
		byID[p.ID] = np
	}
	return &Registry{byID: byID}, nil
}

// Build converts an invocation into a Spec. Unknown IDs, arity mismatches, and validation failures error out.
func (r *Registry) Build(inv Invocation) (Spec, error) {
	p, ok := r.byID[inv.ID]
	if !ok {
		return Spec{}, svcerr.ErrNotAllowlisted
	}
	if len(inv.Params) != len(p.Params) {
		return Spec{}, fmt.Errorf("%w: want %d params, got %d", svcerr.ErrInvalidArgument, len(p.Params), len(inv.Params))
	}
	for i, val := range inv.Params {
		if err := p.Params[i].Validate(val); err != nil {
			name := p.Params[i].Name
			if name == "" {
				name = fmt.Sprintf("param_%d", i)
			}
			return Spec{}, fmt.Errorf("%w: %s: %v", svcerr.ErrInvalidArgument, name, err)
		}
	}
	if len(p.BetweenParams) != 0 && len(p.BetweenParams) != len(inv.Params)-1 {
		return Spec{}, fmt.Errorf("runner: policy %q: between-params must have length len(params)-1", p.ID)
	}

	args := append([]string{}, p.FixedArgs...)
	for i, par := range inv.Params {
		args = append(args, par)
		if i < len(p.BetweenParams) {
			args = append(args, p.BetweenParams[i]...)
		}
	}
	if len(p.SuffixArgs) > 0 {
		args = append(args, p.SuffixArgs...)
	}

	timeout := p.DefaultTimeout
	if inv.Timeout > 0 {
		timeout = inv.Timeout
	}

	maxOut := p.MaxStdoutBytes
	if maxOut <= 0 {
		maxOut = DefaultMaxStreamBytes
	}
	maxErr := p.MaxStderrBytes
	if maxErr <= 0 {
		maxErr = DefaultMaxStreamBytes
	}

	return Spec{
		Path:           p.Path,
		Args:           args,
		Timeout:        timeout,
		MaxStdoutBytes: maxOut,
		MaxStderrBytes: maxErr,
	}, nil
}

// NonEmptyParam rejects blank parameters after trimming spaces.
func NonEmptyParam(name string) ParamRule {
	return ParamRule{
		Name: name,
		Validate: func(value string) error {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("must not be empty")
			}
			return nil
		},
	}
}

// MaxLenParam rejects values longer than n runes (not bytes).
func MaxLenParam(name string, n int) ParamRule {
	return ParamRule{
		Name: name,
		Validate: func(value string) error {
			if len([]rune(value)) > n {
				return fmt.Errorf("longer than %d runes", n)
			}
			return nil
		},
	}
}

func resolveExecutable(binary string) (string, error) {
	if filepath.IsAbs(binary) {
		fi, err := os.Stat(binary)
		if err != nil {
			return "", err
		}
		if fi.IsDir() {
			return "", fmt.Errorf("path %q is a directory", binary)
		}
		return binary, nil
	}
	return exec.LookPath(binary)
}

// AllowedPolicyIDs returns a sorted snapshot of registered IDs for diagnostics.
func (r *Registry) AllowedPolicyIDs() []CommandID {
	out := make([]CommandID, 0, len(r.byID))
	for id := range r.byID {
		out = append(out, id)
	}
	// stable order for tests
	sortCommandIDs(out)
	return out
}

func sortCommandIDs(ids []CommandID) {
	// small n; insertion sort by string
	for i := 1; i < len(ids); i++ {
		for j := i; j > 0 && string(ids[j-1]) > string(ids[j]); j-- {
			ids[j-1], ids[j] = ids[j], ids[j-1]
		}
	}
}
