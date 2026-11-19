package request

type CreateTask struct {
	UserID      uint64 `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTask struct {
	TaskID      uint64 `json:"task_id"`
	UserID      uint64 `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type DeleteTask struct {
	TaskID  uint64 `json:"task_id"`
	UserID  uint64 `json:"user_id"`
	IsForce bool   `json:"is_force"`
}

type ListTask struct {
	UserID uint64 `json:"user_id"`
}

type ToggleCompletion struct {
	TaskID uint64 `json:"task_id"`
	UserID uint64 `json:"user_id"`
}
