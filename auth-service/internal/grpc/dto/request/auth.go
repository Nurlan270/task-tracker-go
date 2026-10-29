package request

import (
	"strings"

	"github.com/go-ozzo/ozzo-validation/v4/is"

	valid "github.com/go-ozzo/ozzo-validation/v4"
)

type RegisterUser struct {
	Name     string
	Email    string
	Password string
}

// Normalize normalizes input values and gets them ready for validation
//
// It should be called before Validate().
func (r *RegisterUser) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.TrimSpace(r.Email)
}

// Validate validates request body.
func (r *RegisterUser) Validate() error {
	return handleValidationErr(valid.ValidateStruct(r,
		//	Name - length: 2-50 chars long; satisfies custom validation rule
		valid.Field(&r.Name, valid.Required, valid.Length(2, 50), valid.By(validateName)),
		//	Email - valid email address
		//	NOTE: I've disabled actual email check by MX records as it'll slow down validation
		//	It can be enabled in future if needed by simply replacing is.EmailFormat to is.Email
		valid.Field(&r.Email, valid.Required, is.EmailFormat),
		//	Password - length: min. 8 chars long
		valid.Field(&r.Password, valid.Required, valid.RuneLength(8, 0)),
	))
}

type LoginUser struct {
	Email    string
	Password string
}

// Normalize normalizes input values and gets them ready for validation
//
// It should be called before Validate().
func (r *LoginUser) Normalize() {
	r.Email = strings.TrimSpace(r.Email)
}

// Validate validates request body.
func (r *LoginUser) Validate() error {
	return handleValidationErr(valid.ValidateStruct(r,
		//	Email - valid email address
		//	NOTE: I've disabled actual email check by MX records as it'll slow down validation
		//	It can be enabled in future if needed by simply replacing is.EmailFormat to is.Email
		valid.Field(&r.Email, valid.Required, is.EmailFormat),
		//	Password - length: min. 8 chars long
		valid.Field(&r.Password, valid.Required, valid.RuneLength(8, 0)),
	))
}

type LogoutUser struct {
	RefreshToken string
}

// Normalize normalizes input values and gets them ready for validation
//
// It should be called before Validate().
func (r *LogoutUser) Normalize() {
	r.RefreshToken = strings.TrimSpace(r.RefreshToken)
}

// Validate validates request body.
func (r *LogoutUser) Validate() error {
	return handleValidationErr(valid.ValidateStruct(r,
		valid.Field(&r.RefreshToken, valid.Required),
	))
}
