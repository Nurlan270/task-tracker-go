package validator

import (
	"fmt"

	"github.com/Nurlan270/task-tracker-go/common/validator"

	gpv "github.com/go-playground/validator/v10"
)

var Messages = map[string]validator.MessageFunc{
	"name": func(f gpv.FieldError) string {
		return fmt.Sprintf("%s has an invalid format", f.StructField())
	},
}
