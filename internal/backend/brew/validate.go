package brew

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/ipolyzos/service-manager-read-mcp/internal/runner"
)

const maxFormulaNameRunes = 128

// FormulaNameParam validates a Homebrew formula name (no path components).
func FormulaNameParam(name string) runner.ParamRule {
	return runner.ParamRule{
		Name: name,
		Validate: func(value string) error {
			return validateFormulaName(value)
		},
	}
}

func validateFormulaName(s string) error {
	if s == "" {
		return fmt.Errorf("empty formula name")
	}
	runes := []rune(s)
	if len(runes) > maxFormulaNameRunes {
		return fmt.Errorf("formula name too long")
	}
	for i, r := range runes {
		if i == 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("formula name must start with letter or digit")
		}
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
		case r == '.', r == '_', r == '-', r == '+', r == '@':
		default:
			return fmt.Errorf("invalid character in formula name")
		}
	}
	return nil
}

// ValidateUTF8Output rejects invalid UTF-8 in captured output before JSON decode.
func ValidateUTF8Output(b []byte) error {
	if !utf8.Valid(b) {
		return fmt.Errorf("invalid utf-8")
	}
	return nil
}
