package brew

import (
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

const (
	cmdVersion runner.CommandID = "brew_version"
	cmdList    runner.CommandID = "brew_services_list_json"
	cmdInfo    runner.CommandID = "brew_services_info_json"
)

// Policies returns argv-only recipes for read-only Homebrew services queries.
func Policies(brewBin string) []runner.Policy {
	return []runner.Policy{
		{
			ID:             cmdVersion,
			Binary:         brewBin,
			FixedArgs:      []string{"--version"},
			DefaultTimeout: 5 * time.Second,
		},
		{
			ID:             cmdList,
			Binary:         brewBin,
			FixedArgs:      []string{"services", "list", "--json"},
			DefaultTimeout: 60 * time.Second,
		},
		{
			ID:             cmdInfo,
			Binary:         brewBin,
			FixedArgs:      []string{"services", "info"},
			Params:         []runner.ParamRule{FormulaNameParam("formula")},
			SuffixArgs:     []string{"--json"},
			DefaultTimeout: 30 * time.Second,
		},
	}
}
