package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
)

func TestDomainJSONRoundTrip(t *testing.T) {
	t.Parallel()
	enabled := true
	pid := int64(42)
	since := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
	ts := time.Date(2026, 4, 4, 12, 1, 0, 0, time.UTC)

	ref := domain.UnitRef{
		Manager:     domain.ManagerSystemd,
		ID:          "nginx.service",
		DisplayName: "nginx",
	}
	payload := struct {
		Unit   domain.Unit                `json:"unit"`
		Status domain.ServiceStatus       `json:"status"`
		Failed domain.FailedServiceEntry  `json:"failed"`
		Log    domain.RecentLogEntry      `json:"log"`
		Caps   domain.BackendCapabilities `json:"caps"`
		ListIn domain.ListUnitsInput      `json:"list_in"`
		LogsIn domain.RecentLogsInput     `json:"logs_in"`
	}{
		Unit: domain.Unit{Ref: ref, State: domain.StateRunning},
		Status: domain.ServiceStatus{
			Ref:         ref,
			Active:      domain.ActiveStateActive,
			SubState:    "running",
			Load:        domain.LoadStateLoaded,
			Enabled:     &enabled,
			MainPID:     &pid,
			Description: "high performance web server",
			Metadata:    map[string]string{"fragment": "nginx.service"},
		},
		Failed: domain.FailedServiceEntry{Ref: ref, Reason: "exit-code", Since: &since},
		Log: domain.RecentLogEntry{
			Timestamp: &ts,
			Line:      "nginx: ready",
			Severity:  domain.LogSeverityInfo,
			Unit:      &ref,
		},
		Caps: domain.BackendCapabilities{
			Manager:                  domain.ManagerSystemd,
			SupportsDetect:           true,
			SupportsListUnits:        true,
			SupportsGetServiceStatus: true,
			SupportsListFailed:       true,
			SupportsRecentLogs:       true,
		},
		ListIn: domain.ListUnitsInput{Limit: 100},
		LogsIn: domain.RecentLogsInput{MaxLines: 50, Since: &since},
	}

	enc, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got struct {
		Unit   domain.Unit                `json:"unit"`
		Status domain.ServiceStatus       `json:"status"`
		Failed domain.FailedServiceEntry  `json:"failed"`
		Log    domain.RecentLogEntry      `json:"log"`
		Caps   domain.BackendCapabilities `json:"caps"`
		ListIn domain.ListUnitsInput      `json:"list_in"`
		LogsIn domain.RecentLogsInput     `json:"logs_in"`
	}
	if err := json.Unmarshal(enc, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Unit.Ref.ID != payload.Unit.Ref.ID || got.Unit.State != payload.Unit.State {
		t.Fatalf("unit mismatch: %#v vs %#v", got.Unit, payload.Unit)
	}
	if got.Status.Active != payload.Status.Active || got.Status.SubState != payload.Status.SubState {
		t.Fatalf("status mismatch: %#v vs %#v", got.Status, payload.Status)
	}
	if got.Failed.Reason != payload.Failed.Reason {
		t.Fatalf("failed mismatch: %#v vs %#v", got.Failed, payload.Failed)
	}
	if got.Log.Line != payload.Log.Line || got.Log.Severity != payload.Log.Severity {
		t.Fatalf("log mismatch: %#v vs %#v", got.Log, payload.Log)
	}
	if got.Caps.Manager != payload.Caps.Manager || got.Caps.SupportsListUnits != payload.Caps.SupportsListUnits {
		t.Fatalf("caps mismatch: %#v vs %#v", got.Caps, payload.Caps)
	}
	if got.ListIn.Limit != payload.ListIn.Limit || got.LogsIn.MaxLines != payload.LogsIn.MaxLines {
		t.Fatalf("inputs mismatch: list=%#v logs=%#v", got.ListIn, got.LogsIn)
	}
}
