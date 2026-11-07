package response

type Success struct {
	Message string `json:"message"`
}

func NewSuccess(msg string) *Success {
	return &Success{Message: msg}
}
