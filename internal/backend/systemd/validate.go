package systemd

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

const maxUnitNameRunes = 256

// UnitNameParam validates a systemd unit name for safe argv use (conservative charset).
func UnitNameParam(name string) runner.ParamRule {
	return runner.ParamRule{
		Name: name,
		Validate: func(value string) error {
			return validateUnitName(value)
		},
	}
}

// ValidateServiceQueryID applies the same rules as [UnitNameParam] for read-only queries.
func ValidateServiceQueryID(s string) error {
	return validateUnitName(s)
}

func validateUnitName(s string) error {
	if s == "" {
		return fmt.Errorf("empty unit name")
	}
	if len([]rune(s)) > maxUnitNameRunes {
		return fmt.Errorf("unit name too long")
	}
	for _, r := range s {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
		case r == '_', r == '-', r == ':', r == '.', r == '@', r == '\\':
		default:
			return fmt.Errorf("invalid character in unit name")
		}
	}
	return nil
}

// JournalLineCountParam validates journalctl -n argument (bounded retrieval).
func JournalLineCountParam(name string, maxLines int) runner.ParamRule {
	if maxLines < 1 {
		maxLines = 500
	}
	return runner.ParamRule{
		Name: name,
		Validate: func(value string) error {
			n, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("not an integer")
			}
			if n < 1 || n > maxLines {
				return fmt.Errorf("lines must be between 1 and %d", maxLines)
			}
			return nil
		},
	}
}

// ValidateUTF8Lines returns an error if b is not valid UTF-8 or exceeds maxBytes.
func ValidateUTF8Lines(b []byte, maxBytes int) error {
	if maxBytes > 0 && len(b) > maxBytes {
		return fmt.Errorf("output exceeds %d bytes", maxBytes)
	}
	if !utf8.Valid(b) {
		return fmt.Errorf("invalid utf-8")
	}
	return nil
}
