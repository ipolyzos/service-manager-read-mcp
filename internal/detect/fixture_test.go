package detect_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/service-manager-read-mcp/internal/detect"
)

type scenarioFile struct {
	GOOS  string            `json:"goos"`
	Paths map[string]string `json:"paths"`
}

func TestDetect_fixtureLinuxMatchesGoldenReport(t *testing.T) {
	t.Parallel()
	root := filepath.Join("testdata", "scenario_linux.json")
	raw, err := os.ReadFile(root)
	if err != nil {
		t.Fatal(err)
	}
	var scen scenarioFile
	if err := json.Unmarshal(raw, &scen); err != nil {
		t.Fatal(err)
	}
	rep := detect.Detect(context.Background(), detect.Environment{
		GOOS:     scen.GOOS,
		LookPath: lookPathMap(scen.Paths),
	})

	goldenPath := filepath.Join("testdata", "report_linux.json")
	wantRaw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	var want detect.Report
	if err := json.Unmarshal(wantRaw, &want); err != nil {
		t.Fatal(err)
	}

	gotRaw, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	wantNorm, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotRaw) != string(wantNorm) {
		t.Fatalf("report mismatch\n--- got\n%s\n--- want\n%s\n", gotRaw, wantNorm)
	}
}
