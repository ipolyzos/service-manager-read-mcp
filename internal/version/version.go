// Package version records release metadata for binaries and MCP initialize.
package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Version is the release identifier (ideally semver; a leading "v" is stripped for MCP).
// Set at link time, for example:
//
//	-ldflags "-X github.com/ipolyzos/service-manager-read-mcp/internal/version.Version=1.2.3"
var Version string

// Commit is the source revision (full or short SHA).
var Commit string

// BuildDate is informational only (UTC ISO-8601 recommended).
var BuildDate string

// MCPVersion returns a semver string suitable for MCP `implementation.version`.
func MCPVersion() string {
	v := strings.TrimSpace(Version)
	if v != "" {
		return strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V")
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "0.0.0-dev"
}

// Summary returns a single line for CLI -version output.
func Summary() string {
	c := strings.TrimSpace(Commit)
	if c == "" {
		c = vcsRevision()
	}
	if c == "" {
		c = "unknown"
	} else if len(c) > 12 {
		c = c[:12]
	}
	d := strings.TrimSpace(BuildDate)
	if d == "" {
		d = "unknown"
	}
	return fmt.Sprintf("service-manager-read-mcp %s (commit %s, built %s)", MCPVersion(), c, d)
}

func vcsRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			return s.Value
		}
	}
	return ""
}
