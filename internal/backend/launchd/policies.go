package launchd

import (
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

const (
	cmdVersion runner.CommandID = "launchd_version"
	cmdList    runner.CommandID = "launchd_list"
	cmdPrint   runner.CommandID = "launchd_print"
)

// Policies returns argv-only recipes for read-only launchctl use.
func Policies(launchctlBin string) []runner.Policy {
	return []runner.Policy{
		{
			ID:             cmdVersion,
			Binary:         launchctlBin,
			FixedArgs:      []string{"version"},
			DefaultTimeout: 5 * time.Second,
		},
		{
			ID:             cmdList,
			Binary:         launchctlBin,
			FixedArgs:      []string{"list"},
			DefaultTimeout: 30 * time.Second,
		},
		{
			ID:             cmdPrint,
			Binary:         launchctlBin,
			FixedArgs:      []string{"print"},
			Params:         []runner.ParamRule{PrintTargetParam("target")},
			DefaultTimeout: 15 * time.Second,
		},
	}
}
