package detect

import (
	"context"
	"os/exec"
	"runtime"
	"sort"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

// Method describes how a finding was produced.
const (
	MethodExecutablePath = "executable_path"
	MethodSkippedOS      = "skipped_os"
)

// Environment controls host context and binary resolution. The zero value uses runtime.GOOS and exec.LookPath.
type Environment struct {
	GOOS     string
	LookPath func(file string) (string, error)
}

func (e Environment) resolved() Environment {
	if e.GOOS == "" {
		e.GOOS = runtime.GOOS
	}
	if e.LookPath == nil {
		e.LookPath = exec.LookPath
	}
	return e
}

// Report is the structured result of a detection pass.
type Report struct {
	GOOS     string    `json:"goos"`
	Findings []Finding `json:"findings"`
}

// Finding describes one supervisor integration on the host.
type Finding struct {
	Manager      domain.ServiceManager      `json:"manager"`
	Applicable   bool                       `json:"applicable"`
	Present      bool                       `json:"present"`
	ResolvedPath string                     `json:"resolved_path,omitempty"`
	Method       string                     `json:"method"`
	Capabilities domain.BackendCapabilities `json:"capabilities"`
}

type probe struct {
	manager domain.ServiceManager
	binary  string
	gooses  []string // empty means attempt on every GOOS
}

// defaultProbes enumerates CLIs used for coarse availability. Discovery uses PATH resolution only (no subprocess).
var defaultProbes = []probe{
	{manager: domain.ManagerSystemd, binary: "systemctl", gooses: []string{"linux"}},
	{manager: domain.ManagerLaunchd, binary: "launchctl", gooses: []string{"darwin"}},
	{manager: domain.ManagerOpenRC, binary: "rc-service", gooses: []string{"linux", "freebsd"}},
	{manager: domain.ManagerRunit, binary: "sv", gooses: []string{"linux", "freebsd", "darwin", "openbsd", "netbsd"}},
	{manager: domain.ManagerBrew, binary: "brew", gooses: []string{"darwin", "linux"}},
}

// Detect evaluates known supervisors for the given environment.
// It does not execute supervisor CLIs; it resolves binaries via LookPath when the GOOS is applicable.
func Detect(ctx context.Context, env Environment) Report {
	if err := ctx.Err(); err != nil {
		return Report{}
	}
	env = env.resolved()
	findings := make([]Finding, 0, len(defaultProbes))
	for _, p := range defaultProbes {
		findings = append(findings, evaluateProbe(env, p))
	}
	sort.Slice(findings, func(i, j int) bool {
		return string(findings[i].Manager) < string(findings[j].Manager)
	})
	return Report{
		GOOS:     env.GOOS,
		Findings: findings,
	}
}

func evaluateProbe(env Environment, p probe) Finding {
	if !goosAllowed(env.GOOS, p.gooses) {
		return Finding{
			Manager:      p.manager,
			Applicable:   false,
			Present:      false,
			Method:       MethodSkippedOS,
			Capabilities: domain.BackendCapabilities{Manager: p.manager},
		}
	}
	path, err := env.LookPath(p.binary)
	present := err == nil && path != ""
	f := Finding{
		Manager:      p.manager,
		Applicable:   true,
		Present:      present,
		Method:       MethodExecutablePath,
		Capabilities: capabilityTemplate(p.manager, present),
	}
	if present {
		f.ResolvedPath = path
	}
	return f
}

func goosAllowed(goos string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, g := range allowed {
		if g == goos {
			return true
		}
	}
	return false
}
