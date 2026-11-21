package validator

import (
	"fmt"

	"github.com/Nurlan270/task-tracker-go/common/validator"

	gpv "github.com/go-playground/validator/v10"
)

const tagTaskType = "taskType"

var Messages = map[string]validator.MessageFunc{
	tagTaskType: func(f gpv.FieldError, fname string) string {
		return fmt.Sprintf("%s should be one of the following: uncompleted, completed, deleted", fname)
	},
}
