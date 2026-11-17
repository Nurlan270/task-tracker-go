package request

type CreateTask struct {
	UserID      uint   `json:"user_id"     validate:"required"`
	Title       string `json:"title"       validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"required,max=5000"`
}

type UpdateTask struct {
	TaskID      uint   `json:"task_id"     validate:"required"`
	UserID      uint   `json:"user_id"     validate:"required"`
	Title       string `json:"title"       validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"required,max=5000"`
}

type DeleteTask struct {
	TaskID  uint `json:"task_id"  validate:"required"`
	UserID  uint `json:"user_id"  validate:"required"`
	IsForce bool `json:"is_force"`
}

type ListTask struct {
	UserID uint `json:"user_id" validate:"required"`
}

type MarkTaskAsCompleted struct {
	TaskID uint `json:"task_id" validate:"required"`
	UserID uint `json:"user_id" validate:"required"`
}
