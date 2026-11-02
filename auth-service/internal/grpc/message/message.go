package message

var (
	ErrUserNotFound      = "User with provided credentials was not found."
	ErrUserAlreadyExists = "User with provided credentials already exists."

	ErrRefreshTokenInvalid = "Invalid refresh token."

	ErrInternal = "Something went wrong, please try again later."
)
