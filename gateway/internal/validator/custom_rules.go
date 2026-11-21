package validator

import (
	"strings"

	"github.com/Nurlan270/task-tracker-go/common/validator"

	"github.com/Nurlan270/task-tracker-go/gateway/internal/entity"

	gpv "github.com/go-playground/validator/v10"
)

func TaskTypeCustomRule() validator.CustomRule {
	return validator.CustomRule{
		Tag: tagTaskType,
		Fn:  validateTaskType,
	}
}

func validateTaskType(fl gpv.FieldLevel) bool {
	value := entity.TaskType(strings.ToLower(fl.Field().String()))

	switch value {
	case entity.TaskTypeUncompleted,
		entity.TaskTypeCompleted,
		entity.TaskTypeDeleted:
		return true
	default:
		return false
	}
}
