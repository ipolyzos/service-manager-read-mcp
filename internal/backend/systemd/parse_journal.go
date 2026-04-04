package systemd

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/ipolyzos/service-manager-read-mcp/internal/domain"
	"github.com/ipolyzos/service-manager-read-mcp/internal/svcerr"
)

// ParseJournalShortISO parses journalctl -o short-iso lines into RecentLogEntry values.
func ParseJournalShortISO(stdout []byte, unit domain.UnitRef) ([]domain.RecentLogEntry, error) {
	if err := ValidateUTF8Lines(stdout, 0); err != nil {
		return nil, err
	}
	var out []domain.RecentLogEntry
	sc := bufio.NewScanner(bytes.NewReader(stdout))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		ent, err := parseJournalLine(line, unit)
		if err != nil {
			return nil, err
		}
		out = append(out, ent)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseJournalLine(line string, unit domain.UnitRef) (domain.RecentLogEntry, error) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return domain.RecentLogEntry{}, svcerr.ErrParse
	}
	tsRaw, _, msg := parts[0], parts[1], parts[2]
	ts, err := parseJournalTimestamp(tsRaw)
	if err != nil {
		return domain.RecentLogEntry{}, err
	}
	u := unit
	return domain.RecentLogEntry{
		Timestamp: &ts,
		Line:      msg,
		Severity:  domain.LogSeverityUnknown,
		Unit:      &u,
	}, nil
}

func parseJournalTimestamp(tsRaw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z07:00",
	}
	var last error
	for _, l := range layouts {
		t, err := time.Parse(l, tsRaw)
		if err == nil {
			return t, nil
		}
		last = err
	}
	return time.Time{}, fmt.Errorf("%w: journal timestamp %q: %v", svcerr.ErrParse, tsRaw, last)
}
