package request

type Register struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type Logout struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Refresh struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
