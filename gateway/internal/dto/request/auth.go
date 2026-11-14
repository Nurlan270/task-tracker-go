package request

type Register struct {
	Name                 string `json:"name" validate:"required"`
	Email                string `json:"email" validate:"required"`
	Password             string `json:"password" validate:"required"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}

type Login struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Logout struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Refresh struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type GetUser struct {
	AccessToken string `json:"access_token" validate:"required"`
}
