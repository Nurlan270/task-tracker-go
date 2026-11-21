package validator

import (
	"github.com/Nurlan270/task-tracker-go/common/validator"

	"github.com/Nurlan270/task-tracker-go/task-service/internal/models"

	gpv "github.com/go-playground/validator/v10"
)

func TaskTypeCustomRule() validator.CustomRule {
	return validator.CustomRule{
		Tag: tagTaskType,
		Fn:  validateTaskType,
	}
}

func validateTaskType(fl gpv.FieldLevel) bool {
	value := models.TaskType(fl.Field().Int())

	switch value {
	case models.TaskTypeUncompleted,
		models.TaskTypeCompleted,
		models.TaskTypeDeleted:
		return true
	default:
		return false
	}
}
