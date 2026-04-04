package openrc

import (
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

const (
	cmdDetect  runner.CommandID = "openrc_rc_service_version"
	cmdList    runner.CommandID = "openrc_rc_service_list"
	cmdStatus  runner.CommandID = "openrc_rc_service_status"
	cmdCrashed runner.CommandID = "openrc_rc_status_crashed"
)

// Policies returns argv-only recipes for read-only OpenRC queries.
func Policies(rcServiceBin, rcStatusBin string) []runner.Policy {
	return []runner.Policy{
		{
			ID:             cmdDetect,
			Binary:         rcServiceBin,
			FixedArgs:      []string{"-V"},
			DefaultTimeout: 5 * time.Second,
		},
		{
			ID:             cmdList,
			Binary:         rcServiceBin,
			FixedArgs:      []string{"-l"},
			DefaultTimeout: 30 * time.Second,
		},
		{
			ID:             cmdStatus,
			Binary:         rcServiceBin,
			Params:         []runner.ParamRule{ServiceNameParam("service")},
			SuffixArgs:     []string{"status"},
			DefaultTimeout: 15 * time.Second,
		},
		{
			ID:             cmdCrashed,
			Binary:         rcStatusBin,
			FixedArgs:      []string{"--crashed"},
			DefaultTimeout: 15 * time.Second,
		},
	}
}
