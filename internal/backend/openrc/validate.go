package openrc

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

const maxServiceNameRunes = 128

// ServiceNameParam validates an OpenRC init script / service name (no path components).
func ServiceNameParam(name string) runner.ParamRule {
	return runner.ParamRule{
		Name: name,
		Validate: func(value string) error {
			return validateServiceName(value)
		},
	}
}

// ValidateServiceQueryID applies the same rules as [ServiceNameParam] for read-only queries.
func ValidateServiceQueryID(s string) error {
	return validateServiceName(s)
}

func validateServiceName(s string) error {
	if s == "" {
		return fmt.Errorf("empty service name")
	}
	runes := []rune(s)
	if len(runes) > maxServiceNameRunes {
		return fmt.Errorf("service name too long")
	}
	for i, r := range runes {
		if i == 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("service name must start with letter or digit")
		}
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
		case r == '.', r == '_', r == '-', r == '+', r == '@':
		default:
			return fmt.Errorf("invalid character in service name")
		}
	}
	return nil
}

// ValidateUTF8Output rejects invalid UTF-8 in captured output.
func ValidateUTF8Output(b []byte) error {
	if !utf8.Valid(b) {
		return fmt.Errorf("invalid utf-8")
	}
	return nil
}
