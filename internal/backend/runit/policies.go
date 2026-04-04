package runit

import (
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

const (
	cmdDetect runner.CommandID = "runit_sv_version"
	cmdStatus runner.CommandID = "runit_sv_status"
)

// Policies returns argv-only recipes for read-only runit `sv` queries.
func Policies(svBin string) []runner.Policy {
	return []runner.Policy{
		{
			ID:             cmdDetect,
			Binary:         svBin,
			FixedArgs:      []string{"-v"},
			DefaultTimeout: 5 * time.Second,
		},
		{
			ID:             cmdStatus,
			Binary:         svBin,
			FixedArgs:      []string{"status"},
			Params:         []runner.ParamRule{ServiceNameParam("service")},
			DefaultTimeout: 15 * time.Second,
		},
	}
}
