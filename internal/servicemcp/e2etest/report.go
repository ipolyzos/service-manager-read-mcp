package e2etest

import (
	"context"

	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

// DeterministicReport is a fixed [detect.Report] used by the e2e MCP binary so CI and
// local runs do not depend on host PATH or OS supervisor layout.
func DeterministicReport() detect.Report {
	caps := domain.BackendCapabilities{
		Manager:                  domain.ManagerSystemd,
		SupportsDetect:           true,
		SupportsListUnits:        true,
		SupportsGetServiceStatus: true,
		SupportsListFailed:       true,
		SupportsRecentLogs:       true,
	}
	return detect.Report{
		GOOS: "linux",
		Findings: []detect.Finding{
			{
				Manager:      domain.ManagerSystemd,
				Applicable:   true,
				Present:      true,
				ResolvedPath: "/e2e-stub/bin/systemctl",
				Method:       detect.MethodExecutablePath,
				Capabilities: caps,
			},
		},
	}
}

// DetectHook returns [DeterministicReport]; suitable for [servicemcp.Options.DetectHook].
func DetectHook(context.Context, detect.Environment) detect.Report {
	return DeterministicReport()
}
