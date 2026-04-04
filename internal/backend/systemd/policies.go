package systemd

import (
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

// Command IDs for the systemd allowlist (stable for diagnostics).
const (
	cmdDetect      runner.CommandID = "systemd_detect"
	cmdListUnits   runner.CommandID = "systemd_list_units_json"
	cmdShowUnit    runner.CommandID = "systemd_show_unit"
	cmdListFailed  runner.CommandID = "systemd_list_failed_json"
	cmdJournalTail runner.CommandID = "systemd_journal_tail"
)

// Policies returns argv-only recipes for systemctl/journalctl read-only queries.
func Policies(systemctlBin, journalctlBin string) []runner.Policy {
	showProps := []string{
		"--no-pager",
		"--property=Id",
		"--property=ActiveState",
		"--property=SubState",
		"--property=LoadState",
		"--property=UnitFileState",
		"--property=Description",
		"--property=MainPID",
		"--property=FragmentPath",
	}
	return []runner.Policy{
		{
			ID:             cmdDetect,
			Binary:         systemctlBin,
			FixedArgs:      []string{"show", "systemd"},
			SuffixArgs:     []string{"--no-pager", "--property=Id"},
			DefaultTimeout: 5 * time.Second,
		},
		{
			ID:             cmdListUnits,
			Binary:         systemctlBin,
			FixedArgs:      []string{"list-units", "--type=service", "--no-pager", "--no-legend", "--all", "-o", "json"},
			DefaultTimeout: 30 * time.Second,
		},
		{
			ID:             cmdShowUnit,
			Binary:         systemctlBin,
			FixedArgs:      []string{"show"},
			Params:         []runner.ParamRule{UnitNameParam("unit")},
			SuffixArgs:     showProps,
			DefaultTimeout: 15 * time.Second,
		},
		{
			ID:             cmdListFailed,
			Binary:         systemctlBin,
			FixedArgs:      []string{"list-units", "--state=failed", "--type=service", "--no-pager", "--no-legend", "-o", "json"},
			DefaultTimeout: 30 * time.Second,
		},
		{
			ID:        cmdJournalTail,
			Binary:    journalctlBin,
			FixedArgs: []string{"--no-pager", "-o", "short-iso", "-n"},
			Params: []runner.ParamRule{
				JournalLineCountParam("lines", 500),
				UnitNameParam("unit"),
			},
			BetweenParams:  [][]string{{"-u"}},
			DefaultTimeout: 20 * time.Second,
		},
	}
}
