package validator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	gpv "github.com/go-playground/validator/v10"
)

type MessageFunc func(f gpv.FieldError) string

type Validate struct {
	*gpv.Validate
	messageMap map[string]MessageFunc
}

func New(opts ...gpv.Option) *Validate {
	v := gpv.New(opts...)

	validator := &Validate{
		Validate:   v,
		messageMap: messages,
	}

	validator.registerTagName()

	return validator
}

type ErrValidation struct {
	errors map[string]string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation errors: %s", e.errors)
}

func (e *ErrValidation) GetErrors() map[string]string {
	return e.errors
}

func (v *Validate) Struct(s any) error {
	err := v.StructCtx(context.Background(), s)
	if err == nil {
		return nil
	}

	var vErrs gpv.ValidationErrors

	if !errors.As(err, &vErrs) {
		return err
	}

	errorMap := make(map[string]string, len(vErrs))
	for _, vErr := range vErrs {
		errorMap[vErr.Field()] = v.getMessage(vErr)
	}

	return &ErrValidation{
		errors: errorMap,
	}
}

func (v *Validate) registerTagName() {
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	})
}

type CustomRule struct {
	//	Tag associated with validation function
	Tag string

	//	Validation function
	Fn gpv.Func
}

// RegisterCustomRules registers passed custom validation rules.
//
// It will panic if there's an error while registering rule
//
//	NOTES:
//	- if the tag name already exists, the previous validation function will be replaced.
//	- this method is not thread-safe it is intended that these all be registered prior to any validation
func (v *Validate) RegisterCustomRules(rules ...CustomRule) {
	for _, rule := range rules {
		if err := v.RegisterValidation(rule.Tag, rule.Fn); err != nil {
			panic(fmt.Sprintf("failed to register custom validation rule %q: %s", rule.Tag, err))
		}
	}
}

// RegisterCustomMessages registers passed custom validation messages.
//
//	NOTES:
//	- if the tag name already exists, the previous validation function will be replaced.
//	- this method is not thread-safe it is intended that these all be registered prior to any validation
func (v *Validate) RegisterCustomMessages(messages map[string]MessageFunc) {
	for tag, fn := range messages {
		v.messageMap[tag] = fn
	}
}

func (v *Validate) getMessage(f gpv.FieldError) string {
	for tag, msg := range v.messageMap {
		if tag == f.Tag() {
			return msg(f)
		}
	}

	return v.messageMap[messageDefault](f)
}
