package request

type RegisterUser struct {
	Name                 string `json:"name" validate:"required,min=2,max=50,name"`
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}

type LoginUser struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LogoutUser struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshUser struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type GetUser struct {
	AccessToken string `json:"access_token" validate:"required"`
}
