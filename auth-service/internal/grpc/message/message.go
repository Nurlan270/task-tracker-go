package message

var (
	ErrUserNotFound = "User with provided credentials does not exists."

	ErrEmailTaken = "This email address is already taken."

	ErrRefreshTokenInvalid = "Invalid refresh token."

	ErrInternal = "Something went wrong, please try again later."
)
