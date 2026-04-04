package launchd

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

const maxLabelRunes = 256

var printTargetPattern = regexp.MustCompile(`^(?:gui|user)/[0-9]+/[A-Za-z0-9._-]+$|^system/[A-Za-z0-9._-]+$`)

// LabelParam validates a launchd job label (reverse-DNS style, no argv metacharacters).
func LabelParam(name string) runner.ParamRule {
	return runner.ParamRule{
		Name: name,
		Validate: func(value string) error {
			return validateLabel(value)
		},
	}
}

func validateLabel(s string) error {
	if s == "" {
		return fmt.Errorf("empty label")
	}
	if len([]rune(s)) > maxLabelRunes {
		return fmt.Errorf("label too long")
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '.', r == '_', r == '-':
		default:
			return fmt.Errorf("invalid character in label")
		}
	}
	return nil
}

// PrintTargetParam validates a launchctl print domain path built by the backend.
func PrintTargetParam(name string) runner.ParamRule {
	return runner.ParamRule{
		Name: name,
		Validate: func(value string) error {
			if !printTargetPattern.MatchString(value) {
				return fmt.Errorf("invalid print target")
			}
			if !utf8.ValidString(value) {
				return fmt.Errorf("invalid utf-8")
			}
			return nil
		},
	}
}
