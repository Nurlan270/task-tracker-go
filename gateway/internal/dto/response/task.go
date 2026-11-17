package response

import "github.com/Nurlan270/task-tracker-go/gateway/internal/entity"

type ListTask struct {
	Tasks []*entity.Task `json:"tasks"`
}
