package validator

import (
	"fmt"

	gpv "github.com/go-playground/validator/v10"
)

const messageDefault = "_default"

// Built-in validation rules' message's.
var messages = map[string]MessageFunc{
	"required": func(f gpv.FieldError, fname string) string {
		return fmt.Sprintf("%s is required.", fname)
	},

	"min": func(f gpv.FieldError, fname string) string {
		return fmt.Sprintf("%s must be at least %s characters long.", fname, f.Param())
	},

	"max": func(f gpv.FieldError, fname string) string {
		return fmt.Sprintf("%s must be at most %s characters long.", fname, f.Param())
	},

	"email": func(f gpv.FieldError, fname string) string {
		return fmt.Sprintf("%s must be a valid email address.", fname)
	},

	"eqfield": func(f gpv.FieldError, fname string) string {
		return fmt.Sprintf("%s must be equal to %s.", fname, f.Param())
	},

	//	Default fallback validation message
	messageDefault: func(f gpv.FieldError, fname string) string {
		return fmt.Sprintf("%s field failed on %s validation rule.", fname, f.Tag())
	},
}
