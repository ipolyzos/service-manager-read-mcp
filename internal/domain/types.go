package domain

import "time"

// ServiceManager identifies a supervisor integration namespace.
type ServiceManager string

const (
	ManagerSystemd ServiceManager = "systemd"
	ManagerLaunchd ServiceManager = "launchd"
	ManagerOpenRC  ServiceManager = "openrc"
	ManagerRunit   ServiceManager = "runit"
	ManagerBrew    ServiceManager = "brew_services"
	ManagerUnknown ServiceManager = "unknown"
)

// CoarseState is a portable summary for list views.
type CoarseState string

const (
	StateRunning CoarseState = "running"
	StateStopped CoarseState = "stopped"
	StateFailed  CoarseState = "failed"
	StateUnknown CoarseState = "unknown"
)

// ActiveState refines runtime state when the backend exposes it (e.g. systemd ActiveState).
type ActiveState string

const (
	ActiveStateUnknown      ActiveState = "unknown"
	ActiveStateActive       ActiveState = "active"
	ActiveStateInactive     ActiveState = "inactive"
	ActiveStateFailed       ActiveState = "failed"
	ActiveStateActivating   ActiveState = "activating"
	ActiveStateDeactivating ActiveState = "deactivating"
)

// LoadState describes whether a unit definition is loaded (when applicable).
type LoadState string

const (
	LoadStateUnknown    LoadState = "unknown"
	LoadStateLoaded     LoadState = "loaded"
	LoadStateNotFound   LoadState = "not_found"
	LoadStateBadSetting LoadState = "bad_setting"
)

// LogSeverity is a coarse hint for recent log lines; backends map native levels here.
type LogSeverity string

const (
	LogSeverityUnknown LogSeverity = "unknown"
	LogSeverityDebug   LogSeverity = "debug"
	LogSeverityInfo    LogSeverity = "info"
	LogSeverityWarning LogSeverity = "warning"
	LogSeverityError   LogSeverity = "error"
)

// UnitRef identifies a unit within a backend namespace.
type UnitRef struct {
	Manager     ServiceManager `json:"manager"`
	ID          string         `json:"id"`
	DisplayName string         `json:"display_name,omitempty"`
}

// Unit is a lightweight row for list operations.
type Unit struct {
	Ref   UnitRef     `json:"ref"`
	State CoarseState `json:"state"`
}

// ServiceStatus is the normalized detail view for a single unit.
type ServiceStatus struct {
	Ref         UnitRef           `json:"ref"`
	Active      ActiveState       `json:"active"`
	SubState    string            `json:"sub_state,omitempty"`
	Load        LoadState         `json:"load,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
	MainPID     *int64            `json:"main_pid,omitempty"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FailedServiceEntry captures a supervisor-reported failure record.
type FailedServiceEntry struct {
	Ref    UnitRef    `json:"ref"`
	Reason string     `json:"reason,omitempty"`
	Since  *time.Time `json:"since,omitempty"`
}

// RecentLogEntry is a single bounded log line with optional structured hints.
type RecentLogEntry struct {
	Timestamp *time.Time  `json:"timestamp,omitempty"`
	Line      string      `json:"line"`
	Severity  LogSeverity `json:"severity,omitempty"`
	Unit      *UnitRef    `json:"unit,omitempty"`
}

// BackendCapabilities describes what a backend implementation offers on the current host.
type BackendCapabilities struct {
	Manager                  ServiceManager `json:"manager"`
	SupportsDetect           bool           `json:"supports_detect"`
	SupportsListUnits        bool           `json:"supports_list_units"`
	SupportsGetServiceStatus bool           `json:"supports_get_service_status"`
	SupportsListFailed       bool           `json:"supports_list_failed_services"`
	SupportsRecentLogs       bool           `json:"supports_get_recent_logs"`
}

// ListUnitsInput narrows list operations; zero value means backend defaults.
type ListUnitsInput struct {
	Limit int `json:"limit,omitempty"`
}

// RecentLogsInput bounds log collection.
type RecentLogsInput struct {
	MaxLines int        `json:"max_lines,omitempty"`
	Since    *time.Time `json:"since,omitempty"`
}
