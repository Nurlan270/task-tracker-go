package validator

import (
	"unicode"

	"github.com/Nurlan270/task-tracker-go/common/validator"

	gpv "github.com/go-playground/validator/v10"
)

func NameCustomRule() validator.CustomRule {
	return validator.CustomRule{
		TagName: "name",
		Fn:      validateName,
	}
}

func validateName(fl gpv.FieldLevel) bool {
	name := fl.Field().String()

	for _, r := range name {
		switch {
		case unicode.IsLetter(r):
			// Latin, Cyrillic, Arabic, etc.
			continue

		case unicode.IsMark(r):
			// Combining marks used in accented characters.
			continue

		case unicode.IsSpace(r):
			continue

		case r == '-' || r == '\'' || r == '\u2019' || r == '.':
			// Hyphens, straight/curly apostrophes, periods.
			continue

		default:
			return false
		}
	}

	return true
}
