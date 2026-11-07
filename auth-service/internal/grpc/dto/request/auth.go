package request

type RegisterUser struct {
	Name     string `validate:"required,min=2,max=50,name"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

type LoginUser struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

type LogoutUser struct {
	RefreshToken string `validate:"required"`
}

type RefreshUser struct {
	RefreshToken string `validate:"required"`
}
