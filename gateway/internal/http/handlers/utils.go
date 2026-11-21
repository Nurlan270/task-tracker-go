package handlers

import (
	"errors"
	"io"

	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/gin-gonic/gin"

	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
)

// validateRequest runs the [before] function (if it's not nil) with request passed into it.
// Then it binds request and fills it with data from request. After that, it will validate all data
// within request and return validated request if validation is successful, otherwise it will return nil.
// It will also return nil if [before] returned error.
func validateRequest[T any](c *gin.Context, v *validator.Validate, before func(*T) error) *T {
	var req T

	//	Configure request before binding & validation
	if before != nil {
		if err := before(&req); err != nil {
			httplib.AbortWithErr(c, err)
			return nil
		}
	}

	//	Bind request body
	if err := c.ShouldBind(&req); err != nil && !errors.Is(err, io.EOF) {
		httplib.AbortWithErr(c, errs.ErrBadRequest)
		return nil
	}

	//	Validate request body
	if err := v.Struct(&req); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			httplib.AbortWithErr(c, errors.Join(errs.ErrBadRequest, validationErr))
			return nil
		}

		httplib.AbortWithErr(c, errors.Join(errs.ErrInternalServer, err))

		return nil
	}

	return &req
}
