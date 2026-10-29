package request

import (
	"unicode"

	valid "github.com/go-ozzo/ozzo-validation/v4"
)

func validateName(value any) error {
	name, ok := value.(string)
	if !ok {
		return valid.NewError(
			"validation_name",
			"must be a string",
		)
	}

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
			return valid.NewError(
				"validation_name",
				"contains unsupported characters",
			)
		}
	}

	return nil
}
