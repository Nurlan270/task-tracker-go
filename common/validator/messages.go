package validator

import (
	"fmt"

	gpv "github.com/go-playground/validator/v10"
)

const messageDefault = "_default"

// Built-in validation rules' message's.
var messages = map[string]MessageFunc{
	"required": func(f gpv.FieldError) string {
		return fmt.Sprintf("%s is required.", f.StructField())
	},

	"min": func(f gpv.FieldError) string {
		return fmt.Sprintf("%s must be at least %s characters.", f.StructField(), f.Param())
	},

	"max": func(f gpv.FieldError) string {
		return fmt.Sprintf("%s must be at most %s characters.", f.StructField(), f.Param())
	},

	"email": func(f gpv.FieldError) string {
		return fmt.Sprintf("%s must be a valid email address.", f.StructField())
	},

	//	Default fallback validation message
	messageDefault: func(f gpv.FieldError) string {
		return fmt.Sprintf("%s field failed on %s validation rule.", f.StructField(), f.Tag())
	},
}
