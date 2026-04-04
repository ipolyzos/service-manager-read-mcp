package version

import "testing"

func TestMCPVersion_nonEmpty(t *testing.T) {
	t.Parallel()
	if MCPVersion() == "" {
		t.Fatal("MCPVersion() empty")
	}
}

func TestSummary_nonEmpty(t *testing.T) {
	t.Parallel()
	if Summary() == "" {
		t.Fatal("Summary() empty")
	}
}
